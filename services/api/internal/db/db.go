package db

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PoolConfig configures the Postgres connection pool.
type PoolConfig struct {
	DatabaseURL           string
	PingConnBeforeAcquire bool
	MaxConns              int32
	MinConns              int32
	StatementTimeoutMs    int
}

// Connect opens the pool using PoolConfig.
// When pingConnBeforeAcquire is true, each connection is pinged before use
// (extra round-trip per checkout). Production should leave this false and rely on HealthCheckPeriod.
func Connect(ctx context.Context, pc PoolConfig) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(pc.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse pool config: %w", err)
	}
	maxC := pc.MaxConns
	if maxC <= 0 {
		maxC = 20
	}
	minC := pc.MinConns
	if minC < 0 {
		minC = 0
	}
	if minC > maxC {
		minC = maxC
	}
	stmtMs := pc.StatementTimeoutMs
	if stmtMs <= 0 {
		stmtMs = 5000
	}
	cfg.MaxConns = maxC
	cfg.MinConns = minC
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.HealthCheckPeriod = 1 * time.Minute
	if cfg.ConnConfig.RuntimeParams == nil {
		cfg.ConnConfig.RuntimeParams = map[string]string{}
	}
	cfg.ConnConfig.RuntimeParams["statement_timeout"] = strconv.Itoa(stmtMs)
	if pc.PingConnBeforeAcquire {
		cfg.BeforeAcquire = func(ctx context.Context, c *pgx.Conn) bool {
			pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()
			return c.Ping(pingCtx) == nil
		}
	} else {
		cfg.BeforeAcquire = nil
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return pool, nil
}
