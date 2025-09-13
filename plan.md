# File Manager

## Implementation Plan

### Project Goals
- Build a local file backup tool in Golang, supporting flexible file selection, multiple destinations, and reliable deduplication.
- Develop a modern Angular web GUI for file management and sync operations.
- Use SQLite for metadata storage and DDD for Go backend structure.
- Plan for future features: cloud sync, file analysis/classification, OS integration, and graph-based UI.

### Development Workflow
- Use VS Code as the primary IDE.
- Follow commit-by-commit implementation for clear progress tracking.
- Use Git for version control.

### Recommended Tools & Extensions
- **GitHub Copilot**: AI-powered code completion and suggestions.
- **Copilot Chat**: Interactive AI assistant for code review and queries.
- **Prettier**: Code formatter for consistent style (Angular).
- **ESLint**: Linting for JavaScript/TypeScript projects.
- **Go**: Official Go extension for VS Code.
- **Angular Language Service**: Enhanced Angular development.
- **Better Comments**: Enhanced code commenting.
- **Project Manager**: Quick project switching.
- **Path Intellisense**: Autocompletion for file paths.

### Initial Setup (Commit 1)
- Initialize monorepo structure using a recommended tool (e.g., Nx, Turborepo, or Go workspace).
- Add `.gitignore`, `README.md`, and license files.
- Set up basic folder structure for Go backend, Angular frontend, and database.

### Planned Commits
1. **Monorepo Initialization**: Set up monorepo, add essential files.
2. **Tooling Setup**: Install and configure recommended extensions and tools.
3. **Backend Core Logic**: Implement basic file operations and sync logic in Go.
4. **Frontend Setup**: Scaffold Angular app, connect to backend API.
5. **Database Integration**: Set up SQLite and basic schema for metadata.
6. **Deduplication Logic**: Implement reliable deduplication in sync process.
7. **Testing & Validation**: Add unit and integration tests for backend and frontend.
8. **Documentation**: Write usage and contribution guides.
9. **Feature Expansion**: Add advanced features (cloud sync, file analysis, graph UI, OS integration).
10. **Refinement & Optimization**: Improve performance, fix bugs, polish UI.

### Next Steps
- Confirm project requirements and technology stack.
- Begin with monorepo initialization commit.
- Follow the plan for incremental development and commit-by-commit progress.

---
*This plan will be updated as the project evolves. Suggestions for tools and extensions will be revisited as needed.*

### PRD: 
Help me build a local file backup tool in golang, between hard drives. I want to start simple and add features in phases. At the end, it will have a web GUI to manage, schedule, generate reports etc. It would have most flexibility in choosing which files (support path patterns and ignore patterns etc) are backed up, and to select multiple destinations at once to sync to. Later we will also add api support for any supported cloud providers - so that we drag & drop, and files are sent to cloud provider. I also want to have a view with an intuitive graph like UI that shows all drives that a file is present at.
- Later I also want to run analysis on each file and summarize the file content, classify the file and add that to metadata, so that I can organize files based on this classification. Suggest good tools for this - do we need an AI for this?
- I also want to add some of these features to native file manager on the Operating system, i.e each file context menu.

But first let us build a simple web gui that lists system  folders and files and option to target a folder on other drive, where the files will be synced to. Files will not be duplicated, only new files are added. Deduping must be reliable. 

Use Modern Angular for UI. Golang as main language, sqlite for database. Make a Developer friendly setup using monorepo with best monorepo management tool, while also keeping it simple.
Use DDD in golang app, as we support multiple modules and domains.