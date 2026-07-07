# ADR-008: API error response format

**Status:** Accepted  
**Date:** 2026-06-13

## Context

Options considered:

- Plain string: `{"error": "something went wrong"}`
- With machine-readable code: `{"error": "message", "code": "snake_case_code"}`
- RFC 7807 Problem Details: `{"type": "...", "title": "...", "status": 400, "detail": "..."}`

## Decision

```json
{
  "error": "Human-readable message",
  "code": "snake_case_error_code"
}
```

HTTP status code carries the category (400, 401, 403, 404, 422, 500). The `code` field lets the frontend handle specific errors programmatically without string-matching the message.

Example codes: `invalid_credentials`, `monitor_limit_reached`, `slug_taken`.

## Consequences

- **Simple:** one struct, no extra fields
- **Extensible:** a `details` field can be added later for validation errors without breaking existing clients
- **RFC 7807 rejected:** overkill for a small API with a single frontend client; the `type` URI convention adds no value at this scale
