package handlers

import (
	"encoding/json"
	"net/http"

	"backend/internal/application/command"
	"backend/internal/application/dto"
	"backend/internal/application/query"
	"backend/internal/application/sync"
	"backend/internal/interfaces/http/presenter"
)

type ValueHandler struct {
	setSvc   *command.SetValueService
	getSvc   *query.GetValueService
	checkSvc *query.CheckConsistencyService
	syncSvc  *sync.ReconcilePendingService
}

func NewValueHandler(
	setSvc *command.SetValueService,
	getSvc *query.GetValueService,
	checkSvc *query.CheckConsistencyService,
	syncSvc *sync.ReconcilePendingService,
) *ValueHandler {
	return &ValueHandler{
		setSvc:   setSvc,
		getSvc:   getSvc,
		checkSvc: checkSvc,
		syncSvc:  syncSvc,
	}
}

func (h *ValueHandler) SetValue(w http.ResponseWriter, r *http.Request) {
	var req dto.SetValueCommand
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		presenter.Error(w, http.StatusBadRequest, "invalid_json", "invalid JSON body")
		return
	}
	res, err := h.setSvc.Execute(r.Context(), req)
	if err != nil {
		presenter.Error(w, http.StatusBadGateway, "set_value_failed", err.Error())
		return
	}
	presenter.JSON(w, http.StatusAccepted, res)
}

func (h *ValueHandler) GetValue(w http.ResponseWriter, r *http.Request) {
	res, err := h.getSvc.Execute(r.Context(), dto.GetValueQuery{Source: r.URL.Query().Get("source")})
	if err != nil {
		presenter.Error(w, http.StatusBadGateway, "get_value_failed", err.Error())
		return
	}
	presenter.JSON(w, http.StatusOK, res)
}

func (h *ValueHandler) Sync(w http.ResponseWriter, r *http.Request) {
	res, err := h.syncSvc.Execute(r.Context(), 100)
	if err != nil {
		presenter.Error(w, http.StatusBadGateway, "sync_failed", err.Error())
		return
	}
	presenter.JSON(w, http.StatusOK, res)
}

func (h *ValueHandler) Check(w http.ResponseWriter, r *http.Request) {
	res, err := h.checkSvc.Execute(r.Context())
	if err != nil {
		presenter.Error(w, http.StatusBadGateway, "check_failed", err.Error())
		return
	}
	presenter.JSON(w, http.StatusOK, res)
}
