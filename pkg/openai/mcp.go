package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"core/internal/model"
	logger "core/pkg/log"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

const enrichPrompt = `
You are a research analyst for a private equity firm.
Score the lead below against the investment criteria.

Lead:
%s

Investment criteria:
%s

Tools:
- revenue_research returns revenue, headcount, and sources
- news_research returns recent news, risks, and sources

Rules:
- Call revenue_research when revenue_estimate or employees is empty
- Call news_research when the lead has no findings
- Do not call a tool when the lead already has the data
- fit_score is an integer from 0 to 100
- priority is "high" when fit_score is 70 or more
- priority is "medium" when fit_score is 40 to 69
- priority is "low" when fit_score is under 40
- fit_summary is 2 to 3 sentences that name the criteria the lead meets and misses
- Do not invent facts

Return ONLY valid JSON in this format:
{
  "fit_score": 0,
  "fit_summary": "string",
  "priority": "high | medium | low"
}
`

const askPrompt = `
You are a research analyst for a private equity firm.
Answer the question about the lead below.

Lead:
%s

Investment criteria:
%s

Chat history:
%s

Question:
%s

Tools:
- revenue_research returns revenue, headcount, and sources
- news_research returns recent news, risks, and sources

Rules:
- Call revenue_research when the question asks for revenue or headcount
- Call news_research when the question asks for news, events, or risks
- Use the lead data when it answers the question
- Cite source URLs from tool results
- Say the data is not available when no source gives it
- Answer in plain text in 4 sentences or fewer
`

var OfFunctionCallOutput = responses.ResponseInputItemParamOfFunctionCallOutput[string]

var tools = []responses.ToolUnionParam{
	{
		OfFunction: &responses.FunctionToolParam{
			Name:        "news_research",
			Description: openai.String("Search the web for recent news, risks, and events about the lead"),
			Strict:      openai.Bool(true),
			Parameters: map[string]any{
				"type":                 "object",
				"properties":           map[string]any{},
				"additionalProperties": false,
			},
		},
	},
	{
		OfFunction: &responses.FunctionToolParam{
			Name:        "revenue_research",
			Description: openai.String("Search the web for the revenue and headcount of the lead"),
			Strict:      openai.Bool(true),
			Parameters: map[string]any{
				"type":                 "object",
				"properties":           map[string]any{},
				"additionalProperties": false,
			},
		},
	},
}

var enrichSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"fit_score":   map[string]any{"type": "integer"},
		"fit_summary": map[string]any{"type": "string"},
		"priority":    map[string]any{"type": "string", "enum": []string{"high", "medium", "low"}},
	},
	"required":             []string{"fit_score", "fit_summary", "priority"},
	"additionalProperties": false,
}

type AgentResult struct {
	Output  string
	News    *NewsResponse
	Revenue *RevenueResponse
}

type EnrichResponse struct {
	FitScore   int    `json:"fit_score"`
	FitSummary string `json:"fit_summary"`
	Priority   string `json:"priority"`

	News    *NewsResponse    `json:"-"`
	Revenue *RevenueResponse `json:"-"`
}

func (c *ClientStruct) Enrich(
	company model.Company,
	icp *model.ICPCriteria,
) (*EnrichResponse, error) {
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
		enrichPrompt,
		companyJSON,
		icpJSON,
	)

	agent, err := c.agent(company, icp, input, format("enrich", enrichSchema))

	if err != nil {
		return nil, err
	}

	var result EnrichResponse

	if err := json.Unmarshal([]byte(agent.Output), &result); err != nil {
		return nil, fmt.Errorf(
			"failed to parse enrich response: %w\nresponse=%s",
			err,
			agent.Output,
		)
	}

	result.News = agent.News
	result.Revenue = agent.Revenue

	return &result, nil
}

func (c *ClientStruct) Ask(
	company model.Company,
	icp *model.ICPCriteria,
	history []model.ResearchMessage,
	question string,
) (*AgentResult, error) {
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

	historyJSON, err := json.MarshalIndent(
		history,
		"",
		"  ",
	)

	if err != nil {
		return nil, err
	}

	input := fmt.Sprintf(
		askPrompt,
		companyJSON,
		icpJSON,
		historyJSON,
		question,
	)

	return c.agent(company, icp, input, responses.ResponseTextConfigParam{})
}

func (c *ClientStruct) agent(
	company model.Company,
	icp *model.ICPCriteria,
	input string,
	text responses.ResponseTextConfigParam,
) (*AgentResult, error) {
	result := AgentResult{}

	response, err := c.Client.Responses.New(
		context.Background(),
		responses.ResponseNewParams{
			Model: openai.ChatModelGPT4_1Mini,
			Input: responses.ResponseNewParamsInputUnion{
				OfString: openai.String(input),
			},
			Text:  text,
			Tools: tools,
		},
	)

	if err != nil {
		return nil, err
	}

	for {
		outputs := responses.ResponseInputParam{}

		for _, item := range response.Output {
			if item.Type != "function_call" {
				continue
			}

			call := item.AsFunctionCall()

			var data []byte

			switch call.Name {
			case "news_research":
				result.News, err = c.NewsResearch(company, icp)
				data, _ = json.Marshal(result.News)
			case "revenue_research":
				result.Revenue, err = c.RevenueResearch(company)
				data, _ = json.Marshal(result.Revenue)
			default:
				err = fmt.Errorf("unknown tool %s", call.Name)
			}

			if err != nil {
				logger.Error(
					"Tool %s failed: %v",
					call.Name,
					err,
				)

				data, _ = json.Marshal(map[string]string{"error": err.Error()})
			}

			output := OfFunctionCallOutput(string(data))
			output.OfFunctionCallOutput.CallID = openai.String(call.CallID)
			outputs = append(outputs, output)
		}

		if len(outputs) == 0 {
			break
		}

		response, err = c.Client.Responses.New(
			context.Background(),
			responses.ResponseNewParams{
				Model:              openai.ChatModelGPT4_1Mini,
				PreviousResponseID: openai.String(response.ID),
				Input: responses.ResponseNewParamsInputUnion{
					OfInputItemList: outputs,
				},
				Text:  text,
				Tools: tools,
			},
		)

		if err != nil {
			return nil, err
		}
	}

	result.Output = strings.TrimSpace(response.OutputText())
	result.Output = strings.TrimPrefix(result.Output, "```json")
	result.Output = strings.TrimSuffix(result.Output, "```")

	return &result, nil
}
