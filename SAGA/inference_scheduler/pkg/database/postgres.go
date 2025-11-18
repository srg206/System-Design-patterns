package database

import (
	"context"
	"fmt"

	"inference_scheduler/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, dbConfig config.DatabaseConfig, poolConfig config.PoolConfig) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Name,
	)

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	poolCfg.MaxConns = poolConfig.MaxConns
	poolCfg.MinConns = poolConfig.MinConns
	poolCfg.MaxConnLifetime = poolConfig.MaxConnLifetime
	poolCfg.MaxConnIdleTime = poolConfig.MaxConnIdleTime
	poolCfg.HealthCheckPeriod = poolConfig.HealthCheckPeriod

	if poolConfig.ConnectTimeout > 0 {
		connectCtx, cancel := context.WithTimeout(ctx, poolConfig.ConnectTimeout)
		defer cancel()
		ctx = connectCtx
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

func Close(pool *pgxpool.Pool) {
	if pool != nil {
		pool.Close()
	}
}
