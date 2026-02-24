package api

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// Brasil API (dados públicos): https://brasilapi.com.br/api/cnpj/v1/{cnpj}
const brasilAPICNPJ = "https://brasilapi.com.br/api/cnpj/v1/"

var onlyDigits = regexp.MustCompile(`[^0-9]`)

type brasilAPIResponse struct {
	RazaoSocial      string `json:"razao_social"`
	NomeFantasia     string `json:"nome_fantasia"`
	Logradouro       string `json:"logradouro"`
	Numero           string `json:"numero"`
	Complemento      string `json:"complemento"`
	Bairro           string `json:"bairro"`
	Municipio        string `json:"municipio"`
	UF               string `json:"uf"`
	CEP              string `json:"cep"`
	CNAEFiscal       string `json:"cnae_fiscal"`
	DescricaoSituacao string `json:"descricao_situacao_cadastral"`
}

// CNPJLookupResponse resposta padronizada para o frontend
type CNPJLookupResponse struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

// CNPJLookupHandler consulta dados do CNPJ na Brasil API e retorna nome e endereço para pré-preenchimento.
// Uso: GET /api/cnpj?q=12345678000199 (apenas dígitos, 14 caracteres).
// Fonte: dados públicos (Brasil API / Receita Federal). Uso permitido para preenchimento de cadastros.
func CNPJLookupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	cnpj := onlyDigits.ReplaceAllString(q, "")
	if len(cnpj) != 14 {
		http.Error(w, "CNPJ deve conter 14 dígitos", http.StatusBadRequest)
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodGet, brasilAPICNPJ+cnpj, nil)
	if err != nil {
		http.Error(w, "Erro interno", http.StatusInternalServerError)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Não foi possível consultar o CNPJ. Tente novamente.", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		http.Error(w, "CNPJ não encontrado", http.StatusNotFound)
		return
	}
	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Serviço de consulta indisponível. Tente mais tarde.", http.StatusBadGateway)
		return
	}

	var data brasilAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		http.Error(w, "Resposta inválida da consulta", http.StatusInternalServerError)
		return
	}

	// Nome: razão social ou nome fantasia
	name := strings.TrimSpace(data.RazaoSocial)
	if name == "" {
		name = strings.TrimSpace(data.NomeFantasia)
	}
	if name == "" {
		name = "Empresa"
	}

	// Endereço: logradouro, número, complemento, bairro, cidade - UF, CEP
	parts := []string{}
	if data.Logradouro != "" {
		parts = append(parts, strings.TrimSpace(data.Logradouro))
		if data.Numero != "" {
			parts = append(parts, strings.TrimSpace(data.Numero))
		}
		if data.Complemento != "" {
			parts = append(parts, strings.TrimSpace(data.Complemento))
		}
	}
	if data.Bairro != "" {
		parts = append(parts, strings.TrimSpace(data.Bairro))
	}
	if data.Municipio != "" || data.UF != "" {
		cityUF := strings.TrimSpace(data.Municipio)
		if data.UF != "" {
			if cityUF != "" {
				cityUF += " - "
			}
			cityUF += strings.TrimSpace(data.UF)
		}
		parts = append(parts, cityUF)
	}
	if data.CEP != "" {
		parts = append(parts, "CEP "+strings.TrimSpace(data.CEP))
	}
	address := strings.Join(parts, ", ")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CNPJLookupResponse{Name: name, Address: address})
}
