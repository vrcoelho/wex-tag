package external

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
	"wex/src/application"
)

const TreasuryApi string = "https://api.fiscaldata.treasury.gov"

type FiscalDataInterface interface {
	QueryRates(
		country, currency string, date application.Time) ([]map[string]string, error)
}

type FiscalDataMiddleware struct {
	ExternalApi string
}

func (f FiscalDataMiddleware) QueryRates(
	country, currency string, date application.Time) ([]map[string]string, error) {

	countryCurrencyDesc := fmt.Sprintf("%s-%s", country, currency)

	dateLowerBound := date.AddDate(0, -6, 0)
	currency_filter := fmt.Sprintf("(%s)", countryCurrencyDesc)
	date_filter := fmt.Sprintf("gte:%s,lte:%s", dateLowerBound.Format(time.DateOnly), date.ToString())

	params := url.Values{}
	params.Add("fields", "country_currency_desc,exchange_rate,record_date")
	params.Add("filter",
		fmt.Sprintf("country_currency_desc:in:%s", currency_filter)+","+
			fmt.Sprintf("record_date:%s", date_filter))
	params.Add("sort", "-record_date")

	completeUrl, _ := url.Parse(f.ExternalApi)
	completeUrl.Path = "/services/api/fiscal_service/v1/accounting/od/rates_of_exchange"

	v, _ := url.QueryUnescape(params.Encode())
	completeUrl.RawQuery = v

	res, err := http.Get(completeUrl.String())
	if err != nil {
		return nil, err
	}

	var resp map[string][]map[string]string

	err = json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		return nil, err
	}
	return resp["data"], nil
}
