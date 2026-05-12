// Package catalog provides the task type catalog used by cm-cicada to
// route TaskComponents to specific Airflow operators and validate user input.
package catalog

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

var (
	store    map[string]TaskTypeDef
	ordered  []TaskTypeDef
	storeMu  sync.RWMutex
)

// Load reads the catalog yaml file and populates the in-memory store.
// Should be called once during application bootstrap, after config is loaded.
func Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read task types catalog (%s): %w", path, err)
	}

	var f catalogFile
	if err := yaml.Unmarshal(data, &f); err != nil {
		return fmt.Errorf("failed to parse task types catalog: %w", err)
	}

	if len(f.TaskTypes) == 0 {
		return errors.New("task types catalog is empty")
	}

	next := make(map[string]TaskTypeDef, len(f.TaskTypes))
	for _, t := range f.TaskTypes {
		if t.ID == "" {
			return errors.New("task type id must not be empty")
		}
		if t.OperatorClass == "" {
			return fmt.Errorf("task type %q operator_class must not be empty", t.ID)
		}
		if _, dup := next[t.ID]; dup {
			return fmt.Errorf("duplicate task type id: %s", t.ID)
		}
		next[t.ID] = t
	}

	storeMu.Lock()
	store = next
	ordered = append(ordered[:0], f.TaskTypes...)
	storeMu.Unlock()
	return nil
}

// Get returns the TaskTypeDef for the given id, or false if not found.
func Get(id string) (TaskTypeDef, bool) {
	storeMu.RLock()
	defer storeMu.RUnlock()
	t, ok := store[id]
	return t, ok
}

// List returns all task type definitions in catalog file order.
func List() []TaskTypeDef {
	storeMu.RLock()
	defer storeMu.RUnlock()
	out := make([]TaskTypeDef, len(ordered))
	copy(out, ordered)
	return out
}

// Has reports whether the given id exists in the catalog.
func Has(id string) bool {
	_, ok := Get(id)
	return ok
}
