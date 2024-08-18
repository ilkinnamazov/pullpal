// main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/v63/github"
	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

func setupClients() (*openai.Client, *github.Client, error) {
	githubToken := os.Getenv("GITHUB_TOKEN")
	openAIToken := os.Getenv("OPENAI_TOKEN")

	openAIClient := openai.NewClient(openAIToken)
	githubClient := github.NewClient(nil).WithAuthToken(githubToken)

	if githubToken == "" {
		return nil, nil, fmt.Errorf("GITHUB_TOKEN environment variable not set")
	}

	if openAIToken == "" {
		return nil, nil, fmt.Errorf("OPENAI_TOKEN environment variable not set")
	}

	return openAIClient, githubClient, nil
}

func getPRDiff(ctx context.Context, githubClient *github.Client) string {
	var prNumber int
	fmt.Print("PR number: ")
	fmt.Scan(&prNumber)

	owner := os.Getenv("OWNER")
	repo := os.Getenv("REPO")
	diff, _, _ := githubClient.PullRequests.GetRaw(ctx, owner, repo, prNumber, github.RawOptions{Type: github.Diff})

	return diff
}

func generatePRDescription(ctx context.Context, openAIClient *openai.Client, diff string) (string, error) {
	prompt := fmt.Sprintf(`Generate a detailed pull request (PR) description based on the following diff data:\n\n%s\n\nThe PR description should include:\n- A summary of the changes\n- The purpose of the changes\n- Any additional context\n\nOutput the description in Markdown format.`, diff)

	message := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: prompt,
	}

	req := openai.ChatCompletionRequest{
		Model:       openai.GPT3Dot5Turbo,
		Messages:    []openai.ChatCompletionMessage{message},
		Temperature: 0.7,
	}

	resp, err := openAIClient.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to generate PR description: %w", err)
	}

	if len(resp.Choices) > 0 {
		return resp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no choices returned in the OpenAI response")
}

func main() {
	ctx := context.Background()
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	openAIClient, githubClient, err := setupClients()
	if err != nil {
		fmt.Println("Error setting up clients:", err)
		return
	}

	diff := getPRDiff(ctx, githubClient)

	description, err := generatePRDescription(ctx, openAIClient, diff)
	if err != nil {
		fmt.Println("Error generating PR description:", err)
		return
	}

	fmt.Println("Generated PR Description:")
	fmt.Println(description)
}
