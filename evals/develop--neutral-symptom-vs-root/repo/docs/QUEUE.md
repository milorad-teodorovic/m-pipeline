# Notification queue

The queue delivers each job at least once. A job can run more than once:
the worker retries a job when the mail call fails or runs past the timeout,
and a slow call can still deliver after the worker gave up waiting.

Every logical notification has one idempotency key, set once by the producer
in `Queue.Enqueue`. The key must stay the same for every attempt of that
notification, including retries. `Mailer.Send` delivers a key at most once,
so a retry that keeps the key cannot send a second email.

Two notifications are different when they have different keys, even if the
recipient and the text are the same. For example, two order confirmations
with the same wording are two emails. Never deduplicate on recipient or text.
