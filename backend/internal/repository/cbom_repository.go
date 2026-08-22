package repository

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/quantumpq/qv-cbom/internal/models"
)

// SessionStore manages per-client isolated MemoryRepository instances.
// Each client browser session gets its own isolated memory bucket preventing
// data collision when multiple users scan simultaneously.
type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*sessionEntry
}

type sessionEntry struct {
	repo      *MemoryRepository
	lastUsed  time.Time
}

func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string]*sessionEntry),
	}
}

// GetOrCreate returns the MemoryRepository for the given session ID,
// creating a new isolated one if it doesn't exist yet.
func (ss *SessionStore) GetOrCreate(sessionID string) *MemoryRepository {
	if sessionID == "" {
		sessionID = "default"
	}

	// Fast path: read lock
	ss.mu.RLock()
	entry, exists := ss.sessions[sessionID]
	ss.mu.RUnlock()

	if exists {
		ss.mu.Lock()
		ss.sessions[sessionID].lastUsed = time.Now()
		ss.mu.Unlock()
		return entry.repo
	}

	// Slow path: create new session bucket
	ss.mu.Lock()
	defer ss.mu.Unlock()

	// Double-check after acquiring write lock
	if entry, exists = ss.sessions[sessionID]; exists {
		ss.sessions[sessionID].lastUsed = time.Now()
		return entry.repo
	}

	newRepo := NewMemoryRepository()
	ss.sessions[sessionID] = &sessionEntry{
		repo:     newRepo,
		lastUsed: time.Now(),
	}
	log.Printf("[Session] Created new session bucket: %s (total sessions: %d)", sessionID[:min(8, len(sessionID))], len(ss.sessions))
	return newRepo
}

// StartCleanup runs a background goroutine that evicts sessions inactive
// longer than the given TTL. Call this once at server startup.
func (ss *SessionStore) StartCleanup(ttl time.Duration) {
	ticker := time.NewTicker(ttl / 2)
	go func() {
		for range ticker.C {
			ss.evictExpired(ttl)
		}
	}()
}

func (ss *SessionStore) evictExpired(ttl time.Duration) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	now := time.Now()
	evicted := 0
	for id, entry := range ss.sessions {
		if now.Sub(entry.lastUsed) > ttl {
			delete(ss.sessions, id)
			evicted++
		}
	}
	if evicted > 0 {
		log.Printf("[Session] Evicted %d expired sessions (remaining: %d)", evicted, len(ss.sessions))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}



type CBOMRepository interface {
	CreateTarget(t models.Target) error
	ListTargets() ([]models.Target, error)
	GetTarget(id string) (*models.Target, error)

	SaveScan(summary models.ScanSummary, assets []models.CryptoAsset) error
	GetScans() ([]models.ScanSummary, error)
	GetScanByID(id string) (*models.ScanDetail, error)
	GetScansByTarget(targetID string) ([]models.ScanSummary, error)
	GetLatestScanByTargetID(targetID string) (*models.ScanDetail, error)
	SaveMigrationPlan(plan models.MigrationPlan) error
	GetMigrationPlanByScanID(scanID string) (*models.MigrationPlan, error)
}

// MemoryRepository implements pure in-memory, thread-safe storage (Zero-DB / Single Active Scan mode)
type MemoryRepository struct {
	mu             sync.RWMutex
	targets        map[string]models.Target
	scans          map[string]models.ScanSummary
	assets         map[string][]models.CryptoAsset
	migrationPlans map[string]models.MigrationPlan
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		targets:        make(map[string]models.Target),
		scans:          make(map[string]models.ScanSummary),
		assets:         make(map[string][]models.CryptoAsset),
		migrationPlans: make(map[string]models.MigrationPlan),
	}
}

func (r *MemoryRepository) CreateTarget(t models.Target) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// In single active scan mode, reset old targets so only current active target exists
	r.targets = make(map[string]models.Target)

	if t.ID == "" {
		t.ID = fmt.Sprintf("target-%s", models.GenerateUUID()[:8])
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
	r.targets[t.ID] = t
	return nil
}

func (r *MemoryRepository) ListTargets() ([]models.Target, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []models.Target
	for _, t := range r.targets {
		list = append(list, t)
	}
	return list, nil
}

func (r *MemoryRepository) GetTarget(id string) (*models.Target, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, exists := r.targets[id]
	if !exists {
		return nil, fmt.Errorf("target not found: %s", id)
	}
	return &t, nil
}

func (r *MemoryRepository) GetScansByTarget(targetID string) ([]models.ScanSummary, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []models.ScanSummary
	for _, scan := range r.scans {
		if scan.TargetID == targetID {
			result = append(result, scan)
		}
	}
	return result, nil
}

func (r *MemoryRepository) SaveScan(summary models.ScanSummary, assets []models.CryptoAsset) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// In single active scan mode, keep only the latest active scan
	r.scans = make(map[string]models.ScanSummary)
	r.assets = make(map[string][]models.CryptoAsset)
	r.migrationPlans = make(map[string]models.MigrationPlan)

	if summary.CreatedAt.IsZero() {
		summary.CreatedAt = time.Now()
	}
	for i := range assets {
		if assets[i].CreatedAt.IsZero() {
			assets[i].CreatedAt = time.Now()
		}
	}

	r.scans[summary.ID] = summary
	r.assets[summary.ID] = assets
	return nil
}

func (r *MemoryRepository) GetScans() ([]models.ScanSummary, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []models.ScanSummary
	for _, scan := range r.scans {
		result = append(result, scan)
	}
	return result, nil
}

func (r *MemoryRepository) GetScanByID(id string) (*models.ScanDetail, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	summary, exists := r.scans[id]
	if !exists {
		return nil, fmt.Errorf("scan not found: %s", id)
	}

	assets := r.assets[id]
	return &models.ScanDetail{
		Summary: summary,
		Assets:  assets,
	}, nil
}

func (r *MemoryRepository) GetLatestScanByTargetID(targetID string) (*models.ScanDetail, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var latestSummary models.ScanSummary
	found := false

	for _, s := range r.scans {
		if s.TargetID == targetID {
			if !found || s.CreatedAt.After(latestSummary.CreatedAt) {
				latestSummary = s
				found = true
			}
		}
	}

	if !found {
		return nil, fmt.Errorf("no scans found for target: %s", targetID)
	}

	assets := r.assets[latestSummary.ID]
	return &models.ScanDetail{
		Summary: latestSummary,
		Assets:  assets,
	}, nil
}

func (r *MemoryRepository) SaveMigrationPlan(plan models.MigrationPlan) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.migrationPlans[plan.ScanID] = plan
	return nil
}

func (r *MemoryRepository) GetMigrationPlanByScanID(scanID string) (*models.MigrationPlan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plan, exists := r.migrationPlans[scanID]
	if !exists {
		return nil, fmt.Errorf("migration plan not found for scan: %s", scanID)
	}
	return &plan, nil
}
