package currency

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/k0kubun/pp"
)

type ExchangeRateResponse struct {
	Result              string             `json:"result"`
	Base                string             `json:"base_code"`
	LastUpdatedUnixTime int                `json:"time_last_update_unix"`
	Rates               map[string]float64 `json:"conversion_rates"`
}

var BaseUrl = "https://v6.exchangerate-api.com/v6/c3d03f555eb8293880257456/latest/"

func Exchange(baseCurrency string) error {
	url := BaseUrl + baseCurrency
	
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var exchangeRateResponse ExchangeRateResponse
	err = json.Unmarshal(body, &exchangeRateResponse)
	if err != nil {
		return err
	}

	if exchangeRateResponse.Result == "success" {
		pp.Println(exchangeRateResponse)
	}

	return nil
}
