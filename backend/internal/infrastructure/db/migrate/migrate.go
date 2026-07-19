// Package migrate is a tiny, dependency-free migration engine. Each migration is a
// Go file that registers itself in init(); Up/Down each run inside a transaction.
package migrate

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"gorm.io/gorm"
)

// Migration is a single, ordered schema change. Up applies it; Down reverts it.
type Migration struct {
	Version int64
	Name    string
	Up      func(tx *gorm.DB) error
	Down    func(tx *gorm.DB) error
}

type record struct {
	Version   int64     `gorm:"primaryKey"`
	Name      string    `gorm:"size:200;not null"`
	AppliedAt time.Time `gorm:"not null;default:now()"`
}

func (record) TableName() string { return "schema_migrations" }

var registry []Migration

// Register adds a migration. Call from init() in each migration file.
func Register(m Migration) {
	if m.Up == nil {
		panic(fmt.Sprintf("migrate: version %d (%s) has nil Up", m.Version, m.Name))
	}
	for _, e := range registry {
		if e.Version == m.Version {
			panic(fmt.Sprintf("migrate: duplicate version %d", m.Version))
		}
	}
	registry = append(registry, m)
}

// All returns registered migrations sorted ascending by version.
func All() []Migration {
	out := make([]Migration, len(registry))
	copy(out, registry)
	sort.Slice(out, func(i, j int) bool { return out[i].Version < out[j].Version })
	return out
}

func ensureTable(db *gorm.DB) error { return db.AutoMigrate(&record{}) }

func applied(db *gorm.DB) (map[int64]record, error) {
	var rows []record
	if err := db.Order("version asc").Find(&rows).Error; err != nil {
		return nil, err
	}
	m := make(map[int64]record, len(rows))
	for _, r := range rows {
		m[r.Version] = r
	}
	return m, nil
}

// Up applies all pending migrations in order.
func Up(db *gorm.DB) ([]int64, error) {
	if err := ensureTable(db); err != nil {
		return nil, fmt.Errorf("ensure table: %w", err)
	}
	done, err := applied(db)
	if err != nil {
		return nil, err
	}
	var ran []int64
	for _, m := range All() {
		if _, ok := done[m.Version]; ok {
			continue
		}
		err := db.Transaction(func(tx *gorm.DB) error {
			if err := m.Up(tx); err != nil {
				return err
			}
			return tx.Create(&record{Version: m.Version, Name: m.Name, AppliedAt: time.Now().UTC()}).Error
		})
		if err != nil {
			return ran, fmt.Errorf("apply %d %s: %w", m.Version, m.Name, err)
		}
		ran = append(ran, m.Version)
	}
	return ran, nil
}

// Down rolls back the last n applied migrations (newest first).
func Down(db *gorm.DB, n int) ([]int64, error) {
	if n <= 0 {
		return nil, errors.New("n must be > 0")
	}
	if err := ensureTable(db); err != nil {
		return nil, fmt.Errorf("ensure table: %w", err)
	}
	done, err := applied(db)
	if err != nil {
		return nil, err
	}
	all := All()
	byVer := make(map[int64]Migration, len(all))
	for _, m := range all {
		byVer[m.Version] = m
	}

	versions := make([]int64, 0, len(done))
	for v := range done {
		versions = append(versions, v)
	}
	sort.Slice(versions, func(i, j int) bool { return versions[i] > versions[j] })
	if n > len(versions) {
		n = len(versions)
	}

	var rolled []int64
	for i := 0; i < n; i++ {
		v := versions[i]
		m, ok := byVer[v]
		if !ok {
			return rolled, fmt.Errorf("rollback: migration %d not found in code", v)
		}
		if m.Down == nil {
			return rolled, fmt.Errorf("rollback: migration %d (%s) has no Down", v, m.Name)
		}
		err := db.Transaction(func(tx *gorm.DB) error {
			if err := m.Down(tx); err != nil {
				return err
			}
			return tx.Delete(&record{}, "version = ?", v).Error
		})
		if err != nil {
			return rolled, fmt.Errorf("revert %d %s: %w", v, m.Name, err)
		}
		rolled = append(rolled, v)
	}
	return rolled, nil
}

// StatusRow describes one migration's applied state.
type StatusRow struct {
	Version   int64
	Name      string
	Applied   bool
	AppliedAt time.Time
}

func Status(db *gorm.DB) ([]StatusRow, error) {
	if err := ensureTable(db); err != nil {
		return nil, err
	}
	done, err := applied(db)
	if err != nil {
		return nil, err
	}
	all := All()
	out := make([]StatusRow, 0, len(all)+len(done))
	seen := make(map[int64]bool, len(all))
	for _, m := range all {
		seen[m.Version] = true
		r := StatusRow{Version: m.Version, Name: m.Name}
		if rec, ok := done[m.Version]; ok {
			r.Applied = true
			r.AppliedAt = rec.AppliedAt
		}
		out = append(out, r)
	}
	for v, rec := range done {
		if !seen[v] {
			out = append(out, StatusRow{Version: v, Name: rec.Name + " (orphan)", Applied: true, AppliedAt: rec.AppliedAt})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Version < out[j].Version })
	return out, nil
}
