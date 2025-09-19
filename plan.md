# Implementation Plan: File Manager

**Branch**: `main` | **Date**: 2025-09-19 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `spec.md`

## Summary
This project is to build a local file backup and synchronization tool. The core functionality includes selecting source and destination directories, performing synchronization with reliable deduplication, and providing a web-based GUI for management. The folder selection will be done through a custom, in-app file explorer modal that communicates with the backend to display the user's file structure, starting from their home directory.

## Technical Context
**Language/Version**: Go 1.21, TypeScript (latest for Angular)
**Primary Dependencies**: Go: `github.com/mattn/go-sqlite3`, Angular: `@angular/core`
**Storage**: SQLite
**Testing**: Go testing library, Jest/Jasmine for Angular
**Target Platform**: Local desktop (cross-platform)
**Project Type**: Web application (frontend + backend)
**Performance Goals**: [NEEDS CLARIFICATION: e.g., sync speed, UI responsiveness]
**Constraints**: [NEEDS CLARIFICATION: e.g., memory usage, CPU load]
**Scale/Scope**: [NEEDS CLARIFICATION: e.g., max number of files, max file size]

## Constitution Check
*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

[Gates determined based on constitution file]

## Project Structure

### Documentation
```
./
├── plan.md              # This file
├── spec.md              # Feature specification
├── tasks.md             # Generated tasks
└── apps/
    ├── frontend/
    │   ├── plan-frontend.md
    │   └── PRD-frontend.md
    └── backend/
```

### Source Code (repository root)
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

**Structure Decision**: Web application (frontend + backend)

## File Explorer and Sync Logic

### Frontend: File Explorer UI
To provide a good user experience for folder selection, a custom file explorer will be built as a modal dialog in the frontend.
- A new `FileExplorerModalComponent` will be created in Angular.
- This component will display the file structure and allow navigation, starting from the user's home directory.
- It will communicate with the backend to fetch directory contents on demand (lazy loading).
- When a folder is selected, the modal will close and pass the selected path to the sync form.

### Backend: File Listing API
- A new API endpoint (`/api/files/list`) will be created to list the contents of a directory.
- This endpoint will be restricted to the user's home directory for security.
- The API will support lazy loading of directory contents by accepting a `path` parameter.

### Backend: Sync Pair Management and Validation
- The database will be updated to store source-destination sync pairs.
- The `sync_jobs` table will have a foreign key to the `sync_pairs` table.
- Validation will be added to prevent the source and destination from being the same.
- A new API flow will be implemented for handling new sync pairs:
  1. The initial `/api/sync` request will check if the source-destination pair is new.
  2. If the pair is new, the backend will return a `409 Conflict` error.
  3. The frontend will catch this error and show a confirmation dialog.
  4. If the user confirms, the frontend will send a second request with a `force=true` parameter to create the new pair and proceed with the sync.

## Future Enhancements and UI/UX Polish

### Design System
- Update the design system to use a sans-serif font for a modern and beautiful aesthetic.

### Advanced File Picker UI
- Enhance the `FileExplorerModalComponent` to mimic a multi-column file explorer (like Mac Finder).
- Implement a single "Select" button at the top of the modal.
- Add a local filter input within the modal for improved user experience.

### Database Migration
- Implement a robust database migration script that runs automatically at server startup, eliminating the need for manual database file deletion.

### User Feedback and Alerts
- Implement a success/toast message to provide feedback after a successful synchronization.
- Replace all browser native alerts and popups with custom, polished UI components.

### Frontend Testing Strategy
Basic unit tests will be written for the Angular components to verify their internal logic. However, more comprehensive integration tests will be deferred to a later stage. These tests will be written using the `testing-library` to simulate real user interactions and ensure the components work correctly together.

## Phase 0: Outline & Research
1. **Extract unknowns from Technical Context** above:
   - Research performance goals for file synchronization tools.
   - Research typical constraints for desktop applications.
   - Define the initial scale and scope for version 1.0.

## Phase 1: Design & Contracts
*Prerequisites: research.md complete*

1. **Data Model**: The data model is defined in `apps/database/init.sql`.
2. **API Contracts**: Define the API endpoints for communication between the frontend and backend.
3. **Quickstart**: Create a `quickstart.md` guide for setting up and running the project.

## Phase 2: Task Planning Approach
*This section describes what the /tasks command will do*

**Task Generation Strategy**:
- Generate tasks based on the `plan.md` and `spec.md`.
- Create tasks for backend development (API endpoints, database integration, sync logic).
- Create tasks for frontend development (UI components, API service, state management).
- Create tasks for testing (unit and integration tests for both frontend and backend).

## Phase 3+: Future Implementation
*These phases are beyond the scope of the /plan command*

**Phase 3**: Task execution (`tasks.md`)
**Phase 4**: Implementation
**Phase 5**: Validation

## Progress Tracking
*This checklist is updated during execution flow*

**Phase Status**:
- [ ] Phase 0: Research complete
- [ ] Phase 1: Design complete
- [ ] Phase 2: Task planning complete
- [ ] Phase 3: Tasks generated
- [ ] Phase 4: Implementation complete
- [ ] Phase 5: Validation passed

**Gate Status**:
- [ ] Initial Constitution Check: PASS
- [ ] Post-Design Constitution Check: PASS
- [ ] All NEEDS CLARIFICATION resolved
- [ ] Complexity deviations documented

## Phase 1: Design & Contracts
*Prerequisites: research.md complete*

1. **Data Model**: The data model is defined in `apps/database/init.sql`.
2. **API Contracts**: Define the API endpoints for communication between the frontend and backend.
3. **Quickstart**: Create a `quickstart.md` guide for setting up and running the project.

## Phase 2: Task Planning Approach
*This section describes what the /tasks command will do*

**Task Generation Strategy**:
- Generate tasks based on the `plan.md` and `spec.md`.
- Create tasks for backend development (API endpoints, database integration, sync logic).
- Create tasks for frontend development (UI components, API service, state management).
- Create tasks for testing (unit and integration tests for both frontend and backend).

## Phase 3+: Future Implementation
*These phases are beyond the scope of the /plan command*

**Phase 3**: Task execution (`tasks.md`)
**Phase 4**: Implementation
**Phase 5**: Validation

## Progress Tracking
*This checklist is updated during execution flow*

**Phase Status**:
- [ ] Phase 0: Research complete
- [ ] Phase 1: Design complete
- [ ] Phase 2: Task planning complete
- [ ] Phase 3: Tasks generated
- [ ] Phase 4: Implementation complete
- [ ] Phase 5: Validation passed

**Gate Status**:
- [ ] Initial Constitution Check: PASS
- [ ] Post-Design Constitution Check: PASS
- [ ] All NEEDS CLARIFICATION resolved
- [ ] Complexity deviations documented
