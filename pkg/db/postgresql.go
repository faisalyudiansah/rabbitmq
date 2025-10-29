package db

import (
	"background-job-service/config"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewPostgreSQL(cfg *config.Config) *sql.DB {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		cfg.DBHost,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBPort,
		cfg.DBSSLMode,
		cfg.Zone,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("error initializing database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}

	db.SetMaxIdleConns(cfg.DBMaxIdleConn)
	db.SetMaxOpenConns(cfg.DBMaxOpenConn)
	db.SetConnMaxLifetime(time.Duration(cfg.DBMaxConnLifetimeMinute) * time.Minute)

	return db
}
