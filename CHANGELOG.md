# Changelog

All notable changes to the UniPost Go SDK are documented here.

## [0.6.0]

- Added explicit managed-user and owner/admin workspace Inbox scopes for direct
  messages, comments, and replies, preventing managed-user reads from falling
  back to the workspace aggregate.
- Added limit-only Inbox listing with source, read, own-message, and limit
  filters, plus unread, item, read-state, thread-state, and media operations.
- Added response-aware Inbox replies with stable idempotency keys: completed
  writes return the created item, while accepted X writes expose reconciliation
  state for polling.
- Added backend-only WebSocket connection details that keep the API key in the
  authorization header and out of URLs.
- Added ordinary Inbox sync and typed, metered X backfill estimate,
  confirmation, in-progress, and completed results.

## [0.5.0]

- Added custom audio overlay jobs and polling for uploaded video/audio media.
- Made media upload reserve-time file sizes optional.
- Preserved typed post failure contract fields and expanded supported analytics
  surfaces.
