package store

// assets_test.go — Unit tests for asset store logic.
//
// NOTE: Tests that require a live DB connection are skipped unless the
// TEST_DATABASE_URL environment variable is set. This makes them safe to run
// in CI without a PostgreSQL instance.
//
// To run with a real DB:
//   TEST_DATABASE_URL="postgres://..." go test ./internal/nxd/store/ -v -run TestCreateAsset

import (
	"database/sql"
	"encoding/json"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/google/uuid"
)

// openTestDB opens a PostgreSQL connection from TEST_DATABASE_URL.
// Returns nil if the env var is not set (skip DB tests in pure unit mode).
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set — skipping DB integration test")
		return nil
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("db.Ping: %v", err)
	}
	return db
}

// TestCreateAsset_OnConflict_PreservesGroupID is the regression test for the
// critical ON CONFLICT bug fixed in Deploy 00036-dlw.
//
// Scenario:
//  1. Create an asset manually (with an explicit group_id simulating a user
//     assigning it to a sector via the UI).
//  2. Call CreateAsset again with groupID=nil (simulating automatic DX ingest).
//  3. Assert that the group_id in the DB is unchanged after the upsert.
//
// Before the fix: group_id was reset to NULL by the ON CONFLICT UPDATE clause.
// After the fix:  group_id is preserved on conflict; only display_name /
//                 description / annotations are updated.
func TestCreateAsset_OnConflict_PreservesGroupID(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	// ── Setup: create a temporary factory + sector ─────────────────────────
	factoryID := uuid.New()
	sectorID := uuid.New()

	_, err := db.Exec(`INSERT INTO nxd.factories (id, name) VALUES ($1, $2)
		ON CONFLICT (id) DO NOTHING`, factoryID, "TestFactory-"+factoryID.String())
	if err != nil {
		t.Fatalf("insert factory: %v", err)
	}
	_, err = db.Exec(`INSERT INTO nxd.sectors (id, factory_id, name) VALUES ($1, $2, $3)
		ON CONFLICT (id) DO NOTHING`, sectorID, factoryID, "TestSector-"+sectorID.String())
	if err != nil {
		t.Fatalf("insert sector: %v", err)
	}

	// Cleanup after test.
	t.Cleanup(func() {
		db.Exec(`DELETE FROM nxd.assets WHERE factory_id = $1`, factoryID)
		db.Exec(`DELETE FROM nxd.sectors WHERE id = $1`, sectorID)
		db.Exec(`DELETE FROM nxd.factories WHERE id = $1`, factoryID)
	})

	// ── Step 1: create asset with explicit group_id (user action) ──────────
	const tagID = "SENSOR-001"
	assetID, err := CreateAsset(db, factoryID, &sectorID, tagID, "Sensor 001", "test", nil)
	if err != nil {
		t.Fatalf("initial CreateAsset: %v", err)
	}
	if assetID == uuid.Nil {
		t.Fatal("expected non-nil assetID")
	}

	// Verify group_id was set.
	var gotGroupID sql.NullString
	db.QueryRow(`SELECT group_id FROM nxd.assets WHERE id = $1`, assetID).Scan(&gotGroupID)
	if !gotGroupID.Valid || gotGroupID.String != sectorID.String() {
		t.Fatalf("after initial insert: expected group_id=%s, got %v", sectorID, gotGroupID)
	}

	// ── Step 2: simulate automatic DX ingest (groupID=nil) ────────────────
	// The ingest path always calls CreateAsset with groupID=nil.
	upsertedID, err := CreateAsset(db, factoryID, nil, tagID, "Sensor 001 Updated", "updated description", map[string]interface{}{"source": "dx"})
	if err != nil {
		t.Fatalf("upsert CreateAsset: %v", err)
	}
	if upsertedID != assetID {
		t.Fatalf("ON CONFLICT should return same id: got %s, want %s", upsertedID, assetID)
	}

	// ── Step 3: verify group_id is preserved ──────────────────────────────
	var afterGroupID sql.NullString
	var afterDisplayName string
	var afterAnnotations []byte
	db.QueryRow(
		`SELECT group_id, display_name, COALESCE(annotations::text,'{}') FROM nxd.assets WHERE id = $1`,
		assetID,
	).Scan(&afterGroupID, &afterDisplayName, &afterAnnotations)

	// group_id MUST be unchanged.
	if !afterGroupID.Valid {
		t.Fatalf("REGRESSION: group_id was reset to NULL by ON CONFLICT upsert — fix failed!")
	}
	if afterGroupID.String != sectorID.String() {
		t.Fatalf("REGRESSION: group_id changed from %s to %s", sectorID, afterGroupID.String)
	}
	t.Logf("✓ group_id preserved: %s", afterGroupID.String)

	// display_name SHOULD be updated (metadata refresh is allowed).
	if afterDisplayName != "Sensor 001 Updated" {
		t.Errorf("display_name not updated: got %q, want %q", afterDisplayName, "Sensor 001 Updated")
	}

	// annotations SHOULD be updated.
	var ann map[string]interface{}
	if err := json.Unmarshal(afterAnnotations, &ann); err == nil {
		if ann["source"] != "dx" {
			t.Errorf("annotations not updated: source=%v", ann["source"])
		}
	}

	t.Logf("✓ TestCreateAsset_OnConflict_PreservesGroupID passed")
}

// TestCreateAsset_NilGroupID_DoesNotPanic tests that creating an asset with
// nil groupID (the ingest path) works without panics.
func TestCreateAsset_NilGroupID_DoesNotPanic(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	factoryID := uuid.New()
	_, err := db.Exec(`INSERT INTO nxd.factories (id, name) VALUES ($1, $2)
		ON CONFLICT (id) DO NOTHING`, factoryID, "TestFactory-nil-"+factoryID.String())
	if err != nil {
		t.Fatalf("insert factory: %v", err)
	}
	t.Cleanup(func() {
		db.Exec(`DELETE FROM nxd.assets WHERE factory_id = $1`, factoryID)
		db.Exec(`DELETE FROM nxd.factories WHERE id = $1`, factoryID)
	})

	assetID, err := CreateAsset(db, factoryID, nil, "TAG-NIL-001", "", "", nil)
	if err != nil {
		t.Fatalf("CreateAsset with nil groupID: %v", err)
	}
	if assetID == uuid.Nil {
		t.Fatal("expected valid assetID")
	}

	var groupID sql.NullString
	db.QueryRow(`SELECT group_id FROM nxd.assets WHERE id = $1`, assetID).Scan(&groupID)
	if groupID.Valid {
		t.Errorf("expected group_id=NULL for nil input, got %v", groupID.String)
	}
	t.Logf("✓ TestCreateAsset_NilGroupID_DoesNotPanic passed (assetID=%s)", assetID)
}
