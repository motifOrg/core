package config

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/joho/godotenv"
)

func Load() (*Config, error) {
	_ = godotenv.Load(".env")
	return &Config{
		Server: ServerConfig{
			Port:      ":" + GetEnv("PORT", "8000"),
			AppOrigin: GetEnv("APP_ORIGIN", "http://localhost:3000"),
		},
		Gmail: GmailConfig{
			ClientID:     GetEnv("GMAIL_CLIENT_ID", ""),
			ClientSecret: GetEnv("GMAIL_CLIENT_SECRET", ""),
			RedirectURL:  GetEnv("GMAIL_REDIRECT_URL", "http://localhost"),
			RefreshToken: GetEnv("GMAIL_REFRESH_TOKEN", ""),
		},
		Database: DatabaseConfig{
			Name:     GetEnv("DB_NAME", "motif"),
			Username: GetEnv("DB_USER", "visshon"),
			Password: GetEnv("DB_PASSWORD", "password"),
			Host:     GetEnv("DB_HOST", "localhost"),
			Port:     GetEnv("DB_PORT", "5432"),
			SSL:      GetEnv("DB_SSLMODE", "disable"),
		},
		OpenAI: OpenAIConfig{
			SecretKey: GetEnv("OPENAI_API_KEY", ""),
		},
	}, nil
}

func (cfg *DatabaseConfig) dsn(name string) string {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.Username,
		name,
		cfg.SSL,
	)
	if cfg.Password != "" {
		dsn += " password=" + cfg.Password
	}
	return dsn
}

func (cfg *DatabaseConfig) GetDatabaseDSN() string {
	return cfg.dsn(cfg.Name)
}

func (cfg *DatabaseConfig) GetAdminDSN() string {
	return cfg.dsn("postgres")
}

func (cfg *DatabaseConfig) GetMigrationDSN() string {
	query := url.Values{}
	query.Set("sslmode", cfg.SSL)

	host := cfg.Host + ":" + cfg.Port
	if strings.HasPrefix(cfg.Host, "/") {
		query.Set("host", cfg.Host)
		query.Set("port", cfg.Port)
		host = ""
	}

	return (&url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.Username, cfg.Password),
		Host:     host,
		Path:     "/" + cfg.Name,
		RawQuery: query.Encode(),
	}).String()
}
