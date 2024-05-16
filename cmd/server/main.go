package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Exchange struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
}

func main() {
	db, err := sql.Open("sqlite3", "../../data/exchange.db")
	if err != nil {
		log.Fatalf("Error with db connection: %s", err)
	}
	defer db.Close()

	err = executeSQLFile(db, "../../data/db.sql")
	if err != nil {
		log.Fatalf("Error executing SQL file: %s", err)
	}

	http.HandleFunc("/quote", func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, db)
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handler(w http.ResponseWriter, r *http.Request, db *sql.DB) { 
	ctx := r.Context()
	log.Println("Request started.")
	defer log.Println("Request stopped.")

	select {
		case <- time.After(5 * time.Second):
			apiCtx, cancelAPI := context.WithTimeout(ctx, 200*time.Millisecond)
			defer cancelAPI()

			exchange, err := getExchange(apiCtx)
			if err != nil {
				log.Printf("Error with request: %s", err)
				http.Error(w, "Error with request.\n", http.StatusInternalServerError)
				return
			}

			dbCtx, cancelDB := context.WithTimeout(ctx, 10*time.Millisecond)
			defer cancelDB()
			
			err = insertExchange(dbCtx, db, exchange)

			if err != nil {
				log.Fatalf("Error with saving data: %s", err)
				http.Error(w, "Error with saving data.\n", http.StatusInternalServerError)
				return
			}

			log.Println("Request processed with success.\n")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(exchange)

		case <-ctx.Done():
			log.Println("Request cancelled by the client.\n")
			http.Error(w, "Request cancelled by the client.\n", http.StatusRequestTimeout)
	}
}

func getExchange(ctx context.Context) (*Exchange, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var data map[string]Exchange
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	exchange, exists := data["USDBRL"]
	if !exists {
		return nil, fmt.Errorf("exchange data not found")
	}

	return &exchange, nil
}

func insertExchange(ctx context.Context, db *sql.DB, exchange *Exchange) error {
	stmt, err := db.PrepareContext(ctx, "INSERT INTO exchange_rates(code, codein, name, high, low, var_bid, pct_change, bid, ask, timestamp, create_date) VALUES(?,?,?,?,?,?,?,?,?,?,?);")
	if err != nil {
	 return err
	}
 
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, exchange.Code, exchange.Codein, exchange.Name, exchange.High, exchange.Low, exchange.VarBid, exchange.PctChange, exchange.Bid, exchange.Ask, exchange.Timestamp, exchange.CreateDate)
	if err != nil {
		return err
	}

	return nil
 }

 func executeSQLFile(db *sql.DB, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	sqlBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	sql := string(sqlBytes)
	_, err = db.Exec(sql)
	return err
}