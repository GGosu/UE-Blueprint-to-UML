# 1.2.0 (05-03-2026)
## Features
- Node Subtitle: Each node now shows its Blueprint graph name (e.g. K2Node_CallFunction_7) as a small subtitle
- Unconnected Nodes: Nodes with no wired pins are now included in the diagram
- Scope-Aware Grouping: Unconnected nodes with a MemberScope are placed in the matching function subgraph

# 1.1.0 (02-03-2026)
## Features
- Graph Grouping: Disconnected execution chains are now grouped into Mermaid subgraphs
- Improved Subgraph Naming: Subgraphs are automatically named after the entry point (Event or Function Entry) they contain
- Custom Event Support: Added support for K2Node_CustomEvent and parsing of CustomFunctionName

## Fixes
- added missing port mapping docker-compose.yml to allow access to port 8080
- fixed loadConfig to correctly handle nested YAML sections
- added sensible default values for server port and body size limits in case they are missing from the config

# 1.0.1 (01-03-2026)
## Fixes
- fixed github button url
- fixed Dockerfile by adding go generate step, so templ files are built correctly in docker

## Changes
- Added live demo link to README
- improved docker-compose.yml

# 1.0.0 (01-03-2026)
## Features
- initial version!