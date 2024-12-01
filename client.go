package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type Response struct {
	Bid string `json:"bid"`
}

func getDollarRate(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/cotacao", nil)
	if err != nil {
		return "", err
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var res Response
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}

	return res.Bid, nil
}

func saveToFile(value string) error {
	content := fmt.Sprintf("Dólar: %s", value)
	return ioutil.WriteFile("cotacao.txt", []byte(content), 0644)
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	rate, err := getDollarRate(ctx)
	if err != nil {
		log.Fatal("Erro ao obter cotação do servidor:", err)
	}

	if err := saveToFile(rate); err != nil {
		log.Fatal("Erro ao salvar no arquivo:", err)
	}

	fmt.Println("Cotação salva com sucesso em 'cotacao.txt'")
}
