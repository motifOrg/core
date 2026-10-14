package gmail

import (
	"context"
	"fmt"
	"net/mail"
	"strings"
	"time"
)

type Message struct {
	ExternalRef string
	SenderName  string
	SenderEmail string
	Subject     string
	Body        string
	ReceivedAt  time.Time
}

func (c *ClientStruct) ListInbound(
	ctx context.Context,
	since time.Time,
	label string,
) ([]Message, error) {
	if label == "" {
		label = "INBOX"
	}

	query := fmt.Sprintf(
		"label:%s after:%s",
		label,
		since.Format("2006/01/02"),
	)

	list, err := c.Users.Messages.List("me").
		Context(ctx).
		Q(query).
		MaxResults(100).
		Do()

	if err != nil {
		return nil, err
	}

	messages := []Message{}

	for _, item := range list.Messages {
		msg, err := c.Users.Messages.Get("me", item.Id).
			Context(ctx).
			Format("metadata").
			MetadataHeaders("From", "Subject").
			Do()

		if err != nil {
			return nil, err
		}

		if msg.Payload == nil {
			continue
		}

		var from, subject string

		for _, item := range msg.Payload.Headers {
			switch strings.ToLower(item.Name) {
			case "from":
				from = item.Value
			case "subject":
				subject = strings.TrimSpace(item.Value)
			}
		}

		sender, err := mail.ParseAddress(from)

		if err != nil {
			continue
		}

		messages = append(messages, Message{
			ExternalRef: msg.Id,
			SenderName:  strings.TrimSpace(sender.Name),
			SenderEmail: strings.TrimSpace(sender.Address),
			Subject:     subject,
			Body:        msg.Snippet,
			ReceivedAt:  time.UnixMilli(msg.InternalDate),
		})
	}

	return messages, nil
}
