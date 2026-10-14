package gmail

import (
	"context"

	cfg "core/internal/config"
	logger "core/pkg/log"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	sdk "google.golang.org/api/gmail/v1"
	option "google.golang.org/api/option"
)

type ClientStruct struct {
	*sdk.Service
	Label string
}

var Client *ClientStruct

func Configure(
	config *cfg.GmailConfig,
) {
	if config.ClientID == "" || config.ClientSecret == "" || config.RefreshToken == "" {
		return
	}

	oauth := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Scopes:       []string{sdk.GmailReadonlyScope},
		Endpoint:     google.Endpoint,
	}

	service, err := sdk.NewService(
		context.Background(),
		option.WithHTTPClient(
			oauth.Client(
				context.Background(),
				&oauth2.Token{RefreshToken: config.RefreshToken},
			),
		),
	)
	logger.Fatal(err, "Failed to start gmail client")

	Client = &ClientStruct{
		Service: service,
		Label:   config.Label,
	}
}
