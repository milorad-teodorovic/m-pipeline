# Outbound call reliability

Owner: Platform team. Applies to every client that calls a third-party API.

## When to retry

- Retry only requests that are safe to repeat: GET, HEAD, OPTIONS, PUT, and DELETE.
- A POST or PATCH is retried only when it carries an `Idempotency-Key` header.
- Retry only when the partner answers 429 Too Many Requests or 503 Service
  Unavailable, or when the transport fails before any response arrives.
  Every other status, including 500 and 502, is final: the partner documents
  those as permanent failures, and a retry repeats a side effect.
- Make at most 3 attempts in total.

## How long to wait

- When the response has a `Retry-After` header, wait exactly that long. The
  header is either a number of seconds or an HTTP date; support both.
- Without a usable `Retry-After`, wait 200 ms before the second attempt and
  400 ms before the third.
- Never start a wait that would end after the deadline of the request context.
  Return the last response instead.

## Request bodies

A retry must send the same body as the first attempt. A request whose body
cannot be read again (`Request.GetBody` is nil) is sent once and never retried.

## Responses

Close the body of every response you do not return to the caller.
