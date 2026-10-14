package openai

import (
	"encoding/json"
	"fmt"

	"core/internal/model"
)

const revenuePrompt = `
You are a research analyst for a private equity firm.
Search the web for the revenue and headcount of the company below.

Company:
%s

Rules:
- revenue_estimate is a USD range, e.g. "$10M-$25M"
- revenue_estimate is "" when no source gives revenue
- employees is an integer or null
- Every source must be a URL from your search
- sources are plain URLs that start with https://, not markdown links
- summary is plain text with no links or citations
- confidence is "high" for filings or company statements
- confidence is "medium" for databases such as ZoomInfo or Crunchbase
- confidence is "low" for estimates from headcount
- Do not invent facts

Return ONLY valid JSON in this format:
{
  "revenue_estimate": "string",
  "employees": 0,
  "confidence": "high | medium | low",
  "summary": "string",
  "sources": ["string"]
}
`

var revenueSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"revenue_estimate": map[string]any{"type": "string"},
		"employees":        map[string]any{"type": []string{"integer", "null"}},
		"confidence":       map[string]any{"type": "string", "enum": []string{"high", "medium", "low"}},
		"summary":          map[string]any{"type": "string"},
		"sources":          map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
	},
	"required":             []string{"revenue_estimate", "employees", "confidence", "summary", "sources"},
	"additionalProperties": false,
}

type RevenueResponse struct {
	RevenueEstimate string   `json:"revenue_estimate"`
	Employees       *int     `json:"employees"`
	Confidence      string   `json:"confidence"`
	Summary         string   `json:"summary"`
	Sources         []string `json:"sources"`
}

func (c *ClientStruct) RevenueResearch(
	company model.Company,
) (*RevenueResponse, error) {
	companyJSON, err := json.MarshalIndent(
		company,
		"",
		"  ",
	)

	if err != nil {
		return nil, err
	}

	input := fmt.Sprintf(
		revenuePrompt,
		companyJSON,
	)

	output, err := c.search(
		input, 
		format("revenue", revenueSchema),
	)

	if err != nil {
		return nil, err
	}

	var result RevenueResponse

	if  err := json.Unmarshal([]byte(output), &result); 
		err != nil {
		return nil, fmt.Errorf(
			"failed to parse revenue response: %w\nresponse=%s",
			err,
			output,
		)
	}

	return &result, nil
}
