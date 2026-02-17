package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/google/uuid"

	"hubsystem/internal/nxd/middleware"
	"hubsystem/internal/nxd/store"
)

func ListGroups(w http.ResponseWriter, r *http.Request) {
	db := store.NXDDB()
	if db == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "NXD nao configurado"})
		return
	}
	factoryID, ok := middleware.FactoryIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "factory_id obrigatorio"})
		return
	}
	parentIDStr := r.URL.Query().Get("parent_id")
	var parentID *uuid.UUID
	if parentIDStr != "" {
		u, err := uuid.Parse(parentIDStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "parent_id invalido"})
			return
		}
		parentID = &u
	}
	list, err := store.ListGroups(db, factoryID, parentID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "erro ao listar setores"})
		return
	}
	// Converter para resposta com metadata
	var responseList []map[string]interface{}
	for _, g := range list {
		m := map[string]interface{}{
			"id":         g.ID,
			"factory_id": g.FactoryID,
			"parent_id":  g.ParentID,
			"name":       g.Name,
			"path":       g.Path,
			"metadata":   g.MetadataMap(),
			"created_at": g.CreatedAt,
		}
		responseList = append(responseList, m)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"groups": responseList})
}

type CreateGroupRequest struct {
	ParentID *string                `json:"parent_id"`
	Name     string                 `json:"name"`
	Metadata map[string]interface{} `json:"metadata"`
}

func CreateGroup(w http.ResponseWriter, r *http.Request) {
	db := store.NXDDB()
	if db == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "NXD nao configurado"})
		return
	}
	factoryID, ok := middleware.FactoryIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "factory_id obrigatorio"})
		return
	}
	var req CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "payload invalido"})
		return
	}
	if req.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "name obrigatorio"})
		return
	}
	var parentID *uuid.UUID
	if req.ParentID != nil && *req.ParentID != "" {
		u, err := uuid.Parse(*req.ParentID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "parent_id invalido"})
			return
		}
		parentID = &u
	}
	id, err := store.CreateGroup(db, factoryID, parentID, req.Name, req.Metadata)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "erro ao criar setor"})
		return
	}
	g, _ := store.GetGroupByID(db, id, factoryID)
	resp := map[string]interface{}{
		"id":         g.ID,
		"factory_id": g.FactoryID,
		"parent_id":  g.ParentID,
		"name":       g.Name,
		"path":       g.Path,
		"metadata":   g.MetadataMap(),
		"created_at": g.CreatedAt,
	}
	userID := middleware.UserIDFromRequest(r)
	_ = store.AuditWrite(db, userID, &factoryID, "GROUP_CREATE", "GROUP", id.String(), "", nil)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"group": resp})
}

type UpdateGroupRequest struct {
	Name     string                 `json:"name"`
	Metadata map[string]interface{} `json:"metadata"`
}

func UpdateGroup(w http.ResponseWriter, r *http.Request) {
	db := store.NXDDB()
	if db == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "NXD nao configurado"})
		return
	}
	factoryID, ok := middleware.FactoryIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "factory_id obrigatorio"})
		return
	}
	vars := mux.Vars(r)
	idStr := vars["id"]
	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "id obrigatorio"})
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "id invalido"})
		return
	}
	var req UpdateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "payload invalido"})
		return
	}
	if req.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "name obrigatorio"})
		return
	}
	if err := store.UpdateGroup(db, id, factoryID, req.Name, req.Metadata); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "erro ao atualizar setor"})
		return
	}
	g, _ := store.GetGroupByID(db, id, factoryID)
	resp := map[string]interface{}{
		"id":         g.ID,
		"factory_id": g.FactoryID,
		"parent_id":  g.ParentID,
		"name":       g.Name,
		"path":       g.Path,
		"metadata":   g.MetadataMap(),
		"created_at": g.CreatedAt,
	}
	userID := middleware.UserIDFromRequest(r)
	_ = store.AuditWrite(db, userID, &factoryID, "RENAME_GROUP", "GROUP", idStr, "", nil)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"group": resp})
}
