package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/ilyasa1211/rag-server/handler"
	"github.com/ilyasa1211/rag-server/infra/milvus"
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
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	defer stop()
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

	vectorDbConnResultChan := make(chan struct {
		vectorClient *milvusclient.Client
		err          error
	}, 1)

	go func() {
		vectorDbTimeoutCtx, vectorDbTimeoutCancel := context.WithTimeout(ctx, 5*time.Second)

		defer vectorDbTimeoutCancel()

		log.Println("Connecting to milvus database...")
		vectorClient, err := milvusclient.New(vectorDbTimeoutCtx, &milvusclient.ClientConfig{
			Address: "localhost:19530",
			// APIKey:  "root:Milvus",
		})

		vectorDbConnResultChan <- struct {
			vectorClient *milvusclient.Client
			err          error
		}{vectorClient: vectorClient, err: err}
	}()

	vectorDbConnResult := <-vectorDbConnResultChan

	if vectorDbConnResult.err != nil {
		return vectorDbConnResult.err
	}

	vectorClient := vectorDbConnResult.vectorClient

	log.Println("Connected to milvus database")

	log.Println("Running database migrations...")
	if err := milvus.NewMilvusMigration(vectorClient).Run(ctx); err != nil {
		return err
	}
	log.Println("Database migrations completed")

	defer func() {
		log.Println("Closing Milvus client")
		if err := vectorClient.Close(context.Background()); err != nil {
			log.Printf("Failed to close Milvus client: %v\n", err)
			return
		}
		log.Println("Milvus client closed")
	}()

	httpSrv := newHttpServer(*flagAddr, vectorClient, llm, embedder)

	httpSrvErr := make(chan error, 1)

	go func() {
		defer close(httpSrvErr)

		httpSrvErr <- httpSrv.ListenAndServe()
	}()

	log.Println("Server listening on ", *flagAddr)

	select {
	case err := <-httpSrvErr:
		return err
	case <-ctx.Done():
		stop()
	}

	log.Println("Shutting down server...")
	if err := httpSrv.Shutdown(context.Background()); err != nil {
		log.Println("Server shutdown error")
		return err
	}
	log.Println("Server shutdown complete")

	return nil
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
