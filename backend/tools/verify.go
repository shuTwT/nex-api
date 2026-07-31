package tools

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
)

var coreTables = []string{
	"User", "Account", "Session", "VerificationToken", "ApiCategory", "Api",
	"ApiParameter", "ApiResponse", "ApiUsage", "SubscriptionPlan", "Subscription",
	"ApiToken", "AuditLog", "SystemSetting", "Payment", "RedemptionCode",
	"Advertisement", "McpService", "McpUsage",
}

// ForeignKeyViolation is one row returned by PRAGMA foreign_key_check.
type ForeignKeyViolation struct {
	Table      string `json:"table"`
	RowID      int64  `json:"rowId"`
	Parent     string `json:"parent"`
	ForeignKey int64  `json:"foreignKey"`
}

// CriticalRecordViolation identifies a failed domain invariant.
type CriticalRecordViolation struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

// IntegrityReport contains deterministic schema and data checks for a backup.
type IntegrityReport struct {
	SchemaChecksum         string                    `json:"schemaChecksum"`
	RowCounts              map[string]int64          `json:"rowCounts"`
	ForeignKeysOK          bool                      `json:"foreignKeysOk"`
	ForeignKeyViolations   []ForeignKeyViolation     `json:"foreignKeyViolations"`
	DomainChecksums        map[string]string         `json:"domainChecksums"`
	CriticalRecords        map[string]int64          `json:"criticalRecords"`
	CriticalRecordFailures []CriticalRecordViolation `json:"criticalRecordFailures"`
}

// VerifyDatabase checks the legacy tables, foreign keys, critical records, and domain checksums.
func VerifyDatabase(ctx context.Context, path string) (report IntegrityReport, err error) {
	db, err := openSQLiteReadOnly(path)
	if err != nil {
		return IntegrityReport{}, fmt.Errorf("open database for verification: %w", err)
	}
	defer func() { err = errors.Join(err, db.Close()) }()
	report, err = verifyDatabase(ctx, db)
	if err != nil {
		return report, err
	}
	return report, nil
}

func verifyDatabase(ctx context.Context, db *sql.DB) (IntegrityReport, error) {
	schemaChecksum, err := schemaChecksum(ctx, db)
	if err != nil {
		return IntegrityReport{}, fmt.Errorf("schema checksum: %w", err)
	}
	rowCounts, err := tableRowCounts(ctx, db)
	if err != nil {
		return IntegrityReport{}, fmt.Errorf("row counts: %w", err)
	}
	foreignKeysOK, foreignKeyViolations, err := foreignKeyCheck(ctx, db)
	if err != nil {
		return IntegrityReport{}, fmt.Errorf("foreign key check: %w", err)
	}
	domainChecksums, err := domainChecksums(ctx, db)
	if err != nil {
		return IntegrityReport{}, fmt.Errorf("domain checksums: %w", err)
	}
	criticalRecords, criticalFailures, err := criticalRecords(ctx, db, rowCounts)
	if err != nil {
		return IntegrityReport{}, fmt.Errorf("critical records: %w", err)
	}
	report := IntegrityReport{
		SchemaChecksum:         schemaChecksum,
		RowCounts:              rowCounts,
		ForeignKeysOK:          foreignKeysOK,
		ForeignKeyViolations:   foreignKeyViolations,
		DomainChecksums:        domainChecksums,
		CriticalRecords:        criticalRecords,
		CriticalRecordFailures: criticalFailures,
	}
	if !foreignKeysOK {
		return report, errors.New("foreign key integrity check failed")
	}
	if len(criticalFailures) > 0 {
		return report, fmt.Errorf("critical record verification failed: %d invariant(s)", len(criticalFailures))
	}
	return report, nil
}

func schemaChecksum(ctx context.Context, db *sql.DB) (string, error) {
	const query = `
		SELECT type, name, tbl_name, COALESCE(sql, '')
		FROM sqlite_master
		WHERE name NOT LIKE 'sqlite_%'
		ORDER BY type, name`
	rows, err := db.QueryContext(ctx, `SELECT name FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%'`)
	if err != nil {
		return "", err
	}
	tables := make(map[string]struct{}, len(coreTables))
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			rows.Close()
			return "", err
		}
		tables[name] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return "", err
	}
	if err := rows.Close(); err != nil {
		return "", err
	}
	for _, table := range coreTables {
		if _, ok := tables[table]; !ok {
			return "", fmt.Errorf("required table %q is missing", table)
		}
	}
	return checksumQuery(ctx, db, "schema", query)
}

func tableRowCounts(ctx context.Context, db *sql.DB) (map[string]int64, error) {
	counts := make(map[string]int64, len(coreTables))
	for _, table := range coreTables {
		query := "SELECT COUNT(*) FROM " + quoteIdentifier(table)
		var count int64
		if err := db.QueryRowContext(ctx, query).Scan(&count); err != nil {
			return nil, fmt.Errorf("%s: %w", table, err)
		}
		counts[table] = count
	}
	return counts, nil
}

func foreignKeyCheck(ctx context.Context, db *sql.DB) (bool, []ForeignKeyViolation, error) {
	var enabled int64
	if err := db.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&enabled); err != nil {
		return false, nil, err
	}
	rows, err := db.QueryContext(ctx, "PRAGMA foreign_key_check")
	if err != nil {
		return false, nil, err
	}
	defer rows.Close()
	violations := make([]ForeignKeyViolation, 0)
	for rows.Next() {
		var violation ForeignKeyViolation
		if err := rows.Scan(&violation.Table, &violation.RowID, &violation.Parent, &violation.ForeignKey); err != nil {
			return false, nil, err
		}
		violations = append(violations, violation)
	}
	if err := rows.Err(); err != nil {
		return false, nil, err
	}
	return enabled == 1 && len(violations) == 0, violations, nil
}

func domainChecksums(ctx context.Context, db *sql.DB) (map[string]string, error) {
	queries := map[string][]string{
		"users": {
			`SELECT id, name, email, emailVerified, image, username, password, role, credits, createdAt, updatedAt FROM "User" ORDER BY id`,
		},
		"credits": {
			`SELECT id, credits FROM "User" ORDER BY id`,
			`SELECT id, userId, apiId, credits, status, createdAt FROM "ApiUsage" ORDER BY id`,
			`SELECT id, userId, mcpId, credits, status, createdAt FROM "McpUsage" ORDER BY id`,
		},
		"tokens": {
			`SELECT id, userId, name, token, permissions, lastUsedAt, expiresAt, isActive, createdAt, updatedAt FROM "ApiToken" ORDER BY id`,
		},
		"payments": {
			`SELECT id, userId, outTradeNo, transactionId, method, amount, currency, status, qrcodeUrl, payUrl, notifyUrl, paidAt, expiredAt, cancelledAt, metadata, createdAt, updatedAt FROM "Payment" ORDER BY id`,
		},
		"subscriptions": {
			`SELECT id, userId, planId, planName, credits, price, startDate, endDate, isActive, createdAt, updatedAt, paymentId FROM "Subscription" ORDER BY id`,
		},
	}
	names := make([]string, 0, len(queries))
	for name := range queries {
		names = append(names, name)
	}
	sort.Strings(names)
	checksums := make(map[string]string, len(queries))
	for _, name := range names {
		checksum, err := checksumQueries(ctx, db, name, queries[name])
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		checksums[name] = checksum
	}
	return checksums, nil
}

func criticalRecords(ctx context.Context, db *sql.DB, rowCounts map[string]int64) (map[string]int64, []CriticalRecordViolation, error) {
	records := map[string]int64{
		"users":         rowCounts["User"],
		"tokens":        rowCounts["ApiToken"],
		"payments":      rowCounts["Payment"],
		"subscriptions": rowCounts["Subscription"],
	}
	queries := map[string]string{
		"negative_user_credits": `SELECT COUNT(*) FROM "User" WHERE credits < 0`,
		"blank_user_identity":   `SELECT COUNT(*) FROM "User" WHERE email = '' OR username = ''`,
		"blank_api_tokens":      `SELECT COUNT(*) FROM "ApiToken" WHERE token = ''`,
		"blank_payment_orders":  `SELECT COUNT(*) FROM "Payment" WHERE outTradeNo = ''`,
		"blank_subscriptions":   `SELECT COUNT(*) FROM "Subscription" WHERE userId = ''`,
	}
	names := make([]string, 0, len(queries))
	for name := range queries {
		names = append(names, name)
	}
	sort.Strings(names)
	failures := make([]CriticalRecordViolation, 0)
	for _, name := range names {
		var count int64
		if err := db.QueryRowContext(ctx, queries[name]).Scan(&count); err != nil {
			return nil, nil, fmt.Errorf("%s: %w", name, err)
		}
		if count > 0 {
			failures = append(failures, CriticalRecordViolation{Name: name, Count: count})
		}
	}
	return records, failures, nil
}
