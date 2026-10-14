package openai

import (
	"context"
	"strings"

	cfg "core/internal/config"

	sdk "github.com/openai/openai-go/v3"
	option "github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
)

type ClientStruct struct { sdk.Client }
var Client *ClientStruct

func Configure(
	config *cfg.OpenAIConfig,
) {
	Client = &ClientStruct{
		Client: sdk.NewClient(
			option.WithAPIKey(
				config.SecretKey,
			),
		),
	}
}

func format(
	name string,
	schema map[string]any,
) responses.ResponseTextConfigParam {
	return responses.ResponseTextConfigParam{
		Format: responses.ResponseFormatTextConfigUnionParam{
			OfJSONSchema: &responses.ResponseFormatTextJSONSchemaConfigParam{
				Name:   name,
				Schema: schema,
				Strict: sdk.Bool(true),
			},
		},
	}
}

func (c *ClientStruct) search(
	input string,
	text responses.ResponseTextConfigParam,
) (string, error) {
	response, err := c.Client.Responses.New(
		context.Background(),
		responses.ResponseNewParams{
			Model: sdk.ChatModelGPT4_1Mini,
			Input: responses.ResponseNewParamsInputUnion{
				OfString: sdk.String(input),
			},
			Text: text,
			Tools: []responses.ToolUnionParam{
				{
					OfWebSearch: &responses.WebSearchToolParam{
						Type: responses.WebSearchToolTypeWebSearch,
					},
				},
			},
		},
	)

	if err != nil {
		return "", err
	}

	output := strings.TrimSpace(response.OutputText())
	output = strings.TrimPrefix(output, "```json")
	output = strings.TrimSuffix(output, "```")

	return strings.TrimSpace(output), nil
}
