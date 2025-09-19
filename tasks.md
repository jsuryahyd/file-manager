# Tasks: File Manager

**Input**: Design documents from `spec.md` and `plan.md`

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
- [x] T007 [P] Set up SQLite database schema in `apps/database/init.sql`.
- [x] T008 Integrate SQLite in the backend in `apps/backend/internal/db/db.go`.
- [x] T009 Implement sync logic with deduplication in `apps/backend/internal/fileops/fileops.go`.
- [x] T010 Update backend unit and integration tests.
- [x] T011 [P] Update the database schema in `apps/database/init.sql` to include the `sync_pairs` table.
- [x] T012 [P] Update `apps/backend/internal/db/db.go` to add functions for managing `sync_pairs`.
- [x] T013 Update `apps/backend/internal/fileops/fileops.go` to refactor `ListFiles` and add validation to `SyncUniqueFiles`.
- [x] T014 Update `apps/backend/cmd/main.go` to implement the new API logic for `/api/files/list` and `/api/sync`.
- [x] T015 Implement robust logging for the backend.
- [ ] T026 Implement database migration script that runs at server startup.

## Phase 3: Frontend Development
- [x] T016 Scaffold Angular app in `apps/frontend/file-manager-frontend/`.
- [x] T017 [P] Create SCSS design system with base styles and primitives in `apps/frontend/file-manager-frontend/src/styles/`.
- [x] T018 [P] Build Angular components for design primitives.
- [x] T019 Create the `FileExplorerModalComponent` in `apps/frontend/file-manager-frontend/src/app/file-explorer-modal/`.
- [x] T020 Update `apps/frontend/file-manager-frontend/src/app/sync/sync.component.ts` to use the new modal.
- [x] T021 Update `apps/frontend/file-manager-frontend/src/app/file-manager-api.service.ts` to add a method for the new `/api/files/list` endpoint and to handle the `409 Conflict` error from `/api/sync`.
- [x] T022 Add unit and integration tests for frontend.
- [ ] T027 Update design system to use sans-serif font.
- [ ] T028 Enhance `FileExplorerModalComponent` for multi-column view and local filter.
- [ ] T029 Implement single "Select" button on top of file explorer modal.
- [ ] T030 Implement success/toast message after sync.
- [ ] T031 Replace browser native alerts and popups with custom UI components.

## Phase 4: Polish
- [ ] T023 [P] Write usage and contribution guides.
- [ ] T024 [P] Polish UI/UX and improve performance.
- [ ] T025 [P] Ensure accessibility and responsiveness.

## Obsolete Tasks
- T014 Implement File Explorer screen with search and selection in `apps/frontend/file-manager-frontend/src/app/file-explorer/`.
- T016 Implement sync and history UI.
