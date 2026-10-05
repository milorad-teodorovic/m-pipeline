package notify

// MaxAttempts is the number of times a notification is tried before it is dropped.
const MaxAttempts = 5

func (w *Worker) requeue(j Job) {
	if j.Attempt >= MaxAttempts {
		return
	}
	next := w.Queue.Enqueue(j.To, j.Body)
	next.Attempt = j.Attempt + 1
}
