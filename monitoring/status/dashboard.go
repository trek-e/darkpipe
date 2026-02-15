// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package status

import (
	"encoding/json"
	"html/template"
	"net/http"
)

// DashboardHandler serves the web dashboard for monitoring status
type DashboardHandler struct {
	aggregator *StatusAggregator
	template   *template.Template
}

// NewDashboardHandler creates a new dashboard handler with the given aggregator
func NewDashboardHandler(aggregator *StatusAggregator, tmpl *template.Template) *DashboardHandler {
	return &DashboardHandler{
		aggregator: aggregator,
		template:   tmpl,
	}
}

// HandleDashboard renders the HTML dashboard
func (h *DashboardHandler) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	status, err := h.aggregator.GetStatus(r.Context())
	if err != nil {
		http.Error(w, "Failed to get system status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.template.Execute(w, status); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

// HandleStatusAPI returns JSON status for AJAX or external tools
func (h *DashboardHandler) HandleStatusAPI(w http.ResponseWriter, r *http.Request) {
	status, err := h.aggregator.GetStatus(r.Context())
	if err != nil {
		http.Error(w, "Failed to get system status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(status); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		return
	}
}

// HandleDashboard returns a HandlerFunc that serves the dashboard
func HandleDashboard(aggregator *StatusAggregator, tmpl *template.Template) http.HandlerFunc {
	handler := NewDashboardHandler(aggregator, tmpl)
	return handler.HandleDashboard
}

// HandleStatusAPI returns a HandlerFunc that serves the JSON API
func HandleStatusAPI(aggregator *StatusAggregator) http.HandlerFunc {
	handler := &DashboardHandler{aggregator: aggregator}
	return handler.HandleStatusAPI
}
