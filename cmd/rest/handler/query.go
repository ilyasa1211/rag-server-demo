package handler

import (
	"encoding/json"
	"net/http"

	"github.com/ilyasa1211/rag-server/cmd/rest/dto"
	"github.com/ilyasa1211/rag-server/internal/usecase"
)

type QueryRequest struct {
	// Text
	Query string `json:"query"`
}

type QueryResponse struct {
	Result []string `json:"result"`
}

type QueryHandler struct {
	UseCase usecase.QueryDocumentHandler
}

func NewQueryHandler(uc usecase.QueryDocumentHandler) QueryHandler {
	return QueryHandler{UseCase: uc}
}

func (h QueryHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req QueryRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	errWithCode, result := h.UseCase.Handle(ctx, usecase.QueryDocument(req))

	if errWithCode != nil {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(errWithCode.Code)

		err := json.NewEncoder(w).Encode(dto.ErrorResponse{
			Code:   errWithCode.Code,
			Detail: errWithCode.Error.Error(),
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(QueryResponse{
		Result: result.Contents,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
