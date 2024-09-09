package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

// BuildQueryParams takes a struct and turns it into URL query parameters based on the struct's JSON tags.
// It can now handle slices of strings or ints, converting them to comma-separated values.
func BuildQueryParams(queryStruct interface{}) (url.Values, error) {
	values := url.Values{}
	v := reflect.ValueOf(queryStruct)
	t := reflect.TypeOf(queryStruct)

	// Ensure the input is a struct
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("queryStruct must be a struct")
	}

	// Iterate over the fields in the struct
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Get the JSON tag from the struct field
		jsonTag := fieldType.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			// Skip fields without a JSON tag or fields marked to be ignored
			continue
		}

		// Handle slices of strings or ints by joining them into a comma-separated string
		switch field.Kind() {
		case reflect.Slice:
			// Convert slice to a slice of strings
			var sliceValues []string
			for j := 0; j < field.Len(); j++ {
				sliceValues = append(sliceValues, fmt.Sprintf("%v", field.Index(j).Interface()))
			}
			// Join the slice into a comma-separated string and add to query
			values.Add(jsonTag, strings.Join(sliceValues, ","))
		default:
			// For non-slice types, add the value as is
			values.Add(jsonTag, fmt.Sprintf("%v", field.Interface()))
		}
	}

	return values, nil
}

// FlexibleRequest is a generic function that builds the request, adds the queries, and decodes the response into the provided response struct.
// Now supports request body serialization.
func FlexibleRequest(method, baseURL string, queryStruct interface{}, bodyStruct interface{}, responseStruct interface{}, token *string) error {
	// Parse the base URL
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %v", err)
	}

	// Build query parameters
	queryParams, err := BuildQueryParams(queryStruct)
	if err != nil {
		return fmt.Errorf("error building query params: %v", err)
	}

	// Append query parameters to the URL
	parsedURL.RawQuery = queryParams.Encode()

	// Create the request body (if provided)
	var bodyBytes []byte
	if bodyStruct != nil {
		bodyBytes, err = json.Marshal(bodyStruct)
		if err != nil {
			return fmt.Errorf("error marshaling request body: %v", err)
		}
	}

	// Create the HTTP request
	req, err := http.NewRequest(method, parsedURL.String(), bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	// If a token is provided, add it to the Authorization header
	if token != nil && *token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *token))
	}

	// If a body is present, set the Content-Type to application/json
	if bodyStruct != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error executing request: %v", err)
	}
	defer resp.Body.Close() // Ensure the response body is closed

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status: %s", resp.Status)
	}

	// Handle JSON response
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		// Decode JSON response
		if err := json.NewDecoder(resp.Body).Decode(responseStruct); err != nil {
			return fmt.Errorf("error decoding JSON response: %v", err)
		}
		return nil
	}

	// If needed, handle other response types here
	return fmt.Errorf("unsupported content type: %s", contentType)
}
