package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"

	// "go/types"
	"os"
	// "path/filepath"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func main() {
	var prompt string
	flag.StringVar(&prompt, "p", "", "Prompt to send to LLM")
	flag.Parse()

	if prompt == "" {
		panic("Prompt must not be empty")
	}

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	baseUrl := os.Getenv("OPENROUTER_BASE_URL")
	if baseUrl == "" {
		baseUrl = "https://openrouter.ai/api/v1"
	}

	if apiKey == "" {
		panic("Env variable OPENROUTER_API_KEY not found")
	}

	params := openai.ChatCompletionNewParams{
			Model: "anthropic/claude-haiku-4.5",
			Messages: []openai.ChatCompletionMessageParamUnion{
				{
					OfUser: &openai.ChatCompletionUserMessageParam{
						Content: openai.ChatCompletionUserMessageParamContentUnion{
							OfString: openai.String(prompt),
						},
					},
				},
			},
			Tools: []openai.ChatCompletionToolUnionParam{
				openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
					Name:        "read",
					Description: openai.String("Give LLM model access to a file"),
					Parameters: openai.FunctionParameters(map[string]any{
						"type": "object",
						"properties": map[string]any{
							"file_path": map[string]any{
								"type": "string",
								"description": "Filepath of file to give access to",
							},
						},
						"required": []string{"file_path"},
					}),
				}),
			},
		}

	client := openai.NewClient(option.WithAPIKey(apiKey), option.WithBaseURL(baseUrl))
	resp, err := client.Chat.Completions.New(context.Background(), params)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if len(resp.Choices) == 0 {
		panic("No choices in response")
	}

	toolCalls := resp.Choices[0].Message.ToolCalls

	if len(toolCalls) == 0 {
		panic("No tool calls in response")
	}

	params.Messages = append(params.Messages, resp.Choices[0].Message.ToParam())
	for _, toolCall := range toolCalls {
		if toolCall.Function.Name == "read" {
			var args map[string]any

			err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
			if err != nil {
				panic(err)
			}

			filePath := args["file_path"].(string)
			fileContent, err := os.ReadFile(filePath)
			if err != nil {
				panic(err)
			}

			fileContentString := string(fileContent)
			fmt.Printf("%s", fileContentString)
			return 

			// params.Messages = append(params.Messages, openai.ToolMessage(fileContentString, toolCall.ID))
		}
	}

	resp, err = client.Chat.Completions.New(context.Background(), params)
	if err != nil {
		panic(err)
	}

	fmt.Fprintln(os.Stderr, "Logs from your program will appear here!")

	fmt.Print(resp.Choices[0].Message.Content)
}
