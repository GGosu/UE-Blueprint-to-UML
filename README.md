# UE Blueprint to UML

[![CI](https://github.com/GGosu/UE-Blueprint-to-UML/actions/workflows/ci.yml/badge.svg)](https://github.com/GGosu/UE-Blueprint-to-UML/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/GGosu/UE-Blueprint-to-UML/graph/badge.svg)](https://codecov.io/gh/GGosu/UE-Blueprint-to-UML)

Convert Unreal Engine Blueprint nodes into Mermaid diagrams.

**Live demo:** https://ue-uml.temper.rocks/

## Why?

Pasting Blueprint nodes raw into an LLM works, but the clipboard format is proprietary and full of node coordinates, GUIDs and pin hashes. A complex graph can produce thousands of lines of noise, burns tokens, and models don't know the format well.

This tool converts that into a clean Mermaid flowchart with just the logic.

1. Select nodes in UE Editor and `Ctrl+C`
2. `Ctrl+V` on the page
3. Copy the Mermaid source or save the SVG

## Quick Start

```bash
# Generate templates (requires templ CLI)
templ generate

# Run the web server
go run .
```

## Usage

| Action              | How                                             |
|---------------------|-------------------------------------------------|
| Convert Blueprint   | `Ctrl+V` anywhere on the page                   |
| Copy Mermaid source | `Ctrl+C` (when nothing is selected)             |
| Save as SVG         | Click **Save SVG**                              |
| Navigation          | Scroll to zoom, drag to pan, click **⊙** to fit |

## Stack

Go, templ, Mermaid.js, Docker
