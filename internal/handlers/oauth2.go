package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"time"
)

type AuthParams struct {
	ResponseType string `json:"response_type"`
	ClientId     string `json:"client_id"`
	RedirectURI  string `json:"redirect_uri"`
	State        string `json:"state"`
	DisplayMode  string `json:"display_mode,omitempty"`
}

type TokenParams struct {
	GrantType    string `json:"grant_type"`
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Code         string `json:"code"`
	RedirectURI  string `json:"redirect_uri"`
}

type RefreshParams struct {
	GrantType    string `json:"grant_type"`
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RefreshToken string `json:"refresh_token"`
}

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

		// Convert the field value to a string and add it as a query parameter
		values.Add(jsonTag, fmt.Sprintf("%v", field.Interface()))
	}

	return values, nil
}

// FlexibleRequest is a generic function that builds the request, adds the queries, and decodes the response into the provided response struct
// or a string if the response is HTML.
func FlexibleRequest(method, baseURL string, queryStruct interface{}, responseStruct interface{}, token *string) (string, error) {
	// Parse the base URL
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %v", err)
	}

	// Build query parameters
	queryParams, err := BuildQueryParams(queryStruct)
	if err != nil {
		return "", fmt.Errorf("error building query params: %v", err)
	}

	// Append query parameters to the URL
	parsedURL.RawQuery = queryParams.Encode()

	// Create the HTTP request
	req, err := http.NewRequest(method, parsedURL.String(), nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	// If a token is provided, add it to the Authorization header
	if token != nil && *token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *token))
	}

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error executing request: %v", err)
	}
	defer resp.Body.Close()

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request failed with status: %s", resp.Status)
	}

	// Check the content type to handle JSON or HTML responses
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		// Decode JSON response
		if err := json.NewDecoder(resp.Body).Decode(responseStruct); err != nil {
			return "", fmt.Errorf("error decoding JSON response: %v", err)
		}
		return "", nil
	} else if strings.Contains(contentType, "text/html") {
		// Read HTML response as string
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("error reading HTML response: %v", err)
		}
		return string(body), nil
	}

	// Handle other content types if needed (e.g., text/plain, etc.)
	return "", fmt.Errorf("unsupported content type: %s", contentType)
}

func Authorize(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		CallbackURI string `json:"callback_uri"`
		State       string `json:"state"`
	}
	type response struct {
		RedirectURI string `json:"redirect_uri"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	authParams := AuthParams{
		ResponseType: "code",
		ClientId:     os.Getenv("TSHEETS_OAUTH_CLIENT_ID"),
		RedirectURI:  params.CallbackURI,
		State:        params.State,
		DisplayMode:  "",
	}

	html, err := FlexibleRequest("GET", "https://rest.tsheets.com/api/v1/authorize", authParams, nil, nil)
	if err != nil {

	}

	startDelimiter := "redirect_uri="
	endDelimiter := "'"

	startIndex := strings.Index(html, startDelimiter)
	if startIndex == -1 {
		respondWithError(w, http.StatusBadRequest, "no redirect_uri found in the response body")
		return
	}

	startIndex += len(startDelimiter)

	endIndex := strings.Index(html[startIndex:], endDelimiter)
	if endIndex == -1 {
		respondWithError(w, http.StatusBadRequest, "no closing single quote found after redirect_uri")
		return
	}

	redirectURI := strings.ReplaceAll(html[startIndex:startIndex+endIndex], `\`, "")

	respondWithJSON(w, http.StatusOK, response{
		RedirectURI: redirectURI,
	})
}

func GetToken(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Fields struct {
			Callback_uri string `json:"callback_uri"`
		} `json:"fields"`
		Code string `json:"code"`
	}
	type reqResponse struct {
		AccessToken  string `json:"access_token"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
	}
	type response struct {
		AccessToken  string `json:"access_token"`
		ExpiresOn    string `json:"expires_on"`
		RefreshToken string `json:"refresh_token"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	tokenParams := TokenParams{
		GrantType:    "authorization_code",
		ClientId:     os.Getenv("TSHEETS_OAUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("TSHEETS_OAUTH_CLIENT_SECRET"),
		Code:         params.Code,
		RedirectURI:  params.Fields.Callback_uri,
	}

	reqRes := reqResponse{}

	_, err = FlexibleRequest("POST", "https://rest.tsheets.com/api/v1/grant", tokenParams, reqRes, nil)
	if err != nil {
		fmt.Printf("request error %s", err)
		respondWithError(w, http.StatusUnauthorized, "Unable to get auth url")
		return
	}

	res := response{
		AccessToken:  reqRes.AccessToken,
		RefreshToken: reqRes.RefreshToken,
		ExpiresOn:    time.Now().UTC().Add(time.Duration(reqRes.ExpiresIn) * time.Second).Format(time.RFC3339),
	}

	respondWithJSON(w, http.StatusOK, res)
}

func ValidateRefresh(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		AccessToken  string `json:"access_token"`
		ExpiresOn    string `json:"expires_on"`
		RefreshToken string `json:"refresh_token"`
	}
	type reqResponse struct {
		AccessToken  string `json:"access_token"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
	}
	type response struct {
		AccessToken  string `json:"access_token"`
		ExpiresOn    string `json:"expires_on"`
		RefreshToken string `json:"refresh_token"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	refreshParams := RefreshParams{
		GrantType:    "refresh_token",
		ClientId:     os.Getenv("TSHEETS_OAUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("TSHEETS_OAUTH_CLIENT_SECRET"),
		RefreshToken: params.RefreshToken,
	}

	reqRes := reqResponse{}

	_, err = FlexibleRequest("POST", "https://rest.tsheets.com/api/v1/grant", refreshParams, reqRes, &params.AccessToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't retreive refresh token")
		return
	}

	res := response{
		AccessToken:  reqRes.AccessToken,
		RefreshToken: reqRes.RefreshToken,
		ExpiresOn:    time.Now().UTC().Add(time.Duration(reqRes.ExpiresIn) * time.Second).Format(time.RFC3339),
	}

	respondWithJSON(w, http.StatusOK, res)
}
