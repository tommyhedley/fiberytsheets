package tsheets

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/go-querystring/query"
	"github.com/tommyhedley/fiberytsheets/internal/utils"
)

type UserRequest struct {
	Active           string `url:"active"`
	Page             int    `url:"page"`
	SupplementalData string `url:"supplemental_data"`
	ModifiedSince    string `url:"modified_since,omitempty"`
}

type UserResponse struct {
	Id         json.Number `json:"id" type:"string"`
	Name       string      `json:"display_name"`
	FirstName  string      `json:"first_name"`
	LastName   string      `json:"last_name"`
	Active     bool        `json:"active"`
	LastActive string      `json:"last_active"`
	GroupId    json.Number `json:"group_id" type:"string"`
	Email      string      `json:"email"`
	SyncAction string      `json:"__syncAction,omitempty"`
}

func (params *UserRequest) GetUsers(URL, token string) (items []UserResponse, more bool, reqError *utils.RequestError) {
	type response struct {
		Results struct {
			Users map[string]UserResponse `json:"users"`
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
	for _, item := range resp.Results.Users {
		item.SyncAction = "SET"
		items = append(items, item)
	}
	if resp.More {
		return items, true, utils.NewRequestError(nil, false)
	} else {
		return items, false, utils.NewRequestError(nil, false)
	}
}
