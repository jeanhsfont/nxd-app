package store

import (
	"database/sql"
	"hubsystem/core"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// ─── API Key lookup cache ─────────────────────────────────────────────────────
// Caches the result of bcrypt comparison (expensive: ~100ms each) per API key.
// TTL = 60s. A DX sending every 10s triggers bcrypt at most once per minute
// instead of on every request.
//
// Security: the cache stores only the resolved factory struct (no hash exposed).
// An attacker who observes network traffic cannot replay the cache — they would
// need the full valid key to populate a cache entry.

const apiKeyCacheTTL = 60 * time.Second

type apiKeyCacheEntry struct {
	factory   *core.Factory // nil means "key was checked and not found"
	expiresAt time.Time
}

var (
	apiKeyCacheMu sync.RWMutex
	apiKeyCache   = map[string]apiKeyCacheEntry{}
)

// InvalidateAPIKeyCache removes a key from cache. Call after key regeneration.
func InvalidateAPIKeyCache(apiKey string) {
	apiKeyCacheMu.Lock()
	delete(apiKeyCache, apiKey)
	apiKeyCacheMu.Unlock()
}

// GetFactoryByAPIKey authenticates an API key and returns the matching factory.
//
// Two-tier strategy:
//  1. In-memory cache: O(1) map lookup, ~1µs. Cache TTL = 60s.
//  2. Cache miss → prefix-indexed DB query: WHERE api_key_prefix = first 16 chars.
//     Returns at most 1 row (unique index). bcrypt runs only on that row (~100ms).
//  3. Fallback: full-scan limited to rows WHERE api_key_prefix IS NULL (legacy
//     factories before migration). Self-heals: backfills prefix on match.
//
// Previous behavior was O(N·bcrypt) for all factories on every request.
func GetFactoryByAPIKey(db *sql.DB, apiKey string) (*core.Factory, error) {
	// ── Tier 1: in-memory cache ───────────────────────────────────────────
	apiKeyCacheMu.RLock()
	if entry, ok := apiKeyCache[apiKey]; ok && time.Now().Before(entry.expiresAt) {
		apiKeyCacheMu.RUnlock()
		return entry.factory, nil
	}
	apiKeyCacheMu.RUnlock()

	// ── Tier 2: prefix-indexed DB lookup ─────────────────────────────────
	prefix := ""
	if len(apiKey) >= 16 {
		prefix = apiKey[:16]
	}

	var factory *core.Factory

	if prefix != "" {
		rows, err := db.Query(
			`SELECT id, user_id, name, api_key_hash, created_at, is_active
			 FROM nxd.factories
			 WHERE api_key_prefix = $1`,
			prefix,
		)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var f core.Factory
			var factoryID, userID uuid.UUID
			var apiKeyHash []byte
			if err := rows.Scan(&factoryID, &userID, &f.Name, &apiKeyHash, &f.CreatedAt, &f.IsActive); err != nil {
				return nil, err
			}
			if bcrypt.CompareHashAndPassword(apiKeyHash, []byte(apiKey)) == nil {
				f.ID = factoryID.String()
				f.UserID = userID.String()
				f.APIKey = apiKey
				factory = &f
				break
			}
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}

	// ── Tier 3: fallback full-scan for legacy rows without prefix ─────────
	// Runs only for factories where api_key_prefix was not backfilled yet.
	// Self-healing: on first match, writes the prefix so next call uses Tier 2.
	if factory == nil {
		log.Printf("⚠️  [AUTH] Prefix miss for %.12s... — scanning legacy factories (api_key_prefix IS NULL)", apiKey)
		allRows, err := db.Query(
			`SELECT id, user_id, name, api_key_hash, created_at, is_active
			 FROM nxd.factories
			 WHERE api_key_prefix IS NULL`,
		)
		if err != nil {
			return nil, err
		}
		defer allRows.Close()

		for allRows.Next() {
			var f core.Factory
			var factoryID, userID uuid.UUID
			var apiKeyHash []byte
			if err := allRows.Scan(&factoryID, &userID, &f.Name, &apiKeyHash, &f.CreatedAt, &f.IsActive); err != nil {
				return nil, err
			}
			if bcrypt.CompareHashAndPassword(apiKeyHash, []byte(apiKey)) == nil {
				f.ID = factoryID.String()
				f.UserID = userID.String()
				f.APIKey = apiKey
				factory = &f
				// Self-heal: backfill prefix so next lookup uses the fast path.
				if prefix != "" {
					if _, upErr := db.Exec(
						`UPDATE nxd.factories SET api_key_prefix = $1 WHERE id = $2`,
						prefix, factoryID,
					); upErr != nil {
						log.Printf("⚠️  [AUTH] Failed to backfill api_key_prefix for factory %s: %v", factoryID, upErr)
					} else {
						log.Printf("✓  [AUTH] Backfilled api_key_prefix for factory %s", factoryID)
					}
				}
				break
			}
		}
		if err := allRows.Err(); err != nil {
			return nil, err
		}
	}

	// ── Store in cache (nil = key not found; also cached to block brute-force) ─
	apiKeyCacheMu.Lock()
	apiKeyCache[apiKey] = apiKeyCacheEntry{factory: factory, expiresAt: time.Now().Add(apiKeyCacheTTL)}
	apiKeyCacheMu.Unlock()

	return factory, nil
}

// RegenerateAPIKey generates a new API key, saves its hash and prefix.
// Also invalidates the old key from the in-memory cache.
func RegenerateAPIKey(db *sql.DB, factoryID uuid.UUID) (string, error) {
	newAPIKey, hash, err := core.GenerateAndHashAPIKey()
	if err != nil {
		return "", err
	}

	prefix := ""
	if len(newAPIKey) >= 16 {
		prefix = newAPIKey[:16]
	}

	_, err = db.Exec(
		`UPDATE nxd.factories
		 SET api_key_hash = $1, api_key = $2, api_key_prefix = $3, updated_at = NOW()
		 WHERE id = $4`,
		hash, newAPIKey, prefix, factoryID,
	)
	if err != nil {
		return "", err
	}

	// Invalidate all cache entries for this factory (we don't know the old key,
	// so we can't invalidate by key directly — instead flush the full cache which
	// is safe since it's just a performance cache, not a security boundary).
	apiKeyCacheMu.Lock()
	apiKeyCache = map[string]apiKeyCacheEntry{} // full flush on key regen
	apiKeyCacheMu.Unlock()

	return newAPIKey, nil
}
