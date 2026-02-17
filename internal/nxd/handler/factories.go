package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"hubsystem/internal/nxd/middleware"
	"hubsystem/internal/nxd/store"

	"github.com/google/uuid"
)

// ListFactories returns factories the current user is a member of. Requires auth.
func ListFactories(w http.ResponseWriter, r *http.Request) {
	db := store.NXDDB()
	if db == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "NXD não configurado"})
		return
	}
	userID := middleware.UserIDFromRequest(r)
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "faça login"})
		return
	}
	list, err := store.ListFactoriesForUser(db, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "erro ao listar fábricas"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"factories": list})
}

// CreateFactoryRequest is the JSON body for POST /nxd/factories.
type CreateFactoryRequest struct {
	Name     string `json:"name"`
	CNPJ     string `json:"cnpj"`
	Location string `json:"location"`
}

// CreateFactory creates a factory and adds the current user as OWNER. Requires auth.
func CreateFactory(w http.ResponseWriter, r *http.Request) {
	db := store.NXDDB()
	if db == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "NXD não configurado"})
		return
	}
	userID := middleware.UserIDFromRequest(r)
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "faça login"})
		return
	}
	var req CreateFactoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "payload inválido"})
		return
	}
	if req.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "name obrigatório"})
		return
	}
	id, err := store.CreateFactory(db, req.Name, req.CNPJ, req.Location, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "erro ao criar fábrica"})
		return
	}
	f, _ := store.GetFactoryByID(db, id)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"factory": f})
}

// GenerateGatewayKey generates a key for the factory, stores only its hash, returns the key once. Requires auth + factory member.
func GenerateGatewayKey(w http.ResponseWriter, r *http.Request) {
	db := store.NXDDB()
	if db == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "NXD não configurado"})
		return
	}
	factoryID, ok := middleware.FactoryIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "factory_id obrigatório"})
		return
	}
	vars := mux.Vars(r)
	idStr := vars["id"]
	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "id obrigatório"})
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil || id != factoryID {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "fábrica não autorizada"})
		return
	}
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "erro ao gerar chave"})
		return
	}
	key := hex.EncodeToString(b)
	hash := sha256Sum(key)
	if err := store.SetFactoryGatewayKeyHash(db, id, hash); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "erro ao salvar chave"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"gateway_key": key,
		"warning":     "Guarde esta chave. Ela não será exibida novamente.",
	})
}

func sha256Sum(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
