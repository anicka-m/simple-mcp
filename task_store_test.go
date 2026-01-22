package main

import (
	"fmt"
	"strings"
	"sync"
	"testing"
)

func TestTaskStore_CreateAndGet(t *testing.T) {
	ts := NewTaskStore()
	id := "job-123"
	tool := "SystemUpgrade"

	task := ts.Create(id, tool)
	if task.ID != id {
		t.Errorf("expected ID %s, got %s", id, task.ID)
	}
	if task.Status != "pending" {
		t.Errorf("expected status pending, got %s", task.Status)
	}

	retrieved, ok := ts.Get(id)
	if !ok {
		t.Fatalf("failed to retrieve task %s", id)
	}
	if retrieved != task {
		t.Errorf("retrieved task pointer mismatch")
	}

	// Test case-insensitivity
	retrievedLower, ok := ts.Get(strings.ToLower(id))
	if !ok {
		t.Errorf("failed to retrieve task with lowercase ID")
	}
	if retrievedLower != task {
		t.Errorf("retrieved lowercase task pointer mismatch")
	}

	retrievedUpper, ok := ts.Get(strings.ToUpper(id))
	if !ok {
		t.Errorf("failed to retrieve task with uppercase ID")
	}
	if retrievedUpper != task {
		t.Errorf("retrieved uppercase task pointer mismatch")
	}
}

func TestTaskStore_Concurrency(t *testing.T) {
	// this test verifies the store doesn't panic under concurrent access
	ts := NewTaskStore()
	var wg sync.WaitGroup

	// concurrently create tasks
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			id := fmt.Sprintf("job-%d", i)
			ts.Create(id, "test-tool")
			ts.SetStatus(id, "running", "working")
			_, _ = ts.Get(id)
		}(i)
	}

	wg.Wait()
	
	tasks := ts.ListActiveTasks()
	// All tasks are pending/running, so we expect 100 active tasks
	if len(tasks) != 100 {
		t.Errorf("expected 100 active tasks, got %d", len(tasks))
	}
}

func TestTaskStore_HasActiveTask(t *testing.T) {
	ts := NewTaskStore()
	ts.Create("1", "Upgrade")
	ts.SetStatus("1", "running", "...")

	if !ts.HasActiveTask("Upgrade") {
		t.Error("expected HasActiveTask to return true")
	}

	ts.SetStatus("1", "completed", "done")
	if ts.HasActiveTask("Upgrade") {
		t.Error("expected HasActiveTask to return false after completion")
	}
}
