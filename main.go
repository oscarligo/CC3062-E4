package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	_ "modernc.org/sqlite"
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

var db *sql.DB

func main() {
	initDB()
	defer db.Close()

	http.HandleFunc("/api/ping", pingHandler)
	http.HandleFunc("/api/countries", countriesHandler)

	log.Println("JSON API running on :80")
	log.Fatal(http.ListenAndServe(":80", nil))
}

func initDB() {
	var err error
	db, err = sql.Open("sqlite", "./data/countries.db")
	if err != nil {
		log.Fatal("Error opening database:", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS country (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			capital TEXT NOT NULL,
			population INTEGER NOT NULL,
			continent TEXT NOT NULL,
			currency TEXT NOT NULL
		)
	`)
	if err != nil {
		log.Fatal("Error creating table:", err)
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

	case http.MethodDelete:
		handelDeleteCountry(w, r)

	default:
		writeJSON(w, http.StatusMethodNotAllowed, Message{Message: "Method not allowed"})
	}
}

func handleGetCountries(w http.ResponseWriter, r *http.Request) {

	query := r.URL.Query()
	idParam := query.Get("id")

	if idParam == "" {
		rows, err := db.Query("SELECT id, name, capital, population, continent, currency FROM country ORDER BY id")
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, Message{Message: "Database query failed"})
			return
		}
		defer rows.Close()

		countries := make([]Countries, 0)
		for rows.Next() {
			var country Countries
			err = rows.Scan(&country.ID, &country.Name, &country.Capital, &country.Population, &country.Continent, &country.Currency)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, Message{Message: "Error reading database rows"})
				return
			}
			countries = append(countries, country)
		}

		if err = rows.Err(); err != nil {
			writeJSON(w, http.StatusInternalServerError, Message{Message: "Database iteration failed"})
			return
		}

		writeJSON(w, http.StatusOK, countries)
		return
	}

	id, err := strconv.Atoi(idParam)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, Message{Message: "Invalid id parameter"})
		return
	}

	var country Countries
	err = db.QueryRow(
		"SELECT id, name, capital, population, continent, currency FROM country WHERE id = ?",
		id,
	).Scan(&country.ID, &country.Name, &country.Capital, &country.Population, &country.Continent, &country.Currency)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, Message{Message: "Country not found"})
			return
		}

		writeJSON(w, http.StatusInternalServerError, Message{Message: "Database query failed"})
		return
	}

	writeJSON(w, http.StatusOK, country)
}

func handleCreateCountry(w http.ResponseWriter, r *http.Request) {

	var newCountry Countries

	err := json.NewDecoder(r.Body).Decode(&newCountry)

	if err != nil {
		writeJSON(w, http.StatusBadRequest, Message{Message: "Invalid JSON body"})
		return
	}

	if newCountry.Name == "" || newCountry.Capital == "" || newCountry.Population <= 0 || newCountry.Continent == "" || newCountry.Currency == "" {
		writeJSON(w, http.StatusBadRequest, Message{Message: "All fields are required and population must be greater than 0"})
		return
	}

	result, err := db.Exec(
		"INSERT INTO country (name, capital, population, continent, currency) VALUES (?, ?, ?, ?, ?)",
		newCountry.Name,
		newCountry.Capital,
		newCountry.Population,
		newCountry.Continent,
		newCountry.Currency,
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, Message{Message: "Database insert failed"})
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, Message{Message: "Could not read inserted id"})
		return
	}

	newCountry.ID = int(id)

	writeJSON(w, http.StatusCreated, newCountry)
}

func handelDeleteCountry(w http.ResponseWriter, r *http.Request) {
	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		writeJSON(w, http.StatusBadRequest, Message{Message: "ID parameter is required"})
		return
	}

	id, err := strconv.Atoi(idParam)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, Message{Message: "Invalid ID parameter"})
		return
	}

	result, err := db.Exec("DELETE FROM country WHERE id = ?", id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, Message{Message: "Database delete failed"})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, Message{Message: "Could not read affected rows"})
		return
	}

	if rowsAffected == 0 {
		writeJSON(w, http.StatusNotFound, Message{Message: "Country not found"})
		return
	}

	writeJSON(w, http.StatusOK, Message{Message: "Country deleted successfully"})
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	err := json.NewEncoder(w).Encode(payload)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
