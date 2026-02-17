package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"hubsystem/internal/nxd/middleware"
	"hubsystem/internal/nxd/store"

	"github.com/google/uuid"
)

// ListAssets returns assets. Query: factory_id (required), ungrouped=true|false, group_id=uuid, search=.
func ListAssets(w http.ResponseWriter, r *http.Request) {
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
	groupIDStr := r.URL.Query().Get("group_id")
	var list []store.AssetRow
	var err error
	if groupIDStr != "" {
		groupID, parseErr := uuid.Parse(groupIDStr)
		if parseErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "group_id inválido"})
			return
		}
		list, err = store.ListAssetsByGroup(db, factoryID, groupID)
	} else {
		ungroupedOnly := r.URL.Query().Get("ungrouped") == "true"
		search := r.URL.Query().Get("search")
		list, err = store.ListAssets(db, factoryID, ungroupedOnly, search)
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "erro ao listar ativos"})
		return
	}
	// Marshal-friendly: annotations as map
	out := make([]map[string]interface{}, 0, len(list))
	for _, a := range list {
		out = append(out, assetToMap(a))
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"assets": out})
}

func assetToMap(a store.AssetRow) map[string]interface{} {
	m := map[string]interface{}{
		"id":            a.ID.String(),
		"factory_id":    a.FactoryID.String(),
		"source_tag_id": a.SourceTagID,
		"display_name":  a.DisplayName,
		"description":   a.Description,
		"created_at":    a.CreatedAt,
	}
	if a.GroupID != nil {
		m["group_id"] = a.GroupID.String()
	}
	if a.AnnotationsMap() != nil {
		m["annotations"] = a.AnnotationsMap()
	}
	return m
}

// CreateAssetRequest is the body for POST /nxd/assets.
type CreateAssetRequest struct {
	SourceTagID string                 `json:"source_tag_id"`
	DisplayName string                 `json:"display_name"`
	Description string                 `json:"description"`
	GroupID     *string                `json:"group_id"`
	Annotations map[string]interface{} `json:"annotations"`
}

// CreateAsset creates or updates an asset (upsert by factory_id + source_tag_id).
func CreateAsset(w http.ResponseWriter, r *http.Request) {
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
	var req CreateAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "payload inválido"})
		return
	}
	if req.SourceTagID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "source_tag_id obrigatório"})
		return
	}
	var groupID *uuid.UUID
	if req.GroupID != nil && *req.GroupID != "" {
		u, err := uuid.Parse(*req.GroupID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "group_id inválido"})
			return
		}
		groupID = &u
	}
	id, err := store.CreateAsset(db, factoryID, groupID, req.SourceTagID, req.DisplayName, req.Description, req.Annotations)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "erro ao criar ativo"})
		return
	}
	asset, _ := store.GetAssetByID(db, id, factoryID)
	userID := middleware.UserIDFromRequest(r)
	_ = store.AuditWrite(db, userID, &factoryID, "CREATE_ASSET", "ASSET", id.String(), "", nil)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"asset": assetToMap(*asset)})
}

// UpdateAssetRequest is the body for PATCH /nxd/assets/:id.
type UpdateAssetRequest struct {
	DisplayName string                 `json:"display_name"`
	Description string                 `json:"description"`
	Annotations map[string]interface{} `json:"annotations"`
}

// UpdateAsset updates display_name, description, annotations.
func UpdateAsset(w http.ResponseWriter, r *http.Request) {
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
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "id inválido"})
		return
	}
	var req UpdateAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "payload inválido"})
		return
	}
	if err := store.UpdateAsset(db, id, factoryID, req.DisplayName, req.Description, req.Annotations); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "erro ao atualizar ativo"})
		return
	}
	asset, _ := store.GetAssetByID(db, id, factoryID)
	userID := middleware.UserIDFromRequest(r)
	_ = store.AuditWrite(db, userID, &factoryID, "RENAME_ASSET", "ASSET", idStr, "", nil)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"asset": assetToMap(*asset)})
}

// MoveAssetRequest is the body for POST /nxd/assets/:id/move.
type MoveAssetRequest struct {
	GroupID *string `json:"group_id"`
}

// MoveAsset sets asset group_id (null = ungroup). Requires factory_id in query.
func MoveAsset(w http.ResponseWriter, r *http.Request) {
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
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "id inválido"})
		return
	}
	var req MoveAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "payload inválido"})
		return
	}
	var groupID *uuid.UUID
	if req.GroupID != nil && *req.GroupID != "" {
		u, err := uuid.Parse(*req.GroupID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "group_id inválido"})
			return
		}
		groupID = &u
	}
	if err := store.MoveAsset(db, id, factoryID, groupID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "erro ao mover ativo"})
		return
	}
	asset, _ := store.GetAssetByID(db, id, factoryID)
	userID := middleware.UserIDFromRequest(r)
	_ = store.AuditWrite(db, userID, &factoryID, "MOVE_ASSET", "ASSET", idStr, "", nil)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"asset": assetToMap(*asset)})
}
