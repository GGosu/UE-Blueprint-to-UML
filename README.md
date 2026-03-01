# UE Blueprint to UML

Convert Unreal Engine Blueprint nodes into Mermaid diagrams.

## Why?

LLMs can't read binary `.uasset` files. This tool converts copied blueprint nodes into Mermaid flowcharts so you can
paste them into AI chats or use them for documentation.

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
