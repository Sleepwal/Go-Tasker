package storage

import (
	"context"
	"testing"
	"time"
)

type storageTestJob struct {
	id string
}

func (tj *storageTestJob) ID() string {
	return tj.id
}

func (tj *storageTestJob) Run(ctx context.Context) error {
	return nil
}

func TestMemoryStorage_SaveAndList(t *testing.T) {
	storage := NewMemoryStorage()

	j := &storageTestJob{id: "test1"}
	runAt := time.Now().Add(1 * time.Second)
	err := storage.Save(j, runAt)
	if err != nil {
		t.Errorf("Save failed: %v", err)
	}

	ready, err := storage.ListReady(time.Now().Add(2 * time.Second))
	if err != nil {
		t.Errorf("ListReady failed: %v", err)
	}

	if len(ready) != 1 {
		t.Errorf("expected 1 ready job, got %d", len(ready))
	}

	if ready[0].ID() != "test1" {
		t.Errorf("expected job id test1, got %s", ready[0].ID())
	}
}

func TestMemoryStorage_Delete(t *testing.T) {
	storage := NewMemoryStorage()

	j := &storageTestJob{id: "test1"}
	storage.Save(j, time.Now())

	err := storage.Delete("test1")
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}

	ready, _ := storage.ListReady(time.Now())
	if len(ready) != 0 {
		t.Errorf("expected 0 ready jobs after delete, got %d", len(ready))
	}
}
