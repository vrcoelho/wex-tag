package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
	"wex/src/application"
	"wex/src/persistance"
)

func TestRegisterBadRequest(t *testing.T) {
	driver := persistance.StartDriver()

	req := httptest.NewRequest(http.MethodGet, "/registerTransaction", nil)
	res := httptest.NewRecorder()

	getRegisterTransaction(driver)(res, req)

	if res.Code != http.StatusBadRequest {
		t.Errorf("got status %d but expected %d", res.Code, http.StatusBadRequest)
	}
}

func TestQueryBadRequest(t *testing.T) {
	driver := persistance.StartDriver()

	req := httptest.NewRequest(http.MethodGet, "/queryTransaction", nil)
	res := httptest.NewRecorder()

	getQueryTransactionHandler(driver)(res, req)

	if res.Code != http.StatusBadRequest {
		t.Errorf("got status %d but expected %d", res.Code, http.StatusBadRequest)
	}
}

func TestRegisterOK(t *testing.T) {
	driver := persistance.StartDriver()
	form := url.Values{}
	form.Add("description", "Sample Transaction")
	form.Add("date", time.Now().Format(time.RFC3339))
	form.Add("amount", "99.99")
	req := httptest.NewRequest(
		http.MethodPost, "/registerTransaction", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res := httptest.NewRecorder()

	getRegisterTransaction(driver)(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("got status %d but expected %d", res.Code, http.StatusOK)
	}

	var resp map[string]string

	err := json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		t.Errorf("Could not parse json response: %v", err)
	}

	transactionId := resp["transactionId"]

	time.Sleep(1 * time.Second)

	param := url.Values{}
	param.Add("transactionId", transactionId)

	url := "/queryTransaction" + "?" + param.Encode()

	req = httptest.NewRequest(
		http.MethodGet, url, nil)

	res = httptest.NewRecorder()

	getQueryTransactionHandler(driver)(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("got status %d but expected %d", res.Code, http.StatusOK)
	}
}

type MockExternalApi struct {
}

func (m MockExternalApi) QueryRates(
	country, currency string, date application.Time) ([]map[string]string, error) {

	mockResponse := []byte(`[{"country_currency_desc":"Mexico-Peso","exchange_rate":"17.077","record_date":"2023-06-30"}]`)

	var response []map[string]string
	_ = json.Unmarshal(mockResponse, &response)
	return response, nil

}

type MockDriver struct {
}

func (m MockDriver) QueryTransaction(transactionId string) (application.IdentifiedTransaction, error) {

	value, _ := application.NewMoney("10.59")
	return application.IdentifiedTransaction{
		Transaction: application.Transaction{
			Description: "Mocking driver test",
			Amount:      value,
			Date:        application.Time{time.Now()},
		},
		Uid: transactionId}, nil

}

func (m MockDriver) RegisterTransaction(tran application.Transaction) string {
	return ""
}

func TestConversionHandle(t *testing.T) {

	driver := MockDriver{}

	params := url.Values{}
	params.Add("transactionId", "182D05C0-DCC8-3EEC-119A-FB708B0A6BB8")
	params.Add("country", "Mexico")
	params.Add("currency", "Peso")

	v, _ := url.QueryUnescape(params.Encode())
	url := "/convertTransaction" + v

	req := httptest.NewRequest(
		http.MethodGet, url, nil)
	res := httptest.NewRecorder()

	middleware := MockExternalApi{}

	getConvertTransaction(driver, middleware)(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("got status %d but expected %d", res.Code, http.StatusOK)
	}

}
