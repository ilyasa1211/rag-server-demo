package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/ilyasa1211/rag-server/handler"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func run() error {
	// Initialize LLM
	llm, err := openai.New(
		openai.WithBaseURL("http://localhost:11434/v1"), // ollama
		openai.WithToken("token"),
		openai.WithModel("gemma3:1b-it-qat"),
	)

	if err != nil {
		return err
	}

	embeddingModel, err := openai.New(
		openai.WithBaseURL("http://localhost:11434/v1"), // ollama
		openai.WithToken("token"),
		openai.WithEmbeddingModel("embeddinggemma:latest"),
		openai.WithEmbeddingDimensions(768),
	)

	if err != nil {
		return err
	}

	embedder, err := embeddings.NewEmbedder(embeddingModel)
	if err != nil {
		return err
	}

	flagAddr := flag.String("listen", ":8080", "Listen Address")

	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGINT)

	defer stop()

	vectorClient, err := milvusclient.New(ctx, &milvusclient.ClientConfig{
		Address: "localhost:19530",
		APIKey:  "token",
	})

	defer func() {
		if err := vectorClient.Close(context.Background()); err != nil {
			fmt.Printf("Failed to close Milvus client: %v\n", err)
		}
	}()

	if err != nil {
		return err
	}

	httpSrv := newHttpServer(*flagAddr, vectorClient, llm, embedder)

	httpSrvErr := make(chan error, 1)

	go func() {
		defer close(httpSrvErr)

		httpSrvErr <- httpSrv.ListenAndServe()
	}()

	fmt.Println("Server listening on ", *flagAddr)

	select {
	case err := <-httpSrvErr:
		return err
	case <-ctx.Done():
		stop()
	}

	return httpSrv.Shutdown(context.Background())
}

func newHttpServer(addr string, client *milvusclient.Client, llm *openai.LLM, embedder *embeddings.EmbedderImpl) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /add", handler.NewAddHandler(client, embedder).Handle)
	mux.HandleFunc("POST /query", handler.NewQueryHandler(client, embedder, llm).Handle)

	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}
}
