package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"wex/src/application"
	"wex/src/external"
	"wex/src/persistance"
)

func getRoot(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		http.ServeFile(w, r, "./../ui/form.html")
	default:
		badRequest(w, "Unsupported Method")
	}

}

func badRequest(w http.ResponseWriter, reason string) {
	errorMessage := fmt.Sprintf("Bad request: %v", reason)
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(errorMessage))
	log.Printf("(%v) %v", http.StatusBadRequest, errorMessage)
}

func getQueryTransactionHandler(driver persistance.PersistanceDriver) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			queryId := r.URL.Query().Get("transactionId")
			transaction, err := driver.QueryTransaction(queryId)
			if err != nil {
				badRequest(w, err.Error())
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(transaction)
			logMessage := fmt.Sprintf("Transaction queried: %v", transaction.Uid)
			log.Printf("(%v) %v", http.StatusOK, logMessage)
		default:
			badRequest(w, "Unsupported method")

		}
	}
}

func getRegisterTransaction(driver persistance.PersistanceDriver) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			if err := r.ParseForm(); err != nil {
				badRequest(w, "Could not parse form")
				return
			}
			description := r.FormValue("description")
			amount := r.FormValue("amount")
			date := r.FormValue("date")

			newTransaction, err := application.NewTransaction(
				description, date, amount,
			)

			if err != nil {
				badRequest(w,
					fmt.Sprintf("Could not create transaction: %v", err))
				return
			}

			newUid := driver.RegisterTransaction(newTransaction)
			w.Header().Set("Content-Type", "application/json")
			resp := make(map[string]string)
			resp["transactionId"] = newUid
			json.NewEncoder(w).Encode(resp)
			logMessage := fmt.Sprintf("Transaction registered: %v", newUid)
			log.Printf("(%v) %v", http.StatusOK, logMessage)
		default:
			badRequest(w, "Unsupported method")
		}
	}
}

func getConvertTransaction(driver persistance.PersistanceDriver,
	middleware external.FiscalDataInterface) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		transactionId := r.URL.Query().Get("transactionId")
		country := r.URL.Query().Get("country")
		currency := r.URL.Query().Get("currency")
		transaction, err := driver.QueryTransaction(transactionId)
		if err != nil {
			badRequest(w, err.Error())
			return
		}

		rates, err := middleware.QueryRates(country, currency, transaction.Date)
		if err != nil {
			badRequest(w, "error getting conversion rate")
			return
		}

		if len(rates) == 0 {
			badRequest(w, "no conversion rate is available within 6 months to purchase date; transaction cannot be converted to the target currency")
			return
		}

		// use first rate
		rate, err := application.NewMoney(rates[0]["exchange_rate"])

		converted := transaction.Amount.PreciseConvert(rate)

		resp := make(map[string]string)
		resp["uid"] = transaction.Uid
		resp["transactionDate"] = transaction.Date.ToString()
		resp["description"] = transaction.Description
		resp["originalValue"] = transaction.Amount.ToString()
		resp["convertedValue"] = converted.ToString()
		resp["exchangeRate"] = rate.ToString()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func main() {
	driver := persistance.StartDriver()

	f := external.FiscalDataMiddleware{ExternalApi: external.TreasuryApi}

	http.HandleFunc("/", getRoot)
	http.HandleFunc("/queryTransaction", getQueryTransactionHandler(driver))
	http.HandleFunc("/registerTransaction", getRegisterTransaction(driver))
	http.HandleFunc("/convertTransaction", getConvertTransaction(driver, f))

	err := http.ListenAndServe(":3333", nil)
	if err != nil {
		log.Fatalf("Server stopped: %v", err)
	}
}
