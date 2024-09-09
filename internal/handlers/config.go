package handlers

import (
	"fmt"
	"net/http"

	"github.com/tommyhedley/fiberytsheets/internal/utils"
)

type Authentication struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Fields      []interface{} `json:"fields"`
}

type ResponsibleFor struct {
	DataSynchronization bool `json:"dataSynchronization"`
}

type AppConfig struct {
	ID             string           `json:"id"`
	Name           string           `json:"name"`
	Version        string           `json:"version"`
	Description    string           `json:"description"`
	Authentication []Authentication `json:"authentication"`
	Sources        []string         `json:"sources"`
	ResponsibleFor ResponsibleFor   `json:"responsibleFor"`
}

type Oauth2Fields struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Id          string `json:"id"`
}

func Config(w http.ResponseWriter, r *http.Request) {
	oauth2 := Oauth2Fields{
		Title:       "callback_uri",
		Description: "OAuth post-auth redirect URI",
		Type:        "oauth",
		Id:          "callback_uri",
	}

	config := AppConfig{
		ID:          "qbtime",
		Name:        "Quickbooks Time",
		Version:     "0.1.0",
		Description: "Integrate Quickbooks Time with Fibery",
		Authentication: []Authentication{
			{
				ID:          "oauth2",
				Name:        "OAuth v2 Authentication",
				Description: "OAuth v2-based authentication and authorization for access to Quickbooks Time",
				Fields:      []interface{}{oauth2},
			},
		},
		Sources: []string{},
		ResponsibleFor: ResponsibleFor{
			DataSynchronization: true,
		},
	}

	utils.RespondWithJSON(w, http.StatusOK, config)
	fmt.Println("Config Returned")
}
