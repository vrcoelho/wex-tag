package external

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"wex/src/application"
)

func TestExternalCall(t *testing.T) {

	date := time.Now()
	country := "Mexico"
	currency := "Peso"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/services/api/fiscal_service/v1/accounting/od/rates_of_exchange"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected to request '%s', got: %s", expectedPath, r.URL.Path)
		}

		fields := r.URL.Query().Get("fields")
		if fields != "country_currency_desc,exchange_rate,record_date" {
			t.Error("Request without required fields")
		}

		filter := r.URL.Query().Get("filter")
		if !strings.Contains(filter, "record_date") {
			t.Error("Request without date filter")
		}

		sort := r.URL.Query().Get("sort")
		if sort != "-record_date" {
			t.Error("Request with wrong sort order")
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":[{"country_currency_desc":"Mexico-Peso","exchange_rate":"17.077","record_date":"2023-06-30"}]}`))
	}))
	defer server.Close()

	f := FiscalDataMiddleware{server.URL}

	_, err := f.QueryRates(
		country, currency, application.Time{date})

	if err != nil {
		t.Errorf("Error querying rates: %v", err)
	}
}
