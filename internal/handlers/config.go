package handlers

import (
	"encoding/json"
	"net/http"
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
		Type:        "oauth2",
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
				Description: "OAuth v2-based authentication and authorization for access to TSheets",
				Fields:      []interface{}{oauth2},
			},
		},
		Sources: []string{},
		ResponsibleFor: ResponsibleFor{
			DataSynchronization: true,
		},
	}

	res, err := json.MarshalIndent(config, "", "	")
	if err != nil {
		http.Error(w, "Error marshalling config", http.StatusInternalServerError)
		return
	}

	// Set the content type to application/json.
	w.Header().Set("Content-Type", "application/json")

	// Write the JSON data to the response.
	w.Write(res)
}
