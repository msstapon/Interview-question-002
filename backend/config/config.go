// Package config loads typed configuration via Viper (env > defaults).
package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App AppConfig `mapstructure:"app"`
	DB  DBConfig  `mapstructure:"db"`
	JWT JWTConfig `mapstructure:"jwt"`
	Obs ObsConfig `mapstructure:"obs"`
}

type AppConfig struct {
	Env            string   `mapstructure:"env"`
	Port           int      `mapstructure:"port"`
	Name           string   `mapstructure:"name"`
	AllowedOrigins []string `mapstructure:"allowed_origins"`
}

type DBConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"sslmode"`
	MaxOpen  int    `mapstructure:"max_open"`
	MaxIdle  int    `mapstructure:"max_idle"`
}

// DSN is the connection string for the application database.
func (d DBConfig) DSN() string { return d.dsnFor(d.Name) }

// MaintenanceDSN targets the always-present "postgres" database — used to run
// CREATE DATABASE (which cannot run against the database being created).
func (d DBConfig) MaintenanceDSN() string { return d.dsnFor("postgres") }

func (d DBConfig) dsnFor(name string) string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, name, d.SSLMode)
}

type JWTConfig struct {
	PrivateKeyPath string        `mapstructure:"private_key_path"`
	PublicKeyPath  string        `mapstructure:"public_key_path"`
	AccessTTL      time.Duration `mapstructure:"access_ttl"`
	Issuer         string        `mapstructure:"issuer"`
}

type ObsConfig struct {
	LogLevel  string `mapstructure:"log_level"`
	LogFormat string `mapstructure:"log_format"`
}

// Load reads config from env (primary source), with defaults applied.
// A local .env file (if present in the working directory) is loaded first so the
// app runs standalone via `go run ./cmd/api` without needing `make` to export vars.
func Load() (*Config, error) {
	loadDotEnv(".env")

	v := viper.New()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	bindEnv(v)
	setDefaults(v)

	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return nil, fmt.Errorf("config unmarshal: %w", err)
	}
	if err := c.validate(); err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *Config) validate() error {
	if c.App.Port == 0 {
		return fmt.Errorf("app.port required")
	}
	if c.DB.Host == "" || c.DB.Name == "" {
		return fmt.Errorf("db.host and db.name required")
	}
	if c.JWT.PrivateKeyPath == "" || c.JWT.PublicKeyPath == "" {
		return fmt.Errorf("jwt key paths required (run `make keys`)")
	}
	return nil
}

func (c *Config) IsProd() bool { return c.App.Env == "production" }

func bindEnv(v *viper.Viper) {
	binds := map[string]string{
		"app.env":             "APP_ENV",
		"app.port":            "APP_PORT",
		"app.name":            "APP_NAME",
		"app.allowed_origins": "APP_ALLOWED_ORIGINS",

		"db.host":     "DB_HOST",
		"db.port":     "DB_PORT",
		"db.user":     "DB_USER",
		"db.password": "DB_PASSWORD",
		"db.name":     "DB_NAME",
		"db.sslmode":  "DB_SSLMODE",
		"db.max_open": "DB_MAX_OPEN",
		"db.max_idle": "DB_MAX_IDLE",

		"jwt.private_key_path": "JWT_PRIVATE_KEY_PATH",
		"jwt.public_key_path":  "JWT_PUBLIC_KEY_PATH",
		"jwt.access_ttl":       "JWT_ACCESS_TTL",
		"jwt.issuer":           "JWT_ISSUER",

		"obs.log_level":  "LOG_LEVEL",
		"obs.log_format": "LOG_FORMAT",
	}
	for k, env := range binds {
		_ = v.BindEnv(k, env)
	}
}

// loadDotEnv reads KEY=VALUE lines from path and sets them as environment variables
// (without overriding vars already present in the real environment). Missing file is
// a no-op. Comments (#...) and blank lines are ignored.
func loadDotEnv(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		i := strings.Index(line, "=")
		if i < 0 {
			continue
		}
		key := strings.TrimSpace(line[:i])
		val := strings.Trim(strings.TrimSpace(line[i+1:]), `"'`)
		if key == "" {
			continue
		}
		if _, ok := os.LookupEnv(key); !ok {
			_ = os.Setenv(key, val)
		}
	}
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.env", "development")
	v.SetDefault("app.port", 8080)
	v.SetDefault("app.name", "interview-question-002")
	v.SetDefault("app.allowed_origins", []string{"http://localhost:5173"})

	v.SetDefault("db.port", 5432)
	v.SetDefault("db.sslmode", "disable")
	v.SetDefault("db.max_open", 25)
	v.SetDefault("db.max_idle", 10)

	v.SetDefault("jwt.access_ttl", "60m")
	v.SetDefault("jwt.issuer", "example.com")

	v.SetDefault("obs.log_level", "info")
	v.SetDefault("obs.log_format", "console")
}
