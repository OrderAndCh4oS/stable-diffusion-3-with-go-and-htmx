package handler

import (
	"net/http"
	"token-based-payment-service-api/view/dashboard"
)

func HandleDashboardIndex(w http.ResponseWriter, r *http.Request) error {
	return Render(w, r, dashboard.Index())
}
