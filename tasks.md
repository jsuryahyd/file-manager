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
- [x] T026 Implement database migration script that runs at server startup.

## Phase 3: Frontend Development
- [x] T016 Scaffold Angular app in `apps/frontend/file-manager-frontend/`.
- [x] T017 [P] Create SCSS design system with base styles and primitives in `apps/frontend/file-manager-frontend/src/styles/`.
- [x] T018 [P] Build Angular components for design primitives.
- [x] T019 Create the `FileExplorerModalComponent` in `apps/frontend/file-manager-frontend/src/app/file-explorer-modal/`.
- [x] T020 Update `apps/frontend/file-manager-frontend/src/app/sync/sync.component.ts` to use the new modal.
- [x] T021 Update `apps/frontend/file-manager-frontend/src/app/file-manager-api.service.ts` to add a method for the new `/api/files/list` endpoint and to handle the `409 Conflict` error from `/api/sync`.
- [x] T022 Add unit and integration tests for frontend.
- [x] T027 Update design system to use sans-serif font.
- [x] T028 Enhance `FileExplorerModalComponent` for multi-column view and local filter.
- [x] T028.1 Manually test the new FileExplorerModalComponent functionality.
- [ ] T028.2 Add debouncing to the path input field in FileExplorerModalComponent.
- [ ] T029 Implement single "Select" button on top of file explorer modal.
- [ ] T030 Implement success/toast message after sync.
- [ ] T031 Replace browser native alerts and popups with custom UI components.

## Phase 4: Polish
- [ ] T023 [P] Write usage and contribution guides.
- [ ] T024 [P] Polish UI/UX and improve performance.
- [ ] T025 [P] Ensure accessibility and responsiveness.

## Phase 5: Duplicate File Detection
- [ ] T032 [P] **Backend**: Create new API endpoint `/api/duplicates/find` to scan a directory for duplicate files.
- [ ] T033 [P] **Backend**: Implement duplicate detection logic using file hashes, metadata, and name patterns.
- [ ] T034 [P] **Backend**: Create new API endpoint `/api/duplicates/delete` to delete specified files.
- [ ] T035 [P] **Frontend**: Create a new component for the duplicate file detection UI.
- [ ] T036 [P] **Frontend**: Implement the UI to display duplicate file sets and allow users to select files for deletion.
- [ ] T037 [P] **Frontend**: Connect the UI to the backend API endpoints for finding and deleting duplicates.

## Phase 6: Sync Module Refactor
- [ ] T038 [P] **Backend**: Refactor sync functionality into a new `sync` module in `apps/backend/internal/sync/`.
- [ ] T039 [P] **Backend**: Implement recursive sync in the new `sync` module.
- [ ] T040 [P] **Backend**: Add option to skip path patterns during sync.
- [ ] T041 [P] **Backend**: Implement "peek mode" (dry run) for sync.
- [ ] T042 [P] **Frontend**: Add a checkbox to enable/disable recursive sync in `apps/frontend/file-manager-frontend/src/app/sync/sync.component.html`.
- [ ] T043 [P] **Frontend**: Add UI for managing skipped path patterns in `apps/frontend/file-manager-frontend/src/app/sync/sync.component.html`.
- [ ] T044 [P] **Frontend**: Add UI to trigger and display "peek mode" results in `apps/frontend/file-manager-frontend/src/app/sync/sync.component.html`.

## Phase 7: Sync Job Polling
- [ ] T045 [P] **Backend**: Add a new `SyncJob` struct and a `GetSyncJobs` function in `apps/backend/internal/db/db.go`.
- [ ] T046 [P] **Backend**: Create a new API endpoint `/api/sync/jobs` in `apps/backend/cmd/main.go` to return the last 10 sync jobs.
- [ ] T047 [P] **Backend**: Modify the `/api/sync` endpoint to be asynchronous by running the sync job in a goroutine.
- [ ] T048 [P] **Frontend**: Add a `getSyncJobs()` method to `apps/frontend/file-manager-frontend/src/app/file-manager-api.service.ts`.
- [ ] T049 [P] **Frontend**: Create a `sync-jobs-table` component to display the list of sync jobs.
- [ ] T050 [P] **Frontend**: Implement polling logic in the `sync` component to periodically refresh the job list.

## Phase 8: Sync Job Error Handling & Actions
- [ ] T051 [P] **Backend**: Create a new database migration file to add a `misc` JSON column to the `sync_jobs` table.
- [ ] T052 [P] **Backend**: Update `db.go` to modify `UpdateSyncJobStatus` to accept and store an error message.
- [ ] T053 [P] **Backend**: Update `sync.go` to save the error message to the database when a job fails.
- [ ] T054 [P] **Frontend**: Update the `SyncJob` interface to include the new `misc` field.
- [ ] T055 [P] **Frontend**: In the `sync-jobs-table` component, add a 3-dots menu to each row.
- [ ] T056 [P] **Frontend**: Implement the 'Show Fail Reason' menu item, visible only for failed jobs, to display the error in a modal.
- [ ] T057 [P] **Frontend**: Implement the 'Run sync again' menu item to emit an event with the job details.
- [ ] T058 [P] **Frontend**: Update the `sync` component to handle the 'Run sync again' event and pre-fill the form.

## Obsolete Tasks
- T014 Implement File Explorer screen with search and selection in `apps/frontend/file-manager-frontend/src/app/file-explorer/`.
- T016 Implement sync and history UI.