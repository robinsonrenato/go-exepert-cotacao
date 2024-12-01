package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type ExchangeRate struct {
	USDBRL struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

func fetchDollarRate(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return "", err
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var rate ExchangeRate
	if err := json.NewDecoder(resp.Body).Decode(&rate); err != nil {
		return "", err
	}

	return rate.USDBRL.Bid, nil
}

func saveToDB(ctx context.Context, db *sql.DB, value string) error {
	query := "INSERT INTO exchange_rates (rate, timestamp) VALUES (?, ?)"
	stmt, err := db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, value, time.Now())
	return err
}

func main() {
	// Banco de dados
	db, err := sql.Open("sqlite3", "./exchange_rates.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS exchange_rates (id INTEGER PRIMARY KEY, rate TEXT, timestamp DATETIME)")
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		ctxAPI, cancelAPI := context.WithTimeout(r.Context(), 200*time.Millisecond)
		defer cancelAPI()

		rate, err := fetchDollarRate(ctxAPI)
		if err != nil {
			http.Error(w, "Erro ao obter a cotação", http.StatusInternalServerError)
			log.Println("Erro na API:", err)
			return
		}

		ctxDB, cancelDB := context.WithTimeout(r.Context(), 10*time.Millisecond)
		defer cancelDB()

		if err := saveToDB(ctxDB, db, rate); err != nil {
			http.Error(w, "Erro ao salvar no banco", http.StatusInternalServerError)
			log.Println("Erro no banco:", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"bid": rate})
	})

	log.Println("Servidor rodando na porta 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
