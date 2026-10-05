package notify

import (
	"fmt"
	"sync"
)

// Job is one attempt to deliver a notification.
type Job struct {
	ID      int
	Key     string
	To      string
	Body    string
	Attempt int
}

// Queue holds pending notification jobs in FIFO order.
type Queue struct {
	mu   sync.Mutex
	jobs []Job
	next int
}

// Enqueue adds a new notification and gives it a new idempotency key.
func (q *Queue) Enqueue(to, body string) Job {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.next++
	j := Job{ID: q.next, Key: fmt.Sprintf("n-%d", q.next), To: to, Body: body, Attempt: 1}
	q.jobs = append(q.jobs, j)
	return j
}

// Push adds an existing job to the end of the queue.
func (q *Queue) Push(j Job) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.jobs = append(q.jobs, j)
}

// Pop removes and returns the oldest job.
func (q *Queue) Pop() (Job, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.jobs) == 0 {
		return Job{}, false
	}
	j := q.jobs[0]
	q.jobs = q.jobs[1:]
	return j, true
}
