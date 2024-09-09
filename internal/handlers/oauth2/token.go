package oauth2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/tommyhedley/fiberytsheets/internal/utils"
)

func GetToken(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Fields struct {
			CallbackURI string `json:"callback_uri"`
		} `json:"fields"`
		Code string `json:"code"`
	}
	type innerResponse struct {
		AccessToken  string `json:"access_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
		Scope        string `json:"scope"`
		RefreshToken string `json:"refresh_token"`
		UserID       string `json:"user_id"`
		CompanyID    string `json:"company_id"`
		ClientURL    string `json:"client_url"`
		ClientType   string `json:"client_type"`
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
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("couldn't decode request body: %v", err))
		return
	}

	baseURL, err := url.Parse("https://rest.tsheets.com/api/v1/grant")
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "invalid base URL")
		return
	}

	body := url.Values{}
	body.Add("grant_type", "authorization_code")
	body.Add("client_id", os.Getenv("TSHEETS_OAUTH_CLIENT_ID"))
	body.Add("client_secret", os.Getenv("TSHEETS_OAUTH_CLIENT_SECRET"))
	body.Add("code", params.Code)
	body.Add("redirect_uri", params.Fields.CallbackURI)

	req, err := http.NewRequest("POST", baseURL.String(), strings.NewReader(body.Encode()))
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, fmt.Sprintf("error creating request: %v", err))
		return
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, fmt.Sprintf("error executing request: %v", err))
		return
	}

	defer res.Body.Close()

	if res.StatusCode > 200 {
		utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("request failed with error: %v", res.StatusCode))
		return
	}

	resDecoder := json.NewDecoder(res.Body)
	innerRes := innerResponse{}
	err = resDecoder.Decode(&innerRes)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("couldn't decode response body: %v", err))
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, response{
		AccessToken:  innerRes.AccessToken,
		RefreshToken: innerRes.RefreshToken,
		ExpiresOn:    time.Now().UTC().Add(time.Duration(innerRes.ExpiresIn) * time.Second).Format(time.RFC3339),
	})
}
