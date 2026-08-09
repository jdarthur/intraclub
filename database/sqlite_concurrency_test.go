package database

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
)

// TestSqliteProviderConcurrentCRUD exercises parallel CRUD against the SQLite
// provider (issue #63). The provider opens the database in WAL mode with a
// busy timeout (see newSqliteDbProvider), which allows concurrent readers
// while writes are serialized on the single writer lock. This test drives
// Create/Read/Update/Delete from many goroutines at once to catch
// "database is locked" failures that could surface from concurrent Gin
// handlers.
func TestSqliteProviderConcurrentCRUD(t *testing.T) {
	ctx := context.Background()
	p := newTestSqliteProvider(t)

	const workers = 16
	const opsPerWorker = 30

	var wg sync.WaitGroup
	errCh := make(chan error, workers*opsPerWorker)
	start := make(chan struct{})

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			<-start
			for i := 0; i < opsPerWorker; i++ {
				rec := &sqliteTestRecord{
					Name:    fmt.Sprintf("w%d-%d", w, i),
					Age:     w + i,
					Enabled: i%2 == 0,
				}
				created, err := CreateOne(ctx, p, rec)
				if err != nil {
					errCh <- fmt.Errorf("create: %w", err)
					continue
				}

				if _, exists, err := GetOneById(ctx, p, &sqliteTestRecord{}, created.GetId()); err != nil || !exists {
					errCh <- fmt.Errorf("get: exists=%v err=%v", exists, err)
					continue
				}

				created.Age++
				if err := UpdateOne(ctx, p, created); err != nil {
					errCh <- fmt.Errorf("update: %w", err)
					continue
				}

				if _, exists, err := DeleteOneById(ctx, p, &sqliteTestRecord{}, created.GetId()); err != nil || !exists {
					errCh <- fmt.Errorf("delete: exists=%v err=%v", exists, err)
					continue
				}
			}
		}(w)
	}

	close(start)
	wg.Wait()
	close(errCh)

	var firstErr error
	for err := range errCh {
		if err == nil {
			continue
		}
		if strings.Contains(err.Error(), "database is locked") {
			t.Fatalf("SQLite returned 'database is locked' during concurrent CRUD: %v", err)
		}
		if firstErr == nil {
			firstErr = err
		}
	}
	if firstErr != nil {
		t.Fatalf("concurrent CRUD failed: %v", firstErr)
	}
}
