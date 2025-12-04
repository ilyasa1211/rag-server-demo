package main

import (
	"net/http"

	"github.com/ilyasa1211/rag-server/cmd/rest/handler"
	"github.com/ilyasa1211/rag-server/internal/usecase"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
)

func newHttpServer(addr string, client *milvusclient.Client, llm *openai.LLM, embedder *embeddings.EmbedderImpl) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /add", handler.NewAddHandler(usecase.NewAddDocumentHandler(client, embedder)).Handle)
	mux.HandleFunc("POST /query", handler.NewQueryHandler(usecase.NewQueryDocumentHandler(client, embedder, llm)).Handle)

	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}
}
