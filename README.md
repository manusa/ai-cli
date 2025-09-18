# AI CLI

[![GitHub License](https://img.shields.io/github/license/manusa/ai-cli)](https://github.com/manusa/ai-cli/blob/main/LICENSE)
[![npm](https://img.shields.io/npm/v/npm-ai-cli)](https://www.npmjs.com/package/npm-ai-cli)
[![PyPI - Version](https://img.shields.io/pypi/v/python-ai-cli)](https://pypi.org/project/python-ai-cli/)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/manusa/ai-cli?sort=semver)](https://github.com/manusa/ai-cli/releases/latest)
[![Build](https://github.com/manusa/ai-cli/actions/workflows/build.yaml/badge.svg)](https://github.com/manusa/ai-cli/actions/workflows/build.yaml)

ai-cli is a command-line interface (CLI) tool that lets you go from zero to AI-powered in seconds in a safe, automated, and tailored way.

[✨ Features](#features) | [🚀 Getting Started](#getting-started) | [🎥 Demos](#demos)

## ✨ Features <a id="features"></a>

- **Policies**: set rules for AI interactions and tool usage
- **Discovery**: find and use **inference providers** and **tools** automatically
- **Extensible**: add plugins to extend functionality
- **Multi-model support**: support for different LLM inference providers (Google Gemini, LMStudio, Ollama, Ramalama) 

## 🚀 Getting Started <a id="getting-started"></a>

The quickest way to get started is by exposing an API key from [Google Gemini](https://aistudio.google.com/u/1/apikey) (`export GEMINI_API_KEY=$YOUR_KEY`) or by pulling one of the [supported Ollama models](https://github.com/manusa/ai-cli/blob/92b559c42f0743edbaefbcd7d8b695cc81adb5f0/pkg/inference/ollama/ollama.go#L25-L29). 

### Node environment available (npm)

If you have Node.js installed, you can run the CLI directly by using `npx`:

```bash
# Show the available commands
npx npm-ai-cli@latest help
# Start a TUI-based chat session
npx npm-ai-cli@latest chat
# Discover available tools and providers
npx npm-ai-cli@latest discover
```

### Python environment available (pip)

If you have Python installed, you can run the CLI directly by using `uvx`:

```bash
# Show the available commands
uvx python-ai-cli@latest help
# Start a TUI-based chat session
uvx python-ai-cli@latest chat
# Discover available tools and providers
uvx python-ai-cli@latest discover
```

### Manual installation

You can also install the CLI manually by downloading a binary compatible with your OS from the [latest release](https://github.com/manusa/ai-cli/releases/latest).

> [!NOTE]
> For macOS users: you might need to run `xattr -rc /path/to/ai-cli` to remove the quarantine attribute.
> We're still not signing the binaries, but it's on our roadmap.


## 🎥 Demos <a id="demos"></a>

### Chat TUI

<img alt="ai-cli chat demo" src="docs/demo.chat.gif" />

### Discovery

<img alt="ai-cli discover demo" src="docs/demo.discover.gif" />
