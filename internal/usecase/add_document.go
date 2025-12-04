package usecase

import (
	"context"
	"fmt"
	"net/http"

	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/tmc/langchaingo/embeddings"
)

type AddDocument struct {
	Documents []string
}

type AddDocumentResult struct {
	IDs []int64
}

type AddDocumentHandler struct {
	vectorClient *milvusclient.Client
	embedder     *embeddings.EmbedderImpl
}

func NewAddDocumentHandler(client *milvusclient.Client, embedder *embeddings.EmbedderImpl) AddDocumentHandler {
	return AddDocumentHandler{
		vectorClient: client,
		embedder:     embedder,
	}
}

func (h AddDocumentHandler) Handle(ctx context.Context, cmd AddDocument) (*ErrorWithStatusCode, AddDocumentResult) {
	// read the document from the request body
	docs, err := h.embedder.EmbedDocuments(ctx, cmd.Documents)

	if err != nil {
		return NewErrorWithStatusCode(fmt.Errorf("failed to embed document %w", err), http.StatusInternalServerError), AddDocumentResult{}
	}

	// store the document in the database or in-memory storage
	insertResult, err := h.vectorClient.Insert(ctx, milvusclient.NewColumnBasedInsertOption("documents").
		// WithInt64Column("id", []int64{1}).
		WithFloatVectorColumn("vector", 768, docs).
		WithVarcharColumn("metadata", cmd.Documents),
	)

	if err != nil {
		return NewErrorWithStatusCode(fmt.Errorf("failed to store document %w", err), http.StatusInternalServerError), AddDocumentResult{}
	}

	r := AddDocumentResult{
		IDs: []int64{},
	}

	for i := 0; i < insertResult.IDs.Len(); i++ {
		if id, err := insertResult.IDs.GetAsInt64(i); err != nil {
			r.IDs = append(r.IDs, id)
		}
	}

	return nil, r
}
