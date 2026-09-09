// Package repository 提供数据访问层，包括数据库连接、Ent 客户端和版本化迁移。
//
// migrate.go 实现了一个轻量级版本化 SQL 迁移运行器，使用嵌入的 SQL 文件
// (migrations/*.sql) 和 schema_migrations 跟踪表，无需外部依赖。

package repository

import (
	"context"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/imageforge/imageforge/ent"
)

// MigrateUp 执行所有待处理的版本化 SQL 迁移。
//
// 它使用 schema_migrations 表跟踪已应用的迁移，按文件名排序执行，
// 每个迁移在独立事务中运行。已应用的迁移不会重复执行。
//
// 该函数是幂等的：可安全地在每次启动时调用。
func MigrateUp(db *ent.Client, migrationFS fs.FS) error {
	ctx := context.Background()

	// 确保 schema_migrations 跟踪表存在。
	// 使用 Ent 的 ExecContext（由 sql/execquery 特性启用）。
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename    VARCHAR(255) PRIMARY KEY,
			applied_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			checksum    VARCHAR(64)
		)
	`); err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	// 读取所有 .sql 文件并按文件名排序。
	entries, err := fs.ReadDir(migrationFS, ".")
	if err != nil {
		return fmt.Errorf("read migration dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		files = append(files, e.Name())
	}
	sort.Strings(files)

	// 逐个执行未应用的迁移。
	for _, name := range files {
		applied, err := isMigrationApplied(ctx, db, name)
		if err != nil {
			return fmt.Errorf("check migration %s: %w", name, err)
		}
		if applied {
			continue
		}

		content, err := fs.ReadFile(migrationFS, name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}

		if err := applyMigration(ctx, db, name, string(content)); err != nil {
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
	}

	return nil
}

// isMigrationApplied 检查指定迁移文件是否已应用。
func isMigrationApplied(ctx context.Context, db *ent.Client, filename string) (bool, error) {
	rows, err := db.QueryContext(ctx, "SELECT COUNT(*) FROM schema_migrations WHERE filename = $1", filename)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	var count int
	if rows.Next() {
		if err := rows.Scan(&count); err != nil {
			return false, err
		}
	}
	return count > 0, nil
}

// applyMigration 在事务中执行迁移 SQL 并记录到 schema_migrations。
func applyMigration(ctx context.Context, db *ent.Client, filename, content string) error {
	tx, err := db.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	// 执行迁移 SQL。
	if _, err := tx.ExecContext(ctx, content); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("exec sql: %w", err)
	}

	// 记录迁移已应用。
	checksum := simpleChecksum(content)
	if _, err := tx.ExecContext(ctx,
		"INSERT INTO schema_migrations (filename, checksum) VALUES ($1, $2)",
		filename, checksum,
	); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("record migration: %w", err)
	}

	return tx.Commit()
}

// simpleChecksum 计算字符串的简单 FNV-1a 哈希，用于迁移完整性校验。
func simpleChecksum(s string) string {
	const (
		offset32 = 2166136261
		prime32  = 16777619
	)
	hash := uint32(offset32)
	for i := range s {
		hash ^= uint32(s[i])
		hash *= prime32
	}
	return fmt.Sprintf("%08x", hash)
}
