package syncronizer

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/tommyhedley/fiberytsheets/internal/utils"
)

type SyncConfig struct {
	Types   []SyncType   `json:"types"`
	Filters []SyncFilter `json:"filters,omitempty"`
}
type SyncType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type SyncFilter struct {
	Id       string `json:"id"`
	Title    string `json:"title"`
	Type     string `json:"type"`
	Datalist bool   `json:"datalist,omitempty"`
	Optional bool   `json:"optional,omitempty"`
	Secured  bool   `json:"secured,omitempty"`
}

func Config(w http.ResponseWriter, r *http.Request) {
	config := SyncConfig{
		Types: []SyncType{
			{
				ID:   "user",
				Name: "User",
			},
			{
				ID:   "group",
				Name: "Group",
			},
		},
		Filters: []SyncFilter{
			{
				Id:       "inactiveUsers",
				Title:    "Include inactive users?",
				Type:     "bool",
				Optional: true,
			},
			{
				Id:       "timesheetStart",
				Title:    "Starting date of timesheet sync after Jan 1, 2020. Jan 1, 2020 will be used if selection is earlier or empty.",
				Type:     "datebox",
				Optional: true,
			},
		},
	}

	utils.RespondWithJSON(w, http.StatusOK, config)
}

func ValidateFilters(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Types   []string       `json:"types"`
		Filter  map[string]any `json:"filter"`
		Account struct {
			Token string `json:"token"`
		} `json:"account"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	if params.Filter["timesheetStart"] != nil {
		if datesString, ok := params.Filter["timesheetStart"].(string); ok {
			_, err := time.Parse(time.RFC3339, datesString)
			if err != nil {
				utils.RespondWithError(w, http.StatusBadRequest, "Couldn't parse date string")
				return
			}
			utils.RespondWithJSON(w, http.StatusOK, nil)
			return
		}
		utils.RespondWithError(w, http.StatusBadRequest, "Filter input value is an invalid type")
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, nil)
}
