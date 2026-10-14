package service

import (
	"errors"
	"time"

	config "core/internal/config"
	gmail "core/pkg/gmail"
)

var ErrGmailNotConnected = errors.New("gmail is not connected")

type Gmail struct {
	client *gmail.ClientStruct
	cfg    *config.GmailConfig
}

type GmailStatus struct {
	Status     string   `json:"status"`
	Configured bool     `json:"configured"`
	Connected  bool     `json:"connected"`
	MissingEnv []string `json:"missing_env"`
	Message    string   `json:"message"`
}

func NewGmail(
	client *gmail.ClientStruct,
	cfg *config.GmailConfig,
) *Gmail {
	return &Gmail{client: client, cfg: cfg}
}

func (g *Gmail) Enabled() bool {
	return g.client != nil
}

func (g *Gmail) Status() GmailStatus {
	if !g.Enabled() {
		missing := []string{}

		if g.cfg.User == "" {
			missing = append(missing, "GMAIL_USER")
		}

		if g.cfg.Password == "" {
			missing = append(missing, "GMAIL_APP_PASSWORD")
		}

		return GmailStatus{
			Status:     "setup_required",
			MissingEnv: missing,
			Message:    "Set the missing Gmail vars in core/.env, then restart the API.",
		}
	}

	return GmailStatus{
		Status:     "ready",
		Configured: true,
		Connected:  true,
		Message:    "Gmail is connected via app password.",
	}
}

func (g *Gmail) Sync(
	since time.Time,
	label string,
) ([]gmail.Message, error) {
	if !g.Enabled() {
		return nil, ErrGmailNotConnected
	}

	return g.client.ListInbound(since, label)
}
