package database

import (
	"fmt"
	"time"

	"whatsapp-api/pkg/config"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func NewPostgresConnection(cfg *config.Config) (*sqlx.DB, error) {
	var dsn string
	if cfg.Database.URL != "" {
		dsn = cfg.Database.URL
	} else {
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.User,
			cfg.Database.Password,
			cfg.Database.Name,
			cfg.Database.SSLMode,
		)
	}

	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.Database.MaxConnections)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConnections)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
