# Feature Specification: File Manager Core Functionality

**Feature Branch**: `[###-feature-name]`  
**Created**: 2025-09-19
**Status**: Draft  
**Input**: User description: "Build a local file backup tool in golang, between hard drives. I want to start simple and add features in phases. At the end, it will have a web GUI to manage, schedule, generate reports etc. It would have most flexibility in choosing which files (support path patterns and ignore patterns etc) are backed up, and to select multiple destinations at once to sync to."

---

## ⚡ Quick Guidelines
- ✅ Focus on WHAT users need and WHY
- ❌ Avoid HOW to implement (no tech stack, APIs, code structure)
- 👥 Written for business stakeholders, not developers

---

## User Scenarios & Testing *(mandatory)*

### Primary User Story
A user wants to reliably back up files from a local folder to another local folder. The system should be smart enough to avoid creating duplicate files, only copying what's new or changed. The user needs a simple web interface to select folders and start the backup process.

### Acceptance Scenarios
1. **Given** a source folder with files and an empty destination folder, **When** the user initiates a sync, **Then** all files from the source are copied to the destination.
2. **Given** a source folder and a destination folder that already contains identical versions of some files, **When** the user initiates a sync, **Then** only the files that are new or have been modified in the source are copied to the destination.

### Edge Cases
- What happens when the destination drive runs out of space during a sync?
- How does the system handle file deletion in the source folder? [NEEDS CLARIFICATION: Should deletions be mirrored to the destination?]

## Requirements *(mandatory)*

### Functional Requirements
- **FR-001**: The system MUST allow a user to select a local directory as a synchronization source.
- **FR-002**: The system MUST allow a user to select a local directory as a synchronization destination.
- **FR-003**: The system MUST copy new or modified files from the source to the destination directory.
- **FR-004**: The system MUST use a reliable method (e.g., file content hashing) to determine if a file is a duplicate and avoid re-copying it.
- **FR-005**: The system MUST provide a web interface to list files and folders.
- **FR-006**: The system MUST store metadata about files and sync operations in a database.
- **FR-007**: The system MUST resolve file name collisions by skipping the file if it already exists in the destination.

### API Endpoints
- `GET /api/files?path=<directory_path>`: Lists files and folders in the given directory.
- `POST /api/sync`: Initiates a synchronization job. The request body should contain the source and destination directories.
  - Request Body: `{ "source": "/path/to/source", "destination": "/path/to/destination" }`

### Key Entities *(include if feature involves data)*
- **File**: Represents a single file on the filesystem. Key attributes include its full path, size, modification date, and a content hash.
- **Sync Job**: Represents a single, user-initiated synchronization task from one directory to another. Key attributes include the source path, destination path, status (e.g., pending, in-progress, completed, failed), and timestamps.

---

## Review & Acceptance Checklist
*GATE: Automated checks run during main() execution*

### Content Quality
- [ ] No implementation details (languages, frameworks, APIs)
- [ ] Focused on user value and business needs
- [ ] Written for non-technical stakeholders
- [ ] All mandatory sections completed

### Requirement Completeness
- [ ] No [NEEDS CLARIFICATION] markers remain
- [ ] Requirements are testable and unambiguous  
- [ ] Success criteria are measurable
- [ ] Scope is clearly bounded
- [ ] Dependencies and assumptions identified
