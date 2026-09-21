# Changelog

All notable changes to `cortiqa-sdk-go` will be documented in this file.

## [v0.1.0] - 2026-09-21

### Added
- Initial release of the official Cortiqa Go SDK.
- Pure Go standard library architecture (zero third-party dependencies).
- Context-aware methods (`context.Context`).
- OpenAI & Anthropic style service interfaces (`client.Chat` and `client.Messages`).
- SSE streaming client (`CreateStream` / `Recv()`).
- Auto-retries with exponential backoff on 429/5xx status codes.
- Typed error models (`AuthenticationError`, `RateLimitError`, `NotFoundError`, `APIError`).
