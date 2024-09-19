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

type CurrentUserResponse struct {
	Id        json.Number `json:"id" type:"string"`
	Name      string      `json:"display_name"`
	FirstName string      `json:"first_name"`
	LastName  string      `json:"last_name"`
	Active    bool        `json:"active"`
	Email     string      `json:"email"`
}

type RefreshTokenRequest struct {
	GrantType    string `url:"grant_type"`
	ClientId     string `url:"client_id"`
	ClientSecret string `url:"client_secret"`
	RefreshToken string `url:"refresh_token"`
}

type RefreshTokenResponse struct {
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

func ValidateHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Id     string `json:"id"`
		Fields struct {
			Name         string `json:"name"`
			AccessToken  string `json:"access_token"`
			ExpiresOn    string `json:"expires_on"`
			RefreshToken string `json:"refresh_token"`
		} `json:"fields"`
	}
	type response struct {
		Name         string `json:"name"`
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

	refreshNeeded, err := RefreshNeeded(params.Fields.ExpiresOn, 24)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("error checking token expiration: %v", err))
		return
	}

	if refreshNeeded {
		requestParams := RefreshTokenRequest{
			GrantType:    "refresh_token",
			ClientId:     os.Getenv("TSHEETS_OAUTH_CLIENT_ID"),
			ClientSecret: os.Getenv("TSHEETS_OAUTH_CLIENT_SECRET"),
			RefreshToken: params.Fields.RefreshToken,
		}
		refreshToken, err := requestParams.Refresh("https://rest.tsheets.com/api/v1/grant", params.Fields.AccessToken)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("error with refresh token request: %v", err))
		}
		currentUser, err := Validate("https://rest.tsheets.com/api/v1/current_user", refreshToken.AccessToken)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("token validation error: %v", err))
		}
		utils.RespondWithJSON(w, http.StatusOK, response{
			Name:         currentUser.Email,
			AccessToken:  refreshToken.AccessToken,
			RefreshToken: refreshToken.RefreshToken,
			ExpiresOn:    time.Now().UTC().Add(time.Duration(refreshToken.ExpiresIn) * time.Second).Format(time.RFC3339),
		})
		return
	}

	currentUser, err := Validate("https://rest.tsheets.com/api/v1/current_user", params.Fields.AccessToken)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("token validation error: %v", err))
	}
	utils.RespondWithJSON(w, http.StatusOK, response{
		Name:         currentUser.Email,
		AccessToken:  params.Fields.AccessToken,
		RefreshToken: params.Fields.RefreshToken,
		ExpiresOn:    params.Fields.ExpiresOn,
	})
}

func Validate(URL, token string) (CurrentUserResponse, error) {
	type response struct {
		Results struct {
			Users map[string]CurrentUserResponse `json:"users"`
		} `json:"results"`
		More bool `json:"more"`
	}
	baseURL, err := url.Parse(URL)
	if err != nil {
		return CurrentUserResponse{}, fmt.Errorf("error parsing base url: %w", err)
	}

	req, err := http.NewRequest("GET", baseURL.String(), nil)
	if err != nil {
		return CurrentUserResponse{}, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return CurrentUserResponse{}, fmt.Errorf("error executing request: %w", err)
	}

	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)
	var resp response
	err = decoder.Decode(&resp)
	if err != nil {
		return CurrentUserResponse{}, fmt.Errorf("unable to decode response: %w", err)
	}

	for _, user := range resp.Results.Users {
		return user, nil
	}

	return CurrentUserResponse{}, fmt.Errorf("no users in response")
}

func (params *RefreshTokenRequest) Refresh(URL, token string) (RefreshTokenResponse, error) {
	baseURL, err := url.Parse(URL)
	if err != nil {
		return RefreshTokenResponse{}, fmt.Errorf("error parsing base url: %w", err)
	}

	body, err := query.Values(params)
	if err != nil {
		return RefreshTokenResponse{}, fmt.Errorf("error extracting query struct values: %w", err)
	}

	req, err := http.NewRequest("POST", baseURL.String(), strings.NewReader(body.Encode()))
	if err != nil {
		return RefreshTokenResponse{}, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return RefreshTokenResponse{}, fmt.Errorf("error executing request: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode > 299 {
		return RefreshTokenResponse{}, fmt.Errorf("request error: %d", res.StatusCode)
	}

	decoder := json.NewDecoder(res.Body)
	var resp RefreshTokenResponse
	err = decoder.Decode(&resp)
	if err != nil {
		return RefreshTokenResponse{}, fmt.Errorf("unable to decode response: %w", err)
	}
	return resp, nil
}

func RefreshNeeded(expiresOn string, hoursToRefresh int) (bool, error) {
	expiration, err := time.Parse(time.RFC3339, expiresOn)
	if err != nil {
		return false, fmt.Errorf("unable to parse token expiration time: %w", err)
	}
	deadline := expiration.Add(time.Duration(hoursToRefresh) * time.Hour)
	return time.Now().UTC().After(deadline), nil
}
