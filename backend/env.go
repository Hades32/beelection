package main

import (
	"os"
	"time"

	"github.com/samber/lo"
)

var (
	EnvPort        = GetEnv("PORT", "8080")
	EnvIdleTimeout = lo.Must(time.ParseDuration(GetEnv("IDLE_TIMEOUT", "30s")))
)

func GetEnv(name, def string) string {
	v, ok := os.LookupEnv(name)
	if !ok {
		return def
	}
	return v
}
