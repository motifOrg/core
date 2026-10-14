package config

type ServerConfig struct {
	Port      string
	AppOrigin string
}

type GmailConfig struct {
	User     string
	Password string
}

type DatabaseConfig struct {
	SSL      string
	Name     string
	Host     string
	Port     string
	Password string
	Username string
}

type OpenAIConfig struct {
	SecretKey string
}

type Config struct {
	Server   ServerConfig
	Gmail    GmailConfig
	Database DatabaseConfig
	OpenAI   OpenAIConfig
}
