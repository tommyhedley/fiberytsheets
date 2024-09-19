package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/google/go-querystring/query"
)

type ResponseData[T any] struct {
	Results struct {
		Items map[string]T `json:"-"`
	} `json:"results"`
	More bool `json:"more"`
}

func (rd *ResponseData[T]) DecodeBody(r io.Reader, fieldName string) error {
	var rawResults struct {
		Results map[string]json.RawMessage `json:"results"`
		More    bool                       `json:"more"`
	}
	decoder := json.NewDecoder(r)
	err := decoder.Decode(&rawResults)
	if err != nil {
		return err
	}
	rd.More = rawResults.More

	// Extract the items using the dynamic field name
	itemsData, ok := rawResults.Results[fieldName]
	if !ok {
		return fmt.Errorf("expected field '%s' not found in results", fieldName)
	}
	var items map[string]T
	err = json.Unmarshal(itemsData, &items)
	if err != nil {
		return err
	}
	rd.Results.Items = items
	return nil
}

func (rd *ResponseData[T]) ExtractItems() ([]T, bool) {
	var items []T
	for _, item := range rd.Results.Items {
		items = append(items, item)
	}
	return items, rd.More
}

func GetData[Req any, Res any](params Req, URL, token string, fieldName string) ([]Res, bool, *RequestError) {
	// Build the request URL with query parameters
	baseURL, err := url.Parse(URL)
	if err != nil {
		return nil, false, NewRequestError(fmt.Errorf("error parsing base URL: %w", err), false)
	}

	queryParams, err := query.Values(params)
	if err != nil {
		return nil, false, NewRequestError(fmt.Errorf("error extracting query parameters: %w", err), false)
	}

	baseURL.RawQuery = queryParams.Encode()

	// Create the HTTP request
	req, err := http.NewRequest("GET", baseURL.String(), nil)
	if err != nil {
		return nil, false, NewRequestError(fmt.Errorf("error creating request: %w", err), false)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	// Execute the HTTP request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, false, NewRequestError(fmt.Errorf("error executing request: %w", err), false)
	}
	defer res.Body.Close()

	// Handle HTTP errors
	if res.StatusCode > 299 {
		if res.StatusCode == 429 {
			return nil, false, NewRequestError(fmt.Errorf("rate limit reached: %d", res.StatusCode), true)
		}
		return nil, false, NewRequestError(fmt.Errorf("request error: %d", res.StatusCode), false)
	}

	// Decode the response body
	var response ResponseData[Res]
	err = response.DecodeBody(res.Body, fieldName)
	if err != nil {
		return nil, false, NewRequestError(fmt.Errorf("unable to decode response: %w", err), false)
	}

	// Extract items from the response
	items, more := response.ExtractItems()
	return items, more, nil
}
