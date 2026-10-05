# Scheduling rules

A daily job runs once per calendar day at a wall-clock time, such as 02:30.

- `NextRun` returns the first run strictly after the given instant.
- Wall-clock times are interpreted in the customer's time zone. The result is
  an instant; callers compare instants, not clock readings.
- When a daylight-saving change skips the wall-clock time (for example 02:30
  on a spring-forward day), the job runs at the first valid instant after the
  gap, which is the moment the clocks change.
- When a daylight-saving change repeats the wall-clock time (for example
  02:30 on a fall-back day in Europe), the job runs once, at the first
  occurrence. The second occurrence is not a run.
- A job never runs twice on the same calendar day, and never skips a day.
