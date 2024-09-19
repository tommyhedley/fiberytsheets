package handlers

import (
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
	Automations         bool `json:"automations"`
}

type Action struct {
	Action      string `json:"action"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Args        []Arg  `json:"args"`
}

type Arg struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	Type         string `json:"type"`
	TextTemplate bool   `json:"textTemplateSupported,omitempty"`
}

type AppConfig struct {
	ID             string           `json:"id"`
	Name           string           `json:"name"`
	Version        string           `json:"version"`
	Description    string           `json:"description"`
	Authentication []Authentication `json:"authentication"`
	Sources        []string         `json:"sources"`
	ResponsibleFor ResponsibleFor   `json:"responsibleFor"`
	Actions        []Action         `json:"actions"`
}

type Oauth2Fields struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Type        string `json:"type"`
	ID          string `json:"id"`
}

func Config(w http.ResponseWriter, r *http.Request) {
	oauth2 := Oauth2Fields{
		Title:       "callback_uri",
		Description: "OAuth post-auth redirect URI",
		Type:        "oauth",
		ID:          "callback_uri",
	}

	config := AppConfig{
		ID:          "qbtime",
		Name:        "Quickbooks Time",
		Version:     "0.1.6",
		Description: "Integrate Quickbooks Time data with Fibery",
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
			Automations:         true,
		},
		Actions: []Action{
			{
				Action:      "createUser",
				Name:        "Create User",
				Description: "Create a new Quickbooks Time user",
				Args: []Arg{
					{
						ID:           "name",
						Name:         "Name",
						Type:         "text",
						Description:  "Full Name",
						TextTemplate: true,
					},
					{
						ID:           "email",
						Name:         "Email",
						Type:         "text",
						Description:  "Email",
						TextTemplate: true,
					},
					{
						ID:           "groupID",
						Name:         "Group ID",
						Type:         "text",
						Description:  "Group ID",
						TextTemplate: true,
					},
				},
			},
		},
	}

	utils.RespondWithJSON(w, http.StatusOK, config)
}
