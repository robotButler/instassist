# Changelog

All notable changes to instassist will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-12-06

### Added
- Initial release of instassist
- Interactive TUI mode for getting AI-powered command suggestions
- Support for both `codex` and `claude` AI CLIs
- Beautiful, color-coded interface with lipgloss styling
- Keyboard shortcuts for navigation and selection
- **Ctrl+Enter** to execute selected command directly
- **Enter** to copy selected command to clipboard
- Non-interactive CLI mode with flags:
  - `-prompt` for specifying prompt
  - `-select` for auto-selecting option by index
  - `-output` for choosing output mode (clipboard/stdout/exec)
  - `-cli` for choosing AI CLI (codex/claude)
  - `-version` for version information
- Stdin support for piping prompts
- Multiple output modes: clipboard, stdout, exec
- Tab key to switch between codex and claude
- Vim-style navigation (j/k keys)
- Structured JSON schema for consistent AI responses
- Makefile for easy building and installation
- Installation script (install.sh) for systems without Make
- Comprehensive README with examples and desktop integration guides
- MIT License

### UI Enhancements
- Color-coded sections (prompts, options, status)
- Rounded borders around options and input
- Selected option highlighted with background color
- Visual icons and emojis for better readability
- Status indicators with appropriate colors
- Divider lines for visual separation

### Features
- Schema file lookup in multiple locations
- Clean, minimal interface optimized for popup usage
- Fast response times
- Error handling with helpful messages
- Raw output display when JSON parsing fails

[1.0.0]: https://github.com/robotButler/instassist/releases/tag/v1.0.0
