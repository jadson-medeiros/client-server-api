package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Exchange struct {
	Bid string `json:"bid"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(),  30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/quote", nil)
	if err != nil {
		fmt.Printf("Error with request: %s", err)
		return 
	}
	
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error with response: %s", err)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		fmt.Printf("Server returned non-200 status code: %d\n", res.StatusCode)
		return
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %s\n", err)
		return
	}

	var exchange Exchange
	err = json.Unmarshal(body, &exchange)
	if err != nil {
		fmt.Printf("Error decoding JSON: %s\n", err)
		return
	}

	err = saveExchange(exchange.Bid)
	if err != nil {
		fmt.Printf("Error saving exchange rate: %s\n", err)
		return
	}

	fmt.Println("Exchange rate saved successfully.")
}

func saveExchange(exchangeRate string) error {
	file, err := os.OpenFile("exchange.txt", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("error opening file: %s", err)
	}
	defer file.Close()

	_, err = fmt.Fprintf(file, "Dólar: %s\n", exchangeRate)
	if err != nil {
		fmt.Println("Error writing to file:", err)
	}

	return nil
}