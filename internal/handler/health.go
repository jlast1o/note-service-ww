package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type HealthHandler struct {
	pool *pgxpool.Pool
	rdb  *redis.Client
}

func NewHealthHandler(pool *pgxpool.Pool, rdb *redis.Client) *HealthHandler {
	return &HealthHandler{
		pool: pool,
		rdb:  rdb,
	}
}

type HealthResponse struct {
	Status   string `json:"status"`
	Postgres string `json:"postgres"`
	Redis    string `json:"redis"`
}

func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)

	defer cancel()

	resp := HealthResponse{Status: "ok"}

	if err := h.pool.Ping(ctx); err != nil {
		resp.Postgres = "unavaileble: " + err.Error()
		resp.Status = "degraded"
	} else {
		resp.Postgres = "ok"
	}

	if h.rdb != nil {
		if err := h.rdb.Ping(ctx).Err(); err != nil {
			resp.Redis = "unavaileble: " + err.Error()
			resp.Status = "degraded"
		} else {
			resp.Redis = "ok"
		}
	} else {
		resp.Redis = "not configured"
	}

	w.Header().Set("Content-Type", "application/json")

	if resp.Status == "ok" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	json.NewEncoder(w).Encode(resp)

}
