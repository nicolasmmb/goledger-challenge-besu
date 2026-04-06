package handlers

import (
	"net/http"

	"backend/internal/application/query"
	"backend/internal/interfaces/http/presenter"
)

type TransactionHandler struct {
	svc *query.GetTransactionStatusService
}

func NewTransactionHandler(svc *query.GetTransactionStatusService) *TransactionHandler {
	return &TransactionHandler{svc: svc}
}

func (h *TransactionHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	txHash := r.PathValue("txHash")
	if txHash == "" {
		presenter.Error(w, http.StatusBadRequest, "invalid_tx_hash", "tx hash is required")
		return
	}
	refresh := r.URL.Query().Get("refresh") == "true"
	res, err := h.svc.Execute(r.Context(), txHash, refresh)
	if err != nil {
		if query.IsNotFound(err) {
			presenter.Error(w, http.StatusNotFound, "tx_not_found", "transaction not found")
			return
		}
		presenter.Error(w, http.StatusBadGateway, "tx_status_failed", err.Error())
		return
	}
	presenter.JSON(w, http.StatusOK, res)
}
