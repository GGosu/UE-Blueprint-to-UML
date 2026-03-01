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

	for _, n := range g.Nodes {
		label := sanitizeLabel(n.Label)
		var nodeDef string
		switch n.Kind {
		case KindEntry, KindEvent:
			// pill shape + blue
			nodeDef = fmt.Sprintf("    %s([\"%s\"]):::entry\n", n.ID, label)
		case KindBranch:
			// diamond + amber
			nodeDef = fmt.Sprintf("    %s{\"%s\"}:::branch\n", n.ID, label)
		case KindVariable:
			// pill shape + green  (same pill as entry, different colour = clearly oval)
			nodeDef = fmt.Sprintf("    %s([\"%s\"]):::variable\n", n.ID, label)
		default:
			// rectangle, theme default colour
			nodeDef = fmt.Sprintf("    %s[\"%s\"]\n", n.ID, label)
		}
		sb.WriteString(nodeDef)
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

func sanitizeLabel(label string) string {
	label = strings.ReplaceAll(label, `"`, "#quot;")
	if len([]rune(label)) > maxLabelLen {
		runes := []rune(label)
		label = string(runes[:maxLabelLen]) + "..."
	}
	return label
}
