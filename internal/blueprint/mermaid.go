package blueprint

import (
	"fmt"
	"strings"
)

const maxLabelLen = 40

func GenerateMermaid(g Graph) string {
	var sb strings.Builder

	sb.WriteString("%%{init: {'theme': 'dark', 'flowchart': {'curve': 'linear'}}}%%\n")
	sb.WriteString("flowchart LR\n")

	// Node colour classes — override theme defaults per kind.
	sb.WriteString("    classDef entry    fill:#143d6b,stroke:#1e6bb8,color:#b3d9ff\n")
	sb.WriteString("    classDef variable fill:#12422a,stroke:#1c7a44,color:#85e5b0\n")
	sb.WriteString("    classDef branch   fill:#4a2a0a,stroke:#9f6020,color:#ffcc80\n")

	components := findComponents(g)
	nodeMap := make(map[string]Node)
	for _, n := range g.Nodes {
		nodeMap[n.ID] = n
	}

	for i, comp := range components {
		subgraphName := fmt.Sprintf("Graph %d", i+1)
		for _, nodeID := range comp {
			n := nodeMap[nodeID]
			if n.Kind == KindEvent || n.Kind == KindEntry {
				subgraphName = n.Label
				break
			}
		}

		fmt.Fprintf(&sb, "    subgraph %s [\"%s\"]\n", fmt.Sprintf("sg_%d", i), sanitizeLabel(subgraphName))
		for _, nodeID := range comp {
			n := nodeMap[nodeID]
			label := sanitizeLabel(n.Label)
			var nodeDef string
			switch n.Kind {
			case KindEntry, KindEvent:
				nodeDef = fmt.Sprintf("        %s([\"%s\"]):::entry\n", n.ID, label)
			case KindBranch:
				nodeDef = fmt.Sprintf("        %s{\"%s\"}:::branch\n", n.ID, label)
			case KindVariable:
				nodeDef = fmt.Sprintf("        %s([\"%s\"]):::variable\n", n.ID, label)
			default:
				nodeDef = fmt.Sprintf("        %s[\"%s\"]\n", n.ID, label)
			}
			sb.WriteString(nodeDef)
		}
		sb.WriteString("    end\n")
	}

	for _, e := range g.Edges {
		label := sanitizeLabel(e.Label)
		switch e.Kind {
		case ExecEdge:
			if label == "" {
				fmt.Fprintf(&sb, "    %s --> %s\n", e.From, e.To)
			} else {
				fmt.Fprintf(&sb, "    %s -- \"%s\" --> %s\n", e.From, label, e.To)
			}
		case DataEdge:
			if label == "" {
				fmt.Fprintf(&sb, "    %s -.-> %s\n", e.From, e.To)
			} else {
				fmt.Fprintf(&sb, "    %s -. \"%s\" .-> %s\n", e.From, label, e.To)
			}
		}
	}

	return sb.String()
}

func findComponents(g Graph) [][]string {
	adj := make(map[string][]string)
	for _, e := range g.Edges {
		adj[e.From] = append(adj[e.From], e.To)
		adj[e.To] = append(adj[e.To], e.From)
	}

	visited := make(map[string]bool)
	var components [][]string

	// Sort nodes by their original order to keep output deterministic
	for _, n := range g.Nodes {
		if !visited[n.ID] {
			var component []string
			stack := []string{n.ID}
			visited[n.ID] = true
			for len(stack) > 0 {
				curr := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				component = append(component, curr)
				for _, neighbor := range adj[curr] {
					if !visited[neighbor] {
						visited[neighbor] = true
						stack = append(stack, neighbor)
					}
				}
			}
			components = append(components, component)
		}
	}
	return components
}

func sanitizeLabel(label string) string {
	label = strings.ReplaceAll(label, `"`, "#quot;")
	if len([]rune(label)) > maxLabelLen {
		runes := []rune(label)
		label = string(runes[:maxLabelLen]) + "..."
	}
	return label
}
