# GEMINI Development Guidelines

Auto-generated from all feature plans. Last updated: 2025-09-19

## Active Technologies
- Go 1.21
- Angular (TypeScript)
- SQLite

## Project Structure
```
apps/
├── backend/
│   ├── cmd/
│   ├── internal/
│   └── ...
└── frontend/
    └── file-manager-frontend/
        ├── src/
        └── ...
```

## Commands
- `go test ./...` (run from `apps/backend`)
- `npm start` (run from `apps/frontend/file-manager-frontend`)

## Code Style
- Go: Standard Go formatting (`gofmt`)
- Angular/TypeScript: Follows Angular style guide, uses Prettier for formatting.

## Recent Changes
- **feat(frontend): implement basic sync UI and connect to backend API**
  - Created a new `sync` component.
  - Updated the API service to match the backend spec.
  - Implemented new API handlers in the backend.

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
