package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"core/internal/model"

	sdk "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

const emailPrompt = `
You are a deal sourcing associate at a private equity firm.
Write a first outreach email to the owner of the company below.

Company:
%s

Investment criteria:
%s

Rules:
- Address the owner by first name when owner_name is set, otherwise use "Hi there"
- Open with one specific detail about the company from the company data or findings
- Say in one sentence why the company fits the investment criteria
- Ask for a short call in the next two weeks
- Body is 80 to 140 words in plain text with no markdown and no links
- Sign off as "Best," with no name after it
- subject is under 8 words and names the company
- Do not mention fit scores, priorities, or internal research
- Do not invent facts

Return ONLY valid JSON in this format:
{
  "subject": "string",
  "body": "string"
}
`

var emailSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"subject": map[string]any{"type": "string"},
		"body":    map[string]any{"type": "string"},
	},
	"required":             []string{"subject", "body"},
	"additionalProperties": false,
}

type EmailResponse struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func (c *ClientStruct) Email(
	company model.Company,
	icp *model.ICPCriteria,
) (*EmailResponse, error) {
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
		emailPrompt,
		companyJSON,
		icpJSON,
	)

	response, err := c.Client.Responses.New(
		context.Background(),
		responses.ResponseNewParams{
			Model: sdk.ChatModelGPT4_1Mini,
			Input: responses.ResponseNewParamsInputUnion{
				OfString: sdk.String(input),
			},
			Text: format("email", emailSchema),
		},
	)

	if err != nil {
		return nil, err
	}

	output := strings.TrimSpace(response.OutputText())

	var result EmailResponse

	if err := json.Unmarshal([]byte(output), &result); 
		err != nil {
		return nil, fmt.Errorf(
			"failed to parse email response: %w\nresponse=%s",
			err,
			output,
		)
	}

	return &result, nil
}
