package config

import "os"

type Config struct {
	DBPath        string
	Port          string
	JWTSecret     string
	JWTAccessTTL  int // seconds
	JWTRefreshTTL int // seconds
}

func Default() Config {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret-do-not-use-in-production"
	}
	return Config{
		DBPath:        "finances.db",
		Port:          ":3000",
		JWTSecret:     secret,
		JWTAccessTTL:  3600,   // 1h
		JWTRefreshTTL: 604800, // 7d
	}
}
