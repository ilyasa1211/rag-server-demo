package usecase

import (
	"context"
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

type QueryDocument struct {
	Query string
}

type QueryDocumentResult struct {
	Contents []string
}

type QueryDocumentHandler struct {
	vectorClient *milvusclient.Client
	embedder     *embeddings.EmbedderImpl
	llm          *openai.LLM
}

func NewQueryDocumentHandler(client *milvusclient.Client, embedder *embeddings.EmbedderImpl, llm *openai.LLM) QueryDocumentHandler {
	return QueryDocumentHandler{
		vectorClient: client,
		embedder:     embedder,
		llm:          llm,
	}
}

func (h QueryDocumentHandler) Handle(ctx context.Context, q QueryDocument) (*ErrorWithStatusCode, QueryDocumentResult) {
	vector, err := h.embedder.EmbedQuery(ctx, q.Query)

	if err != nil {
		return NewErrorWithStatusCode(fmt.Errorf("failed to embed query %w", err), http.StatusInternalServerError), QueryDocumentResult{}
	}

	loadTask, err := h.vectorClient.LoadCollection(ctx, milvusclient.NewLoadCollectionOption("documents"))

	if err != nil {
		return NewErrorWithStatusCode(fmt.Errorf("failed to load collection: %w", err), http.StatusInternalServerError), QueryDocumentResult{}
	}

	if err = loadTask.Await(ctx); err != nil {
		return NewErrorWithStatusCode(fmt.Errorf("failed to load collection: %s", err), http.StatusInternalServerError), QueryDocumentResult{}
	}

	result, err := h.vectorClient.Search(ctx, milvusclient.NewSearchOption("documents", 3, []entity.Vector{
		entity.FloatVector(vector),
	}).WithOutputFields("id", "metadata"),
	)

	if err != nil {
		return NewErrorWithStatusCode(fmt.Errorf("failed to search documents: %s", err), http.StatusInternalServerError), QueryDocumentResult{}
	}

	res := []string{}

	for i, r := range result {
		if col, err := r.GetColumn("metadata").GetAsString(i); err == nil {
			res = append(res, col)
		}
	}

	promptTemp := prompts.NewPromptTemplate(promptTemplate, []string{"Context", "Question"})

	prompt, err := promptTemp.Format(map[string]any{
		"Context":  strings.Join(res, "\n"),
		"Question": q.Query,
	})

	if err != nil {
		return NewErrorWithStatusCode(fmt.Errorf("failed to format prompt: %w", err), http.StatusInternalServerError), QueryDocumentResult{}
	}

	llmResponse, err := h.llm.GenerateContent(ctx, []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, prompt),
	})

	if err != nil {
		return NewErrorWithStatusCode(fmt.Errorf("failed to generate response %w", err), http.StatusInternalServerError), QueryDocumentResult{}
	}

	r := QueryDocumentResult{
		Contents: []string{},
	}

	for _, choice := range llmResponse.Choices {
		r.Contents = append(r.Contents, choice.Content)
	}

	return nil, r
}
