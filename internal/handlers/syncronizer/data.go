package syncronizer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/tommyhedley/fiberytsheets/internal/lib/tsheets"
	"github.com/tommyhedley/fiberytsheets/internal/utils"
)

func Data(w http.ResponseWriter, r *http.Request) {
	type nextPageConfig struct {
		Page int `json:"page"`
	}
	type pagination struct {
		HasNext        bool           `json:"hasNext"`
		NextPageConfig nextPageConfig `json:"nextPageConfig"`
	}
	type parameters struct {
		RequestedType string         `json:"requestedType"`
		Types         []string       `json:"types"`
		Filter        map[string]any `json:"filter"`
		Account       struct {
			AccessToken string `json:"access_token"`
		} `json:"account"`
		LastSyncronized string                               `json:"lastSynchronizedAt"`
		Pagination      pagination                           `json:"pagination"`
		Schema          map[string]map[string]map[string]any `json:"schema"`
	}
	type response[T any] struct {
		Items               []T        `json:"items"`
		Pagination          pagination `json:"pagination"`
		SynchronizationType string     `json:"synchronizationType"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("unable to decode request parameters: %v", err))
		return
	}

	var lastSyncronized string

	if params.LastSyncronized != "" {
		lastSyncronizedTime, err := time.Parse(time.RFC3339, params.LastSyncronized)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("unable to parse last sync time: %v", err))
			return
		}
		lastSyncronized = lastSyncronizedTime.Format("2006-01-02T15:04:05-07:00")
	}

	var page int

	if params.Pagination.NextPageConfig.Page == 0 {
		page = 1
	} else {
		page = params.Pagination.NextPageConfig.Page
	}

	switch params.RequestedType {
	case "user":
		sync := "delta"
		if lastSyncronized == "" {
			sync = "full"
		}
		var active string

		if val, ok := params.Filter["inactiveUsers"].(bool); ok {
			if val {
				active = "both"
			} else {
				active = "yes"
			}
		}

		request := tsheets.UserRequest{
			Active:           active,
			Page:             page,
			SupplementalData: "no",
			ModifiedSince:    lastSyncronized,
		}

		items, more, requestError := request.GetUsers("https://rest.tsheets.com/api/v1/users", params.Account.AccessToken)
		if requestError.Err != nil {
			if requestError.RateLimit {
				utils.RespondWithTryLater(w, http.StatusTooManyRequests, fmt.Sprintf("rate limit reached: %v", requestError.Err))
				return
			}
			utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("error with user request: %v", requestError.Err))
			return
		}

		resp := response[tsheets.UserResponse]{
			Items: items,
			Pagination: pagination{
				HasNext: more,
				NextPageConfig: nextPageConfig{
					Page: page + 1,
				},
			},
			SynchronizationType: sync,
		}

		utils.RespondWithJSON(w, http.StatusOK, resp)
		return
	case "group":
		sync := "delta"
		if lastSyncronized == "" {
			sync = "full"
		}
		request := tsheets.GroupRequest{
			Page:             page,
			SupplementalData: "no",
			ModifiedSince:    lastSyncronized,
		}

		items, more, requestError := request.GetGroups("https://rest.tsheets.com/api/v1/groups", params.Account.AccessToken)
		if requestError.Err != nil {
			if requestError.RateLimit {
				utils.RespondWithTryLater(w, http.StatusTooManyRequests, fmt.Sprintf("rate limit reached: %v", requestError.Err))
				return
			}
			utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("error with user request: %v", requestError.Err))
			return
		}

		resp := response[tsheets.GroupResponse]{
			Items: items,
			Pagination: pagination{
				HasNext: more,
				NextPageConfig: nextPageConfig{
					Page: page + 1,
				},
			},
			SynchronizationType: sync,
		}

		utils.RespondWithJSON(w, http.StatusOK, resp)
		return
	default:
		utils.RespondWithError(w, http.StatusBadRequest, "invalid requested datatype")
		return
	}
}
