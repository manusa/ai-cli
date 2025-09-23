# ai-cli - AI Agents Instructions

Always reference these instructions first and fallback to search or bash commands only when you encounter unexpected information that does not match the info here.

This file provides guidance to AI coding agents (GitHub Copilot, Claude Code, etc.) when working with code in this repository.

## Project Overview

`ai-cli` is a command-line interface (CLI) tool that enables users to interact with AI models and tools in a safe, automated, and tailored manner.
It supports multiple inference providers and a variety of tools, allowing users to create AI-powered workflows directly from the terminal.

The main difference between `ai-cli` and other AI tools is its support for policies, discovery of providers and tools, and extensibility through plugins.

By using policies, organization admins can set rules for AI interactions and tool usage, ensuring compliance and safety for enterprise users.
The discovery feature allows the CLI to automatically find and use available inference providers and tools based on the user's environment, making it easy to get started without extensive configuration.
The extensible architecture allows developers to add new plugins, tools, and providers, making `ai-cli` a flexible solution for various AI use cases.

## Development Commands

### Build and Development
```bash
make build          # Build the project (includes clean, tidy, format, lint)
make test           # Run all tests
make format         # Format Go code
make tidy           # Tidy Go modules
make lint           # Run golangci-lint
make clean          # Clean build artifacts
```

### Running the Application
```bash
./ai-cli help       # Show available commands after building
./ai-cli chat       # Start TUI-based chat session
./ai-cli discover   # Discover available tools and providers
./ai-cli version    # Show version information
```

### Testing
```bash
go test ./...                    # Run all tests
go test -v ./pkg/cmd/...         # Run tests for specific package with verbose output
go test -count=1 -v ./...        # Run tests without cache
```

### Multi-platform Builds
```bash
make build-all-platforms         # Build for all platforms (darwin, linux, windows on amd64, arm64)
```

## Architecture Overview

### Core Components

**AI Engine (`pkg/ai/`)**
- `ai.go`: Main AI interface using Eino framework for LLM interactions
- `session.go`: Session management for conversation history
- `tools.go`: Tool integration and execution
- Uses ReAct (Reasoning + Acting) agent pattern for tool-calling workflows

**Command Structure (`pkg/cmd/`)**
- `root.go`: Root command definition with three main subcommands
- `chat.go`: Interactive TUI chat interface using Bubble Tea
- `discover.go`: Discovery of available inference providers and tools
- `version.go`: Version information display

**Inference Providers (`pkg/inference/`)**
- `gemini/`: Google Gemini API integration
- `ollama/`: Local Ollama model support
- `lmstudio/`: LMStudio integration
- `ramalama/`: Ramalama support
- `discover.go`: Auto-discovery of available providers

**Tools System (`pkg/tools/`)**
- Modular tool architecture with providers for:
  - `github/`: GitHub operations
  - `kubernetes/`: Kubernetes cluster management
  - `postgresql/`: Database operations
  - `playwright/`: Browser automation
  - `browsers/`: Browser integration via MCP
  - `fs/`: Filesystem operations
- `discover.go`: Auto-discovery of available tools

**User Interface (`pkg/ui/`)**
- `model.go`: Bubble Tea application model
- `viewport.go`: Message display viewport
- `components/`: Reusable UI components (footer, scrollbar)
- `styles/`: Theme and styling definitions

**Configuration (`pkg/config/`)**
- `config.go`: Main configuration management
- `context.go`: Context-aware configuration
- Supports provider enabling/disabling and enforcement policies

**Policies (`pkg/policy/`)**
- Loading of policy files

### Key Technologies

- **Eino Framework**: AI workflow orchestration and tool calling
- **Bubble Tea v2**: Terminal user interface framework
- **Cobra**: CLI command structure
- **Model Context Protocol (MCP)**: External tool integration
- **Multiple LLM Providers**: Gemini, Ollama, LMStudio, Ramalama

### Data Flow

1. **Command Parsing**: Cobra handles CLI commands and flags
2. **Configuration Loading**: Auto-discover providers and tools based on environment
3. **AI Session**: Initialize AI engine with selected inference provider and tools
4. **TUI Interaction**: Bubble Tea manages chat interface and user interactions
5. **Tool Execution**: Eino orchestrates LLM responses and tool calls
6. **MCP Integration**: External tools via Model Context Protocol

### Testing Strategy

- Tests are implemented using Go's testing package and testify
- Test definitions are behavior-driven, focusing on expected outcomes
- Mocking and stubbing are reduced to the minimum necessary (black-box testing) to allow flexibility in implementation changes
- Coverage should focus on behavior rather than implementation details or percentage metrics
- Integration of different areas is encouraged as long as tests are quick and based on specific behaviors or user journeys

### Release Process

- Semantic versioning (currently in PoC mode with patch releases)
- GitHub Actions for build and release automation
- Multi-platform binary distribution (npm, pip, go install, manual)
- GitHub UI-based release creation with auto-generated notes

### Notable Patterns

- **Provider Pattern**: Consistent interface for inference providers and tools
- **Discovery Pattern**: Auto-detection of available services and tools
- **Plugin Architecture**: Extensible tool system via MCP and native providers
- **Policy Enforcement**: Configurable rules for AI interactions and tool usage
