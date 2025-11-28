package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
)

const promptTemplate = `
You are a helpful assistant that helps people find information.
Use the following pieces of context to answer the question at the end.
If you don't know the answer, just say that you don't know, don't try to make up an answer.
Always answer in markdown format.

{{ .Context }}

Question: {{ .Question }}

Helpful answer in markdown:
`

type TemplateData struct {
	Context  string
	Question string
}

type QueryRequest struct {
	// Text
	Query string `json:"query"`
}

type QueryResponse struct {
	Data []string `json:"result"`
}

type QueryHandler struct {
	vectorClient *milvusclient.Client
	embedder     *embeddings.EmbedderImpl
	llm          *openai.LLM
}

func NewQueryHandler(client *milvusclient.Client, embedder *embeddings.EmbedderImpl, llm *openai.LLM) *QueryHandler {
	return &QueryHandler{
		vectorClient: client,
		embedder:     embedder,
		llm:          llm,
	}
}

func (h QueryHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req QueryRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	vector, err := h.embedder.EmbedQuery(ctx, req.Query)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to embed query %s", err.Error()), http.StatusInternalServerError)
		return
	}

	loadTask, err := h.vectorClient.LoadCollection(ctx, milvusclient.NewLoadCollectionOption("documents"))

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load collection: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	if err = loadTask.Await(ctx); err != nil {
		http.Error(w, fmt.Sprintf("Failed to load collection: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	result, err := h.vectorClient.Search(ctx, milvusclient.NewSearchOption("documents", 3, []entity.Vector{
		entity.FloatVector(vector),
	}).WithOutputFields("id", "metadata"),
	)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to search documents: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	res := []string{}

	for _, r := range result {
		data := r.GetColumn("metadata").FieldData().GetScalars().GetStringData().Data

		res = append(res, data...)
	}

	promptTemp := prompts.NewPromptTemplate(promptTemplate, []string{"Context", "Question"})

	prompt, err := promptTemp.Format(map[string]any{
		"Context":  strings.Join(res, "\n"),
		"Question": req.Query,
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to format prompt: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	llmResponse, err := h.llm.GenerateContent(ctx, []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, prompt),
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate response %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(llmResponse.Choices[0].Content))
}
