package gmail

import (
	cfg "core/internal/config"
)

type ClientStruct struct {
	User     string
	Password string
}

var Client *ClientStruct

func Configure(
	config *cfg.GmailConfig,
) {
	if config.User == "" || config.Password == "" {
		return
	}

	Client = &ClientStruct{
		User:     config.User,
		Password: config.Password,
	}
}
