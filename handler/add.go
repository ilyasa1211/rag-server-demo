package handler

import (
	"encoding/json"
	"net/http"

	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/tmc/langchaingo/embeddings"
)

type AddRequest struct {
	// Text
	Document []string `json:"document"`
}

type AddHandler struct {
	vectorClient *milvusclient.Client
	embedder     *embeddings.EmbedderImpl
}

func NewAddHandler(client *milvusclient.Client, embedder *embeddings.EmbedderImpl) *AddHandler {
	return &AddHandler{
		vectorClient: client,
		embedder:     embedder,
	}
}

func (h *AddHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// read the document from the request body
	var req AddRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	docs, err := h.embedder.EmbedDocuments(ctx, req.Document)

	if err != nil {
		http.Error(w, "Failed to embed document", http.StatusInternalServerError)
	}

	// store the document in the database or in-memory storage
	_, err = h.vectorClient.Insert(ctx, milvusclient.NewColumnBasedInsertOption("documents").
		WithInt64Column("id", []int64{1}).
		WithFloatVectorColumn("vector", 768, docs).
		WithVarcharColumn("metadata", req.Document),
	)

	if err != nil {
		http.Error(w, "Failed to store document", http.StatusInternalServerError)
		return
	}

	// respond with success or error message
	b, err := json.Marshal(map[string]string{"status": "success"})

	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}
