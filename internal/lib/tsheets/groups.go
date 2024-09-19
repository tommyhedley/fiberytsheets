package tsheets

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/go-querystring/query"
	"github.com/tommyhedley/fiberytsheets/internal/utils"
)

type GroupRequest struct {
	Active           string `url:"active, omitempty"`
	Page             int    `url:"page"`
	SupplementalData string `url:"supplemental_data"`
	ModifiedSince    string `url:"modified_since,omitempty"`
}

type GroupResponse struct {
	Id         json.Number `json:"id" type:"string"`
	Name       string      `json:"name"`
	Active     bool        `json:"active"`
	SyncAction string      `json:"__syncAction,omitempty"`
}

func (params *GroupRequest) GetGroups(URL, token string) (items []GroupResponse, more bool, reqError *utils.RequestError) {
	type response struct {
		Results struct {
			Items map[string]GroupResponse `json:"groups"`
		} `json:"results"`
		More bool `json:"more"`
	}
	baseURL, err := url.Parse(URL)
	if err != nil {
		return nil, false, utils.NewRequestError(fmt.Errorf("error parsing base url: %w", err), false)
	}

	queryParams, err := query.Values(params)
	if err != nil {
		return nil, false, utils.NewRequestError(fmt.Errorf("error extracting query struct values: %w", err), false)
	}

	baseURL.RawQuery = queryParams.Encode()

	req, err := http.NewRequest("GET", baseURL.String(), nil)
	if err != nil {
		return nil, false, utils.NewRequestError(fmt.Errorf("error creating request: %w", err), false)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, false, utils.NewRequestError(fmt.Errorf("error executing request: %w", err), false)
	}

	defer res.Body.Close()

	if res.StatusCode > 299 {
		if res.StatusCode == 429 {
			return nil, false, utils.NewRequestError(fmt.Errorf("rate limit reached: %d", res.StatusCode), true)
		}
		return nil, false, utils.NewRequestError(fmt.Errorf("request error: %d", res.StatusCode), false)
	}

	decoder := json.NewDecoder(res.Body)
	var resp response
	err = decoder.Decode(&resp)
	if err != nil {
		return nil, false, utils.NewRequestError(fmt.Errorf("unable to decode response: %w", err), false)
	}
	for _, item := range resp.Results.Items {
		item.SyncAction = "SET"
		items = append(items, item)
	}
	if resp.More {
		return items, true, utils.NewRequestError(nil, false)
	} else {
		return items, false, utils.NewRequestError(nil, false)
	}
}
