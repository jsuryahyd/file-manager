# Tasks: File Manager

**Input**: Design documents from the root directory.
**Prerequisites**: plan.md (required)

## Execution Flow (main)
1. Load plan.md from the root directory.
2. Generate tasks by category based on the plan.
3. Apply task rules for parallel execution.
4. Number tasks sequentially.
5. Create this tasks.md file.

## Format: `[ID] [P?] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- Include exact file paths in descriptions

## Path Conventions
- **Backend**: `apps/backend/`
- **Frontend**: `apps/frontend/file-manager-frontend/`
- **Database**: `apps/database/`

## Phase 1: Setup
- [x] T001 [P] Initialize monorepo structure.
- [x] T002 [P] Add `.gitignore`, `README.md`, and `LICENSE` files.
- [x] T003 [P] Set up basic folder structure for Go backend, Angular frontend, and database.
- [x] T004 [P] Configure linting and formatting tools (ESLint, Prettier for frontend; Go linting for backend).

## Phase 2: Backend Development
- [x] T005 Initialize Go project in `apps/backend/`.
- [x] T006 [P] Implement basic file listing API endpoint in `apps/backend/internal/fileops/explorer.go`.
- [ ] T007 [P] Set up SQLite database schema in `apps/database/init.sql`.
- [ ] T008 Integrate SQLite in the backend in `apps/backend/internal/db/db.go`.
- [ ] T009 Implement sync logic with deduplication in `apps/backend/internal/fileops/fileops.go`.
- [ ] T010 Add unit and integration tests for backend.

## Phase 3: Frontend Development
- [x] T011 Scaffold Angular app in `apps/frontend/file-manager-frontend/`.
- [x] T012 [P] Create SCSS design system with base styles and primitives in `apps/frontend/file-manager-frontend/src/styles/`.
- [x] T013 [P] Build Angular components for design primitives.
- [x] T014 Implement File Explorer screen with search and selection in `apps/frontend/file-manager-frontend/src/app/file-explorer/`.
- [ ] T015 Connect frontend to backend API in `apps/frontend/file-manager-frontend/src/app/file-manager-api.service.ts`.
- [ ] T016 Implement sync and history UI.
- [ ] T017 Add unit and integration tests for frontend.

## Phase 4: Polish
- [ ] T018 [P] Write usage and contribution guides.
- [ ] T019 [P] Polish UI/UX and improve performance.
- [ ] T020 [P] Ensure accessibility and responsiveness.

## Dependencies
- Backend API (T006) before frontend integration (T015).
- Database schema (T007) before database integration (T008).

## Parallel Example
```
# Launch T001-T004 together:
Task: "Initialize monorepo structure."
Task: "Add .gitignore, README.md, and LICENSE files."
Task: "Set up basic folder structure for Go backend, Angular frontend, and database."
Task: "Configure linting and formatting tools (ESLint, Prettier for frontend; Go linting for backend)."
```