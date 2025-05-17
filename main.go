package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"text/template"
)

type WeatherResponse struct {
	TempC  float64 `json:"temp_c"`
	TempF  float64 `json:"temp_f"`
	TempK  float64 `json:"temp_k"`
	Cidade string
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type ViaCEPResponse struct {
	Localidade string `json:"localidade"`
	UF         string `json:"uf"`
	Erro       bool   `json:"erro"` // Indica se houve erro na consulta do CEP.
}

type WeatherAPIResponse struct {
	Current struct {
		TempC float64 `json:"temp_c"`
		TempF float64 `json:"temp_f"`
	} `json:"current"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

const viaCEPURL = "https://viacep.com.br/ws/%s/json/"             // URL da API ViaCEP para buscar dados do CEP.
const weatherAPIURL = "http://api.weatherapi.com/v1/current.json" // URL da API WeatherAPI para buscar dados climáticos.

func main() {

	http.HandleFunc("/", handler)
	fmt.Println("Servidor rodando na porta 7000...")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.ListenAndServe(":"+port, nil)

}

func handler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.New("temperatura.html").ParseFiles("temperatura.html"))

	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	cepParam := r.URL.Query().Get("cep")
	if cepParam == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !validarCEP(cepParam) {
		json.NewEncoder(w).Encode(ErrorResponse{Message: "CEP inválido"})
		return
	}

	localizacao, err := buscarLocalizacao(cepParam)
	if err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Erro ao buscar localização"})
		return
	}
	if localizacao.Erro {
		json.NewEncoder(w).Encode(ErrorResponse{Message: "CEP não encontrado"})
		return
	}
	clima, err := buscarClimaAtual(localizacao.Localidade)
	if err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Erro ao buscar clima atual"})
		return
	}

	tempC := clima.Current.TempC
	tempF := clima.Current.TempF
	tempK := tempC + 273.15

	t.Execute(w, WeatherResponse{
		TempC:  tempC,
		TempF:  tempF,
		TempK:  tempK,
		Cidade: localizacao.Localidade,
	})
}

func validarCEP(cep string) bool {
	re := regexp.MustCompile(`^[0-9]{8}$`)
	if !re.MatchString(cep) {
		return false
	}
	// Implementar a validação do CEP (ex: verificar se tem 8 dígitos)
	if len(cep) != 8 {
		return false
	}
	return true
}

func buscarLocalizacao(cep string) (ViaCEPResponse, error) {
	url := fmt.Sprintf(viaCEPURL, cep)
	resp, err := http.Get(url)
	if err != nil {
		return ViaCEPResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ViaCEPResponse{}, fmt.Errorf("erro ao buscar localização: %v", resp.StatusCode)
	}
	var viaCEP ViaCEPResponse
	err = json.NewDecoder(resp.Body).Decode(&viaCEP)
	return viaCEP, err
}

func buscarClimaAtual(localidade string) (WeatherAPIResponse, error) {
	// apiKey := os.Getenv("WEATHER_KEY")
	var apiKey = "5912bbabd3ed4d60a8323331251505"
	if apiKey == "" {
		return WeatherAPIResponse{}, fmt.Errorf("chave da API não configurada")
	}

	cidade := url.QueryEscape(localidade)
	fmt.Printf("%s", localidade)
	url := fmt.Sprintf("%s?key=%s&q=%s", weatherAPIURL, apiKey, cidade)
	resp, err := http.Get(url)
	if err != nil {
		return WeatherAPIResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return WeatherAPIResponse{}, fmt.Errorf("erro ao buscar clima: %s", resp.Status)
	}

	var climaAPIResp WeatherAPIResponse
	err = json.NewDecoder(resp.Body).Decode(&climaAPIResp)
	return climaAPIResp, err
}
