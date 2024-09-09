package syncronizer

import (
	"net/http"

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
	}

	utils.RespondWithJSON(w, http.StatusOK, config)
}
