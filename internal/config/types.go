package config

type ServerConfig struct {
	Port      string
	AppOrigin string
}

type StorageConfig struct {
	UploadDir   string
	MaxUploadMB int
}

type GmailConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	TokenKey     string
}

type DatabaseConfig struct {
	SSL      string
	Name     string
	Host     string
	Port     string
	Password string
	Username string
}

type Config struct {
	Server   ServerConfig
	Storage  StorageConfig
	Gmail    GmailConfig
	Database DatabaseConfig
}
