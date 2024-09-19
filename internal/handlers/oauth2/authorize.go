package oauth2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/tommyhedley/fiberytsheets/internal/utils"
)

func AuthorizeHandler(w http.ResponseWriter, r *http.Request) {
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
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("unable to decode request parameters: %v", err))
		return
	}

	redirectURI, err := url.Parse("https://rest.tsheets.com/api/v1/authorize")
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, fmt.Sprintf("error parsing base url: %v", err))
		return
	}

	queryParams := url.Values{}
	queryParams.Add("response_type", "code")
	queryParams.Add("client_id", os.Getenv("TSHEETS_OAUTH_CLIENT_ID"))
	queryParams.Add("redirect_uri", params.CallbackURI)
	queryParams.Add("state", params.State)

	redirectURI.RawQuery = queryParams.Encode()

	utils.RespondWithJSON(w, http.StatusOK, response{
		RedirectURI: redirectURI.String(),
	})
}
