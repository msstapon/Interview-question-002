package migrations

import (
	"gorm.io/gorm"

	"example.com/interview-question-002/internal/infrastructure/db/migrate"
)

func init() {
	migrate.Register(migrate.Migration{
		Version: 1,
		Name:    "create_users",
		Up: func(tx *gorm.DB) error {
			stmts := []string{
				`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`,

				`CREATE TABLE IF NOT EXISTS users (
					id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
					username      VARCHAR(100) NOT NULL,
					password_hash TEXT NOT NULL,
					created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					CONSTRAINT uq_users_username UNIQUE (username)
				)`,
				`CREATE INDEX IF NOT EXISTS idx_users_username ON users (username)`,
			}
			for _, s := range stmts {
				if err := tx.Exec(s).Error; err != nil {
					return err
				}
			}
			return nil
		},
		Down: func(tx *gorm.DB) error {
			return tx.Exec(`DROP TABLE IF EXISTS users`).Error
		},
	})
}
