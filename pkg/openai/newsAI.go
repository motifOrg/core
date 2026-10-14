package openai

import (
	"encoding/json"
	"fmt"

	"core/internal/model"
)

const newsPrompt = `
You are a research analyst for a private equity firm.
Search the web for recent news about the company below.

Company:
%s

Investment criteria:
%s

Rules:
- Use only news from the last 24 months
- Every item must have a source URL from your search
- source_url is a plain URL that starts with https://, not a markdown link
- summary and content are plain text with no links or citations
- kind is "fit" when the news matches the criteria
- kind is "risk" when the news is a lawsuit, layoff, debt, sale, or closure
- kind is "finding" for any other news
- confidence is "high", "medium", or "low"
- Return an empty items array when no news is found
- Do not invent facts

Return ONLY valid JSON in this format:
{
  "summary": "string",
  "items": [
    {
      "kind": "fit | risk | finding",
      "title": "string",
      "content": "string",
      "source_url": "string",
      "published_at": "YYYY-MM-DD",
      "confidence": "high | medium | low"
    }
  ]
}
`

var newsSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"summary": map[string]any{"type": "string"},
		"items": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"kind":         map[string]any{"type": "string", "enum": []string{"fit", "risk", "finding"}},
					"title":        map[string]any{"type": "string"},
					"content":      map[string]any{"type": "string"},
					"source_url":   map[string]any{"type": "string"},
					"published_at": map[string]any{"type": "string"},
					"confidence":   map[string]any{"type": "string", "enum": []string{"high", "medium", "low"}},
				},
				"required":             []string{"kind", "title", "content", "source_url", "published_at", "confidence"},
				"additionalProperties": false,
			},
		},
	},
	"required":             []string{"summary", "items"},
	"additionalProperties": false,
}

type NewsItem struct {
	Kind        string `json:"kind"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	SourceURL   string `json:"source_url"`
	PublishedAt string `json:"published_at"`
	Confidence  string `json:"confidence"`
}

type NewsResponse struct {
	Summary string     `json:"summary"`
	Items   []NewsItem `json:"items"`
}

func (c *ClientStruct) NewsResearch(
	company model.Company,
	icp *model.ICPCriteria,
) (*NewsResponse, error) {
	companyJSON, err := json.MarshalIndent(
		company,
		"",
		"  ",
	)

	if err != nil {
		return nil, err
	}

	icpJSON, err := json.MarshalIndent(
		icp,
		"",
		"  ",
	)

	if err != nil {
		return nil, err
	}

	input := fmt.Sprintf(
		newsPrompt,
		companyJSON,
		icpJSON,
	)

	output, err := c.search(input, format("news", newsSchema))

	if err != nil {
		return nil, err
	}

	var result NewsResponse

	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return nil, fmt.Errorf(
			"failed to parse news response: %w\nresponse=%s",
			err,
			output,
		)
	}

	return &result, nil
}
