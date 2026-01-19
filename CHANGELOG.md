# Changelog

All notable changes to SQDesk will be documented in this file.

---

## v0.1.1 - January 20, 2026

### ✨ Added
- **Keywords Panel (F3)**: Toggle right-side panel showing SQL keywords and suggestions.
- **Auto-Completion Engine**: Context-aware suggestions as you type.
  - Keywords: SQL keywords (SELECT, FROM, WHERE, etc.).
  - Schema: Tables and columns from connected database.
  - History: Previously executed queries.
- **Help Popup (F4)**: Paginated keyboard shortcuts reference.
  - 6 categories: Navigation, Editor, Query, AI, Results, General.
  - Navigate pages with ← → arrows.

### 🚀 Improved
- Tab now replaces current word with suggestion instead of inserting.
- Keywords panel stays visible when switching focus between panes.
- Better status messages for panel toggle actions.

---

## v0.1.0 - January 19, 2026

### ✨ Added
- **Interactive TUI**: Modern and responsive terminal-based interface.
- **Multi-Database Support**: Support for MySQL, PostgreSQL, and SQLite.
- **AI Integration**: Text-to-SQL and SQL Refactoring features using AI (Gemini, Claude, OpenAI).
- **Connection Management**:
  - CRUD Connection (Create, Read, Update, Delete).
  - Test connection before saving.
  - Visual indicator for active connection.
- **Result Viewer**:
  - Query results displayed in table format.
  - Copy row/all data to clipboard.
  - Data visualization (Bar, Line, Pie Chart).
- **Sidebar Navigation**:
  - Visual indicators for active database and selected table.
  - Quick actions (Enter to select/query).

### 🐛 Fixed
- Fixed connection switching logic.
- Fixed settings form validation.
- Handled data types in the sidebar.

### 🚀 Improved
- **UX Enhancements**:
  - Panel navigation using `F1`/`F2`.
  - More informative connection validation status in the modal.
  - Removed setup wizard for faster startup.
- **Performance**: Optimized TUI rendering.

---

## Initial Release

### ✨ Added
- Basic TUI application structure.
- Basic database integration.
- SQL editor with syntax highlighting.
