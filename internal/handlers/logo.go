package handlers

import (
	"io"
	"net/http"
	"os"
)

func Logo(w http.ResponseWriter, r *http.Request) {
	// Open the SVG file
	file, err := os.Open("logo.svg")
	if err != nil {
		http.Error(w, "Unable to open SVG file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Read the file contents
	svgData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Unable to read SVG file", http.StatusInternalServerError)
		return
	}

	// Set response header to image/svg+xml
	w.Header().Set("Content-Type", "image/svg+xml")

	// Write the SVG data to the response
	w.Write(svgData)
}
