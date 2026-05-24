package api

import (
	"strings"
	"time"

	"scamshield/internal/domain"
)

type WhatsAppWebhookPayload struct {
	Entry []struct {
		Changes []struct {
			Value struct {
				Messages []WhatsAppMessage `json:"messages"`
			} `json:"value"`
		} `json:"changes"`
	} `json:"entry"`
}

type WhatsAppMessage struct {
	ID        string `json:"id"`
	From      string `json:"from"`
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Text      struct {
		Body string `json:"body"`
	} `json:"text"`
	Image struct {
		ID      string `json:"id"`
		Caption string `json:"caption"`
	} `json:"image"`
	Document struct {
		ID      string `json:"id"`
		Caption string `json:"caption"`
	} `json:"document"`
	Button struct {
		Text string `json:"text"`
	} `json:"button"`
	Interactive struct {
		ButtonReply struct {
			Title string `json:"title"`
		} `json:"button_reply"`
	} `json:"interactive"`
}

func NormalizeWhatsAppPayload(payload WhatsAppWebhookPayload) []domain.WhatsAppInbound {
	var inbound []domain.WhatsAppInbound
	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			for _, message := range change.Value.Messages {
				item := domain.WhatsAppInbound{
					MessageID: message.ID,
					From:      message.From,
					Type:      message.Type,
					CreatedAt: time.Now().UTC(),
				}
				switch message.Type {
				case "text":
					item.Body = message.Text.Body
				case "image":
					item.MediaID = message.Image.ID
					item.Caption = message.Image.Caption
					item.Body = message.Image.Caption
				case "document":
					item.MediaID = message.Document.ID
					item.Caption = message.Document.Caption
					item.Body = message.Document.Caption
				case "button":
					item.Body = message.Button.Text
				case "interactive":
					item.Body = message.Interactive.ButtonReply.Title
				default:
					item.Body = strings.TrimSpace(message.Text.Body + " " + message.Image.Caption + " " + message.Document.Caption)
				}
				if item.From != "" {
					inbound = append(inbound, item)
				}
			}
		}
	}
	return inbound
}
