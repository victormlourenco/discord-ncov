package model

import (
	"encoding/json"
	"net/http"
	"strings"
)

type Latest struct {
	Confirmed uint64 `json:"Confirmed"`
	Deaths    uint64 `json:"Deaths"`
	Recovered uint64 `json:"Recovered"`
	Country   string `json:"Country_Region"`
}

type Result struct {
	Features []struct {
		Attributes Latest `json:"attributes"`
	} `json:"features"`
}

func (l *Latest) Equals(la Latest) bool {
	return l.Deaths == la.Deaths && l.Confirmed == la.Confirmed && l.Recovered == la.Recovered && l.Country == la.Country
}

func (l *Latest) Get(country string, uri string) error {
	resp, err := http.Get(uri)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	result := Result{}
	err = decoder.Decode(&result)
	if err != nil {
		return err
	}
	latest := Latest{}
	if country == "" {
		latest.Country = "Global"
		for _, v := range result.Features {
			latest.Deaths += v.Attributes.Deaths
			latest.Confirmed += v.Attributes.Confirmed
			latest.Recovered += v.Attributes.Recovered
		}
	} else {
		for _, v := range result.Features {
			if strings.ToLower(country) == strings.ToLower(v.Attributes.Country) {
				latest.Country = v.Attributes.Country
				latest.Deaths += v.Attributes.Deaths
				latest.Confirmed += v.Attributes.Confirmed
				latest.Recovered += v.Attributes.Recovered
			}
		}
	}
	*l = latest
	return nil
}
