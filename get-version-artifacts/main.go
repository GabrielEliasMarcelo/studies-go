package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

// Package representa um pacote retornado pela API do Azure DevOps
type Package struct {
	Name     string   `json:"name"`
	Versions []struct {
		Version string `json:"version"`
	} `json:"versions"`
}

// Response representa a resposta da API do Azure DevOps
type Response struct {
	Value []Package `json:"value"`
}

func main() {
	// Verifica se todos os argumentos foram passados
	if len(os.Args) < 5 {
		fmt.Println("Uso: go run script.go <user> <token> <org> <feed>")
		os.Exit(1)
	}

	// Parâmetros recebidos da linha de comando
	user := os.Args[1]
	token := os.Args[2]
	org := os.Args[3]
	feed := os.Args[4]

	// URL da API do Azure DevOps
	url := fmt.Sprintf("https://feeds.dev.azure.com/%s/_apis/packaging/feeds/%s/packages?api-version=6.0-preview.1", org, feed)

	// Criando a requisição HTTP
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Erro ao criar requisição:", err)
		os.Exit(1)
	}

	// Autenticação básica
	req.SetBasicAuth(user, token)
	req.Header.Set("Content-Type", "application/json")

	// Fazendo a requisição
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Erro ao realizar requisição:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Lendo a resposta
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Erro ao ler resposta:", err)
		os.Exit(1)
	}

	// Decodificando JSON
	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println("Erro ao decodificar JSON:", err)
		os.Exit(1)
	}

	// Criando um mapa para armazenar os pacotes e suas últimas versões
	latestVersions := make(map[string]string)
	for _, pkg := range response.Value {
		if len(pkg.Versions) > 0 {
			latestVersions[pkg.Name] = pkg.Versions[0].Version
		}
	}

	// Convertendo para JSON
	outputJSON, err := json.MarshalIndent(latestVersions, "", "  ")
	if err != nil {
		fmt.Println("Erro ao gerar JSON:", err)
		os.Exit(1)
	}

	// Salvando no arquivo
	outputFile := "latest_versions.json"
	err = ioutil.WriteFile(outputFile, outputJSON, 0644)
	if err != nil {
		fmt.Println("Erro ao salvar arquivo:", err)
		os.Exit(1)
	}

	fmt.Println("Últimas versões dos pacotes salvas em", outputFile)
}
