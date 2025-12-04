package handler

import (
	"encoding/json"
	"net/http"

	"github.com/ilyasa1211/rag-server/cmd/rest/dto"
	"github.com/ilyasa1211/rag-server/internal/usecase"
)

type AddRequest struct {
	// Text
	Document []string `json:"document"`
}
type AddResponse struct {
	Message string `json:"message"`
}

type AddHandler struct {
	UseCase usecase.AddDocumentHandler
}

func NewAddHandler(uc usecase.AddDocumentHandler) *AddHandler {
	return &AddHandler{UseCase: uc}
}

func (h *AddHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req AddRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	errWithCode, _ := h.UseCase.Handle(ctx, usecase.AddDocument{
		Documents: req.Document,
	})

	if errWithCode != nil {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(errWithCode.Code)

		if err := json.NewEncoder(w).Encode(dto.ErrorResponse{
			Code:   errWithCode.Code,
			Detail: errWithCode.Error.Error(),
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(AddResponse{
		Message: "success",
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}
