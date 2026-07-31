package tools

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"strconv"
	"strings"
	"time"
)

func checksumQueries(ctx context.Context, db *sql.DB, label string, queries []string) (string, error) {
	hasher := sha256.New()
	for index, query := range queries {
		if err := hashPart(hasher, label+":"+strconv.Itoa(index)); err != nil {
			return "", err
		}
		if err := hashRows(ctx, db, hasher, query); err != nil {
			return "", err
		}
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func checksumQuery(ctx context.Context, db *sql.DB, label, query string) (string, error) {
	hasher := sha256.New()
	if err := hashPart(hasher, label); err != nil {
		return "", err
	}
	if err := hashRows(ctx, db, hasher, query); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func hashRows(ctx context.Context, db *sql.DB, hasher hash.Hash, query string) (err error) {
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, rows.Close()) }()
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	for _, column := range columns {
		if err := hashPart(hasher, column); err != nil {
			return err
		}
	}
	for rows.Next() {
		values := make([]any, len(columns))
		pointers := make([]any, len(columns))
		for index := range values {
			pointers[index] = &values[index]
		}
		if err := rows.Scan(pointers...); err != nil {
			return err
		}
		for _, value := range values {
			if err := hashPart(hasher, canonicalValue(value)); err != nil {
				return err
			}
		}
	}
	return rows.Err()
}

func hashPart(hasher hash.Hash, value string) error {
	if _, err := hasher.Write([]byte(strconv.Itoa(len(value)) + ":")); err != nil {
		return err
	}
	if _, err := hasher.Write([]byte(value)); err != nil {
		return err
	}
	_, err := hasher.Write([]byte{0})
	return err
}

func canonicalValue(value any) string {
	switch typed := value.(type) {
	case nil:
		return "null"
	case []byte:
		return "bytes:" + hex.EncodeToString(typed)
	case string:
		return "string:" + typed
	case int64:
		return "int:" + strconv.FormatInt(typed, 10)
	case float64:
		return "float:" + strconv.FormatFloat(typed, 'g', -1, 64)
	case bool:
		return "bool:" + strconv.FormatBool(typed)
	case time.Time:
		return "time:" + typed.UTC().Format(time.RFC3339Nano)
	default:
		return fmt.Sprintf("%T:%v", value, value)
	}
}

func quoteIdentifier(identifier string) string {
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}
