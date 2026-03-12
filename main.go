package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Team struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Countries struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Capital    string `json:"capital"`
	Population int    `json:"population"`
	Continent  string `json:"continent"`
	Currency   string `json:"currency"`
}

type Message struct {
	Message string `json:"message"`
}

var teams []Team
var countries []Countries

func main() {
	loadCountries()

	http.HandleFunc("/api/ping", pingHandler)
	http.HandleFunc("/api/countries", countriesHandler)

	log.Println("POST JSON API running on :80")
	log.Fatal(http.ListenAndServe(":80", nil))
}

func loadCountries() {
	file, err := os.ReadFile("./data/countries.json")
	if err != nil {
		log.Fatal("Error reading file:", err)
	}

	err = json.Unmarshal(file, &countries)
	if err != nil {
		log.Fatal("Error parsing JSON:", err)
	}
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	response := Message{
		Message: "pong",
	}

	writeJSON(w, http.StatusOK, response)
}

func countriesHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {

	case http.MethodGet:
		handleGetCountries(w, r)

	case http.MethodPost:
		handleCreateCountry(w, r)

	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func handleGetCountries(w http.ResponseWriter, r *http.Request) {

	query := r.URL.Query()
	idParam := query.Get("id")

	if idParam == "" {
		writeJSON(w, http.StatusOK, countries)
		return
	}

	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "Invalid id parameter", http.StatusBadRequest)
		return
	}

	for _, country := range countries {
		if country.ID == id {
			writeJSON(w, http.StatusOK, country)
			return
		}
	}

	http.Error(w, "Country not found", http.StatusNotFound)
}

func handleCreateCountry(w http.ResponseWriter, r *http.Request) {

	var newCountry Countries

	err := json.NewDecoder(r.Body).Decode(&newCountry)

	if err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if newCountry.Name == "" || newCountry.Capital == "" || newCountry.Population <= 0 || newCountry.Continent == "" || newCountry.Currency == "" {
		http.Error(w, "All fields are required and population must be greater than 0", http.StatusBadRequest)
		return
	}

	newCountry.ID = generateNextID()

	countries = append(countries, newCountry)
	saveCountries()

	writeJSON(w, http.StatusCreated, newCountry)
}

func generateNextID() int {
	maxID := 0

	for _, country := range countries {
		if country.ID > maxID {
			maxID = country.ID
		}
	}

	return maxID + 1
}

func saveCountries() {
	data, err := json.MarshalIndent(countries, "", "  ")
	if err != nil {
		log.Println("Error marshaling JSON:", err)
		return
	}

	err = os.WriteFile("./data/countries.json", data, 0644)
	if err != nil {
		log.Println("Error writing file:", err)
	}
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	err := json.NewEncoder(w).Encode(payload)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
