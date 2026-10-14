package gmail

import (
	"context"
	"os"

	cfg "core/internal/config"
	logger "core/pkg/log"

	"golang.org/x/oauth2/google"
	sdk "google.golang.org/api/gmail/v1"
	option "google.golang.org/api/option"
)

type ClientStruct struct {
	*sdk.Service
}

var Client *ClientStruct

func Configure(
	config *cfg.GmailConfig,
) {
	if config.Credentials == "" || config.User == "" {
		return
	}

	key, err := os.ReadFile(config.Credentials)
	logger.Fatal(err, "Failed to read gmail service account key")

	jwt, err := google.JWTConfigFromJSON(key, sdk.GmailReadonlyScope)
	logger.Fatal(err, "Failed to parse gmail service account key")

	jwt.Subject = config.User
	service, err := sdk.NewService(
		context.Background(),
		option.WithHTTPClient(jwt.Client(context.Background())),
	)
	logger.Fatal(err, "Failed to start gmail client")
	Client = &ClientStruct{Service: service}
}
