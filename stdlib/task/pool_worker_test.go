package task_test

import (
	"sync"
	"testing"
	"time"

	"github.com/arcspace/go-arc-sdk/stdlib/task"
	"github.com/stretchr/testify/require"
)

func TestPoolWorker(t *testing.T) {
	t.Run("disallows simultaneous processing of items with the same UniqueID", func(t *testing.T) {
		t.Parallel()

		var wg sync.WaitGroup
		item := makeItems()
		item[1].block = make(chan struct{})
		item[1].retry = true

		w, err := task.StartNewPoolWorker("", 2, task.NewStaticScheduler(100*time.Millisecond, 2*time.Second))

		require.NoError(t, err)
		defer w.Close()

		w.Add(item[1])
		w.Add(item[2])

		var which *workItem
		wg.Add(1)
		go func() {
			select {
			case <-item[1].processed:
				which = item[1]
			case <-item[2].processed:
				which = item[2]
			}

			select {
			case <-item[1].processed:
				t.Fatalf("nope")
			case <-item[2].processed:
				t.Fatalf("nope")
			case <-time.After(1 * time.Second):
			}
			wg.Done()
		}()
		wg.Wait()

		close(which.block)
		wg.Add(1)
		go func() {
			select {
			case <-item[1].processed:
			case <-item[2].processed:
			}

			select {
			case <-item[1].processed:
				t.Fatalf("nope")
			case <-item[2].processed:
				t.Fatalf("nope")
			case <-time.After(1 * time.Second):
			}
			wg.Done()
		}()
		wg.Wait()
	})
}
