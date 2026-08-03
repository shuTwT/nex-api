package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"

	"github.com/shuTwT/nex-api/backend/ent"
	"github.com/shuTwT/nex-api/backend/internal/infra/config"
)

// Open 打开数据库并返回 Ent client。
//
// 当前仅链接 SQLite(modernc.org/sqlite,纯 Go 驱动,无 cgo 依赖)。
// 全新的 SQLite 数据库会在启动时自动创建 schema;已存在的数据库由
// atlas 迁移(migrations/)管理,启动时不做任何改动。
func Open(ctx context.Context, cfg config.Database, logger *slog.Logger) (*ent.Client, error) {
	scheme, dsn := splitDatabaseDSN(cfg.URL)
	switch scheme {
	case "sqlite":
		return openSQLite(ctx, cfg, dsn, logger)
	case "mysql", "postgres", "postgresql":
		return nil, fmt.Errorf("database scheme %q is not linked: add the driver blank import to cmd/server/main.go", scheme)
	default:
		return nil, fmt.Errorf("unsupported database url %q", cfg.URL)
	}
}

// splitDatabaseDSN 从配置的 URL 中分离出数据库类型与 DSN。
func splitDatabaseDSN(raw string) (scheme, dsn string) {
	lower := strings.ToLower(strings.TrimSpace(raw))
	for _, s := range []string{"mysql", "postgres", "postgresql"} {
		if strings.HasPrefix(lower, s+"://") {
			return s, raw[len(s)+3:]
		}
	}
	if strings.HasPrefix(lower, "sqlite://") {
		return "sqlite", raw[len("sqlite://"):]
	}
	// file: 前缀或裸路径(如 data/dev.db)按 SQLite 处理。
	return "sqlite", raw
}

func openSQLite(ctx context.Context, cfg config.Database, dsn string, logger *slog.Logger) (*ent.Client, error) {
	// 必须在打开连接之前判断:SQLite 首次打开/Ping 会创建文件。
	newDatabase := isNewSQLiteFile(dsn)
	if err := ensureSQLiteParentDirectory(dsn); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", withSQLiteFK(dsn))
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	// SQLite 是单写者数据库,限制连接数避免写锁竞争。
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.SQLite, db)))

	if newDatabase {
		if err := client.Schema.Create(ctx); err != nil {
			_ = client.Close()
			return nil, fmt.Errorf("create sqlite schema: %w", err)
		}
		logger.Info("created database schema for new sqlite database")
	} else {
		logger.Info("database already exists; schema is managed by atlas migrations")
	}
	return client, nil
}

// ensureSQLiteParentDirectory 创建 SQLite 文件数据库的父目录。
// SQLite 驱动会创建不存在的数据库文件，但不会创建其父目录。
func ensureSQLiteParentDirectory(dsn string) error {
	path := sqliteFilePath(dsn)
	if path == "" {
		return nil
	}
	parent := filepath.Dir(path)
	if parent == "." {
		return nil
	}
	if err := os.MkdirAll(parent, 0o750); err != nil {
		return fmt.Errorf("create sqlite database directory %q: %w", parent, err)
	}
	return nil
}

// withSQLiteFK 确保 DSN 启用了外键 pragma(ent 的 schema 创建/查询依赖)。
func withSQLiteFK(dsn string) string {
	if strings.Contains(dsn, "_fk=") || strings.Contains(dsn, "_pragma=") {
		return dsn
	}
	sep := "?"
	if strings.Contains(dsn, "?") {
		sep = "&"
	}
	return dsn + sep + "_fk=1"
}

// isNewSQLiteFile 判断 DSN 指向的 SQLite 文件是否尚不存在(全新库)。
// 内存库始终返回 false。
func isNewSQLiteFile(dsn string) bool {
	path := sqliteFilePath(dsn)
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return errors.Is(err, os.ErrNotExist)
}

// sqliteFilePath 从 SQLite DSN 中解析出数据库文件路径。
func sqliteFilePath(dsn string) string {
	rest := strings.TrimSpace(dsn)
	for _, prefix := range []string{"file:", "sqlite:"} {
		if strings.HasPrefix(rest, prefix) {
			rest = strings.TrimPrefix(rest, prefix)
			break
		}
	}
	if i := strings.IndexByte(rest, '?'); i >= 0 {
		rest = rest[:i]
	}
	if rest == "" || rest == ":memory:" || strings.HasPrefix(rest, "mode=memory") {
		return ""
	}
	if strings.HasPrefix(rest, "//") {
		// file:///abs/path 形式
		return strings.TrimPrefix(rest, "//")
	}
	return rest
}
