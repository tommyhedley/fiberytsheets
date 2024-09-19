package oauth2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/go-querystring/query"
	"github.com/tommyhedley/fiberytsheets/internal/utils"
)

type AccessTokenRequest struct {
	GrantType    string `url:"grant_type"`
	ClientId     string `url:"client_id"`
	ClientSecret string `url:"client_secret"`
	Code         string `url:"code"`
	RedirectURI  string `url:"redirect_uri"`
}

type AccessTokenResponse struct {
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

func TokenHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Fields struct {
			CallbackURI string `json:"callback_uri"`
		} `json:"fields"`
		Code string `json:"code"`
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
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("unable to decode request parameters: %v", err))
		return
	}

	requestParams := AccessTokenRequest{
		GrantType:    "authorization_code",
		ClientId:     os.Getenv("TSHEETS_OAUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("TSHEETS_OAUTH_CLIENT_SECRET"),
		Code:         params.Code,
		RedirectURI:  params.Fields.CallbackURI,
	}

	accessToken, err := requestParams.GetToken("https://rest.tsheets.com/api/v1/grant")
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("error with access token request: %v", err))
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, response{
		AccessToken:  accessToken.AccessToken,
		RefreshToken: accessToken.RefreshToken,
		ExpiresOn:    time.Now().UTC().Add(time.Duration(accessToken.ExpiresIn) * time.Second).Format(time.RFC3339),
	})
}

func (params *AccessTokenRequest) GetToken(URL string) (AccessTokenResponse, error) {
	baseURL, err := url.Parse(URL)
	if err != nil {
		return AccessTokenResponse{}, fmt.Errorf("error parsing base url: %w", err)
	}

	body, err := query.Values(params)
	if err != nil {
		return AccessTokenResponse{}, fmt.Errorf("error extracting query struct values: %w", err)
	}

	req, err := http.NewRequest("POST", baseURL.String(), strings.NewReader(body.Encode()))
	if err != nil {
		return AccessTokenResponse{}, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return AccessTokenResponse{}, fmt.Errorf("error executing request: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode > 299 {
		return AccessTokenResponse{}, fmt.Errorf("request error: %d", res.StatusCode)
	}

	decoder := json.NewDecoder(res.Body)
	var resp AccessTokenResponse
	err = decoder.Decode(&resp)
	if err != nil {
		return AccessTokenResponse{}, fmt.Errorf("unable to decode response: %w", err)
	}
	return resp, nil
}
