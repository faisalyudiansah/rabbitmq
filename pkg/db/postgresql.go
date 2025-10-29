package db

import (
	"background-job-service/config"
	dbLogger "background-job-service/pkg/logger/db"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jackc/pgx/v5/tracelog"
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

	pgxConfig, err := pgx.ParseConfig(dsn)
	if err != nil {
		log.Fatalf("error parsing config: %v", err)
	}

	pgxConfig.Tracer = &tracelog.TraceLog{
		Logger:   &dbLogger.PgxLogger{},
		LogLevel: tracelog.LogLevelTrace,
	}

	db := stdlib.OpenDB(*pgxConfig)

	db.SetMaxIdleConns(cfg.DBMaxIdleConn)
	db.SetMaxOpenConns(cfg.DBMaxOpenConn)
	db.SetConnMaxLifetime(time.Duration(cfg.DBMaxConnLifetimeMinute) * time.Minute)

	return db
}
