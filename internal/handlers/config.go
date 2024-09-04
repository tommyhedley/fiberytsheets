package handlers

import (
	"encoding/json"
	"net/http"
)

type AppConfig struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Version        string `json:"version"`
	Description    string `json:"description"`
	Authentication []struct {
		ID          string   `json:"id"`
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Fields      []string `json:"fields"`
	} `json:"authentication"`
	Sources        []string `json:"sources"`
	ResponsibleFor struct {
		DataSynchronization bool `json:"dataSynchronization"`
	} `json:"responsibleFor"`
}

func (cfg *AppConfig) Config(w http.ResponseWriter, r *http.Request) {
	config, err := json.MarshalIndent(cfg, "", "	")
	if err != nil {
		http.Error(w, "Error marshalling config", http.StatusInternalServerError)
		return
	}

	// Set the content type to application/json.
	w.Header().Set("Content-Type", "application/json")

	// Write the JSON data to the response.
	w.Write(config)
}
