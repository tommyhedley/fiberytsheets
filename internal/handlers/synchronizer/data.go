package synchronizer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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

	sync := "delta"
	if lastSyncronized == "" {
		sync = "full"
	}

	var page int

	if params.Pagination.NextPageConfig.Page == 0 {
		page = 1
	} else {
		page = params.Pagination.NextPageConfig.Page
	}

	switch params.RequestedType {
	case "user":
		var active string

		if val, ok := params.Filter["inactiveUsers"].(bool); ok {
			if val {
				active = "both"
			} else {
				active = "yes"
			}
		}

		type userRequest struct {
			Active           string `url:"active"`
			Page             int    `url:"page"`
			SupplementalData string `url:"supplemental_data"`
			ModifiedSince    string `url:"modified_since,omitempty"`
		}

		type userResponse struct {
			Id         json.Number `json:"id" type:"string"`
			Name       string      `json:"display_name"`
			FirstName  string      `json:"first_name"`
			LastName   string      `json:"last_name"`
			Active     bool        `json:"active"`
			LastActive string      `json:"last_active"`
			GroupID    json.Number `json:"group_id" type:"string"`
			Email      string      `json:"email"`
		}

		type item struct {
			Id         string `json:"id"`
			TimeID     string `json:"timeId"`
			DiplayName string `json:"display_name"`
			FirstName  string `json:"first_name"`
			LastName   string `json:"last_name"`
			Active     bool   `json:"active"`
			Email      string `json:"email"`
			LastActive string `json:"last_active"`
			SyncAction string `json:"__syncAction,omitempty"`
			GroupID    string `json:"group_id"`
		}

		userReq := userRequest{
			Active:           active,
			Page:             page,
			SupplementalData: "no",
			ModifiedSince:    lastSyncronized,
		}

		users, more, requestError := utils.GetData[userRequest, userResponse](&userReq, "https://rest.tsheets.com/api/v1/users", params.Account.AccessToken, "users")
		if requestError.Err != nil {
			if requestError.RateLimit {
				utils.RespondWithTryLater(w, http.StatusTooManyRequests, fmt.Sprintf("rate limit reached: %v", requestError.Err))
				return
			}
			utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("error with user request: %v", requestError.Err))
			return
		}

		var items []item

		for _, user := range users {
			if sync == "full" {
				items = append(items, item{
					Id:         user.Id.String(),
					TimeID:     user.Id.String(),
					DiplayName: user.Name,
					FirstName:  user.FirstName,
					LastName:   user.LastName,
					Active:     user.Active,
					Email:      user.Email,
					LastActive: user.LastActive,
					GroupID:    user.GroupID.String(),
				})
			} else {
				items = append(items, item{
					Id:         user.Id.String(),
					TimeID:     user.Id.String(),
					DiplayName: user.Name,
					FirstName:  user.FirstName,
					LastName:   user.LastName,
					Active:     user.Active,
					Email:      user.Email,
					LastActive: user.LastActive,
					GroupID:    user.GroupID.String(),
					SyncAction: "SET",
				})
			}
		}

		resp := response[item]{
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
		type groupRequest struct {
			Active           string `url:"active,omitempty"`
			Page             int    `url:"page"`
			SupplementalData string `url:"supplemental_data"`
			ModifiedSince    string `url:"modified_since,omitempty"`
		}

		type groupResponse struct {
			Id         json.Number `json:"id" type:"string"`
			Name       string      `json:"name"`
			Active     bool        `json:"active"`
			SyncAction string      `json:"__syncAction,omitempty"`
		}

		type item struct {
			Id         string `json:"id"`
			TimeID     string `json:"timeId"`
			Name       string `json:"name"`
			Active     bool   `json:"active"`
			SyncAction string `json:"__syncAction,omitempty"`
		}

		groupReq := groupRequest{
			Page:             page,
			SupplementalData: "no",
			ModifiedSince:    lastSyncronized,
		}

		groups, more, requestError := utils.GetData[groupRequest, groupResponse](&groupReq, "https://rest.tsheets.com/api/v1/groups", params.Account.AccessToken, "groups")
		if requestError.Err != nil {
			if requestError.RateLimit {
				utils.RespondWithTryLater(w, http.StatusTooManyRequests, fmt.Sprintf("rate limit reached: %v", requestError.Err))
				return
			}
			utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("error with user request: %v", requestError.Err))
			return
		}

		var items []item

		for _, group := range groups {
			if sync == "full" {
				items = append(items, item{
					Id:     group.Id.String(),
					TimeID: group.Id.String(),
					Name:   group.Name,
					Active: group.Active,
				})
			} else {
				items = append(items, item{
					Id:         group.Id.String(),
					TimeID:     group.Id.String(),
					Name:       group.Name,
					Active:     group.Active,
					SyncAction: "SET",
				})
			}
		}

		resp := response[item]{
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
