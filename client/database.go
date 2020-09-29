package client

import (
	"errors"

	"github.com/brutella/hc/db"
)

var _ db.Database = (*memoryDB)(nil)

// memoryDB implements transient storage for db.Database
type memoryDB struct {
	entities map[string]db.Entity
}

// EntityWithName returns the entity referenced by name. Returns an error
// if the entity is not found.
func (m *memoryDB) EntityWithName(name string) (db.Entity, error) {
	entity, ok := m.entities[name]
	if !ok {
		return db.Entity{}, errors.New("not found")
	}
	return entity, nil
}

// SaveEntity saves a entity in the database
func (m *memoryDB) SaveEntity(entity db.Entity) error {
	if entity.Name == "" {
		return errors.New("entity missing name")
	}

	if m.entities == nil {
		m.entities = make(map[string]db.Entity)
	}

	m.entities[entity.Name] = entity
	return nil
}

// MustSaveEntity saves a entity in the database and panics on error
func (m *memoryDB) MustSaveEntity(entity db.Entity) {
	if err := m.SaveEntity(entity); err != nil {
		panic(err)
	}
}

// DeleteEntity deletes a entity from the database
func (m *memoryDB) DeleteEntity(entity db.Entity) {
	if entity.Name == "" {
		return
	}

	delete(m.entities, entity.Name)
}

// Entities returns all entities
func (m *memoryDB) Entities() ([]db.Entity, error) {
	result := make([]db.Entity, 0, len(m.entities))
	for _, e := range m.entities {
		result = append(result, e)
	}
	return result, nil
}
