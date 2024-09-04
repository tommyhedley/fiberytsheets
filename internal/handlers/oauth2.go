package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

type AuthParams struct {
	responseType string
	clientId     string
	redirectURI  string
	state        string
	displayMode  string
}

func (pms *AuthParams) buildPreAuthURL(baseURL string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	query := url.Values{}

	query.Add("response_type", pms.responseType)
	query.Add("client_id", pms.clientId)
	query.Add("redirect_uri", pms.redirectURI)
	query.Add("state", pms.state)
	if pms.displayMode != "" {
		query.Add("display_mode", pms.displayMode)
	}

	u.RawQuery = query.Encode()

	finalURL := u.String()

	return finalURL, nil
}

func requestAuthURL(reqURL string) (string, error) {
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		fmt.Println("Request Build Failed")
		return "", err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("GET Failed")
		return "", err
	}

	defer res.Body.Close()

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	fmt.Println(res)
	fmt.Println(string(body))

	return string(body), nil

}

func Authorize(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		CallbackURI string `json:"callback_uri"`
		State       string `json:"state"`
	}
	type response struct {
		URL string `json:"URL"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	clientID := os.Getenv("TSHEETS_OAUTH_CLIENT_ID")

	authParams := AuthParams{
		responseType: "code",
		clientId:     clientID,
		redirectURI:  params.CallbackURI,
		state:        params.State,
		displayMode:  "",
	}

	url, err := authParams.buildPreAuthURL("https://rest.tsheets.com/api/v1/authorize")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Authorization request url couldn't be formed")
		return
	}

	authURL, err := requestAuthURL(url)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error requesting auth URL")
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		URL: authURL,
	})
}
