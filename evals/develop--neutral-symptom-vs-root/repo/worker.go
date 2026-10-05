package notify

import (
	"context"
	"sync"
	"time"
)

// Worker takes jobs from a Queue and sends them with a Mailer.
type Worker struct {
	Queue   *Queue
	Mailer  *Mailer
	Timeout time.Duration
	pending sync.WaitGroup
}

// Drain processes jobs until the queue is empty, then waits for calls that
// outlived their timeout to finish.
func (w *Worker) Drain() {
	for {
		j, ok := w.Queue.Pop()
		if !ok {
			break
		}
		if err := w.attempt(j); err != nil {
			w.requeue(j)
		}
	}
	w.pending.Wait()
}

func (w *Worker) attempt(j Job) error {
	ctx, cancel := context.WithTimeout(context.Background(), w.Timeout)
	defer cancel()
	done := make(chan error, 1)
	w.pending.Add(1)
	go func() {
		defer w.pending.Done()
		done <- w.Mailer.Send(context.Background(), j.Key, j.To, j.Body)
	}()
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
