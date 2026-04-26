package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/insmtx/SingerOS/backend/internal/agent"
	"github.com/insmtx/SingerOS/backend/internal/agent/simplechat"
	"github.com/insmtx/SingerOS/backend/internal/worker"
	"github.com/spf13/cobra"
	"github.com/ygpkg/yg-go/logs"
)

var (
	simpleChatAPIKey  string
	simpleChatModel   string
	simpleChatBaseURL string
	simpleChatServer  string
)

var simpleChatCmd = &cobra.Command{
	Use:   "simplechat [question]",
	Short: "Start a chat session for testing Agent Runtime",
	Long:  `Start a chat session that directly invokes the Agent Runtime for testing purposes. If a question is provided as argument, it will answer and exit. Otherwise, it starts an interactive multi-turn chat.`,
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := simpleChatAPIKey
		if apiKey == "" {
			apiKey = os.Getenv("OPENAI_API_KEY")
		}

		if apiKey == "" {
			logs.Warn("OPENAI_API_KEY not set, please provide --api-key flag or set environment variable")
			logs.Warn("Example: singer worker simplechat --api-key sk-xxx")
			return
		}

		scCfg := &simplechat.Config{
			LLMProvider: "openai",
			APIKey:      apiKey,
			Model:       simpleChatModel,
			BaseURL:     simpleChatBaseURL,
		}

		ctx := context.Background()
		scRuntime, err := simplechat.New(ctx, scCfg)
		if err != nil {
			logs.Fatalf("Failed to create SimpleChat runtime: %v", err)
			return
		}

		workerCfg := &worker.WorkerConfig{
			Runtime:    scRuntime,
			ServerAddr: simpleChatServer,
		}

		w, err := worker.NewWorker(ctx, workerCfg)
		if err != nil {
			logs.Fatalf("Failed to create worker: %v", err)
			return
		}

		logs.Infof("Worker %s initialized", w.GetWorkerID())

		if len(args) > 0 {
			runSingleQuestion(ctx, w, args)
		} else {
			runChat(ctx, w)
		}
	},
}

func init() {
	simpleChatCmd.Flags().StringVar(&simpleChatAPIKey, "api-key", "", "OpenAI API key (or set OPENAI_API_KEY env)")
	simpleChatCmd.Flags().StringVar(&simpleChatModel, "model", "gpt-4", "LLM model to use")
	simpleChatCmd.Flags().StringVar(&simpleChatBaseURL, "base-url", "", "Custom API base URL")
	simpleChatCmd.Flags().StringVar(&simpleChatServer, "server", "", "Server URL for WebSocket connection (e.g., localhost:8080)")
	workerCmd.AddCommand(simpleChatCmd)
}

func runSingleQuestion(ctx context.Context, w *worker.Worker, args []string) {
	question := strings.Join(args, " ")
	if question == "" {
		logs.Warn("No question provided")
		return
	}

	logs.Infof("Asking: %s", question)

	req := &agent.RequestContext{
		Input: agent.InputContext{
			Type: agent.InputTypeMessage,
			Text: question,
		},
	}

	result, err := w.Run(ctx, req)
	if err != nil {
		logs.Fatalf("Failed to get answer: %v", err)
		return
	}

	fmt.Println(result.Message)
}

func runChat(ctx context.Context, w *worker.Worker) {
	logs.Info("Starting chat session. Type 'quit' or 'exit' to stop.")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}

		question := strings.TrimSpace(scanner.Text())
		if question == "" {
			continue
		}
		if strings.ToLower(question) == "quit" || strings.ToLower(question) == "exit" {
			logs.Info("Exiting chat")
			break
		}

		req := &agent.RequestContext{
			Input: agent.InputContext{
				Type: agent.InputTypeMessage,
				Text: question,
			},
		}

		result, err := w.Run(ctx, req)
		if err != nil {
			logs.Errorf("Failed to get answer: %v", err)
			continue
		}

		fmt.Println(result.Message)
	}

	if err := scanner.Err(); err != nil {
		logs.Errorf("Error reading input: %v", err)
	}
}
