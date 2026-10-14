package gmail

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	imap "github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	_ "github.com/emersion/go-message/charset"
	"github.com/emersion/go-message/mail"
)

const (
	address     = "imap.gmail.com:993"
	maxMessages = 100
	maxBody     = 4000
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
	since time.Time,
	label string,
) ([]Message, error) {
	if label == "" {
		label = "INBOX"
	}

	client, err := imapclient.DialTLS(address, nil)

	if err != nil {
		return nil, err
	}

	defer client.Close()

	if err := client.Login(c.User, c.Password).Wait();
		err != nil {
		return nil, err
	}

	defer func() { _ = client.Logout().Wait() }()

	if _, err := client.Select(
		label, 
		&imap.SelectOptions{ReadOnly: true},
	).Wait();
		err != nil {
		return nil, fmt.Errorf("open label %q: %w", label, err)
	}

	search, err := client.UIDSearch(
		&imap.SearchCriteria{Since: since}, 
		nil,
	).Wait()

	if err != nil {
		return nil, err
	}

	uids := search.AllUIDs()
	messages := []Message{}

	if len(uids) == 0 {
		return messages, nil
	}

	if len(uids) > maxMessages {
		uids = uids[len(uids)-maxMessages:]
	}

	section := &imap.FetchItemBodySection{Peek: true}

	fetched, err := client.Fetch(
		imap.UIDSetNum(uids...),
		&imap.FetchOptions{
			UID:          true,
			Envelope:     true,
			InternalDate: true,
			BodySection:  []*imap.FetchItemBodySection{section},
		},
	).Collect()

	if err != nil {
		return nil, err
	}

	for _, msg := range fetched {
		if msg.Envelope == nil || len(msg.Envelope.From) == 0 {
			continue
		}

		sender := msg.Envelope.From[0]

		ref := msg.Envelope.MessageID
		if ref == "" {
			ref = fmt.Sprintf("%s:%d", label, msg.UID)
		}

		var body string
		reader, err := mail.CreateReader(
			bytes.NewReader(msg.FindBodySection(section)),
		)

		for err == nil {
			var part *mail.Part

			if  part, err = reader.NextPart(); 
				err != nil {
				break
			}

			header, ok := part.Header.(*mail.InlineHeader)
			if !ok {
				continue
			}

			if kind, _, _ := header.ContentType(); kind != "text/plain" {
				continue
			}

			data, _ := io.ReadAll(io.LimitReader(part.Body, maxBody))
			body = strings.TrimSpace(strings.ToValidUTF8(string(data), ""))
			break
		}

		messages = append(messages, Message{
			ExternalRef: ref,
			SenderName:  strings.TrimSpace(sender.Name),
			SenderEmail: strings.ToLower(sender.Addr()),
			Subject:     strings.TrimSpace(msg.Envelope.Subject),
			Body:        body,
			ReceivedAt:  msg.InternalDate,
		})
	}

	return messages, nil
}
