package blueprint

import (
	"regexp"
	"strings"
)

type NodeKind int

const (
	KindDefault NodeKind = iota
	KindEntry
	KindEvent
	KindBranch
	KindVariable
)

type EdgeKind int

const (
	ExecEdge EdgeKind = iota // solid arrow  -->
	DataEdge                 // dashed arrow -.->
)

type Node struct {
	ID    string
	Kind  NodeKind
	Label string
}

type Edge struct {
	From  string
	To    string
	Label string
	Kind  EdgeKind
}

type Graph struct {
	Nodes []Node
	Edges []Edge
}

var (
	reBeginObject     = regexp.MustCompile(`^Begin Object Class=/Script/[^.]+\.(\w+) Name="(\w+)"`)
	reFunctionRef     = regexp.MustCompile(`FunctionReference=\([^)]*MemberName="([^"]+)"`)
	reEventRef        = regexp.MustCompile(`EventReference=\([^)]*MemberName="([^"]+)"`)
	reVariableRef     = regexp.MustCompile(`VariableReference=\([^)]*MemberName="([^"]+)"`)
	reCustomEventName = regexp.MustCompile(`CustomFunctionName="([^"]+)"`)
	rePinId           = regexp.MustCompile(`\bPinId=([A-F0-9]{32})\b`)
	rePinName         = regexp.MustCompile(`\bPinName="([^"]*)"`)
	rePinCategory     = regexp.MustCompile(`PinType\.PinCategory="([^"]*)"`)
	rePinDirection    = regexp.MustCompile(`\bDirection="([^"]*)"`)
	rePinHidden       = regexp.MustCompile(`\bbHidden=(True|False)\b`)
	rePinFriendly     = regexp.MustCompile(`PinFriendlyName=NSLOCTEXT\("[^"]*",\s*"[^"]*",\s*"([^"]+)"\)`)
	reLinkedTo        = regexp.MustCompile(`LinkedTo=\(([^)]*)\)`)
	reLinkedRef       = regexp.MustCompile(`(\w+)\s+([A-F0-9]{32})`)
)

type linkedRef struct {
	nodeName string
	pinID    string
}

type pinInfo struct {
	pinID       string
	pinName     string
	pinFriendly string
	isExec      bool
	isOutput    bool
	isHidden    bool
	edgeLabel   string
	linkedTo    []linkedRef
}

type nodeState struct {
	node    Node
	varName string
	pins    []pinInfo
}

func ParseBlueprint(text string) (Graph, error) {
	lines := strings.Split(text, "\n")

	var nodes []nodeState
	nodeIndex := map[string]int{}
	currentIdx := -1
	pinIDToName := map[string]string{}

	var currentPin *pinInfo

	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)

		if m := reBeginObject.FindStringSubmatch(line); m != nil {
			className := m[1]
			nodeID := m[2]

			ns := nodeState{}
			ns.node.ID = nodeID
			ns.node.Kind = classToKind(className)

			currentIdx = len(nodes)
			nodes = append(nodes, ns)
			nodeIndex[nodeID] = currentIdx
			currentPin = nil
			continue
		}

		if strings.HasPrefix(line, "End Object") {
			currentIdx = -1
			currentPin = nil
			continue
		}

		if currentIdx < 0 {
			continue
		}

		ns := &nodes[currentIdx]

		if m := reFunctionRef.FindStringSubmatch(line); m != nil {
			if ns.node.Label == "" {
				ns.node.Label = m[1]
			}
			continue
		}

		if m := reEventRef.FindStringSubmatch(line); m != nil {
			if ns.node.Label == "" {
				ns.node.Label = m[1]
			}
			continue
		}

		if m := reCustomEventName.FindStringSubmatch(line); m != nil {
			if ns.node.Label == "" {
				ns.node.Label = m[1]
			}
			continue
		}

		if m := reVariableRef.FindStringSubmatch(line); m != nil {
			if ns.varName == "" {
				ns.varName = m[1]
			}
			continue
		}

		if strings.Contains(line, "CustomProperties Pin (") {
			pin := parsePinLine(line)
			ns.pins = append(ns.pins, pin)
			currentPin = &ns.pins[len(ns.pins)-1]
			if pin.pinID != "" {
				name := pin.pinName
				if pin.pinFriendly != "" && pin.pinFriendly != "true" && pin.pinFriendly != "false" {
					name = pin.pinFriendly
				}
				if name != "" {
					pinIDToName[pin.pinID] = name
				}
			}
			continue
		}

		_ = currentPin
	}

	// Phase 2: label fallback
	for i := range nodes {
		ns := &nodes[i]
		if strings.HasPrefix(ns.node.ID, "K2Node_FunctionResult") {
			ns.node.Label = "Return Node"
			continue
		}
		if ns.node.Label == "" {
			className := kindToClassSuffix(ns.node)
			switch {
			case strings.HasPrefix(ns.node.ID, "K2Node_VariableSet"):
				ns.node.Label = "Set " + ns.varName
			case strings.HasPrefix(ns.node.ID, "K2Node_VariableGet"):
				ns.node.Label = ns.varName // oval shape is self-explanatory — no "Get " prefix
			default:
				ns.node.Label = strings.TrimPrefix(className, "K2Node_")
			}
		}
	}

	// Phase 3: build edges from all output pins.
	seen := map[string]bool{}
	var edges []Edge

	for _, ns := range nodes {
		for _, pin := range ns.pins {
			if !pin.isOutput {
				continue
			}
			if !pin.isExec && pin.isHidden {
				continue
			}

			var kind EdgeKind
			if pin.isExec {
				kind = ExecEdge
			} else {
				kind = DataEdge
			}
			kindStr := "exec"
			if kind == DataEdge {
				kindStr = "data"
			}

			for _, ref := range pin.linkedTo {
				if ref.nodeName == ns.node.ID {
					continue
				}

				var label string
				if pin.isExec {
					label = pin.edgeLabel
				} else {
					if !strings.HasPrefix(ns.node.ID, "K2Node_VariableGet") &&
						pin.pinName != "ReturnValue" {
						label = pin.pinName
					} else if pin.pinName == "ReturnValue" &&
						strings.HasPrefix(ref.nodeName, "K2Node_FunctionResult") {
						label = pinIDToName[ref.pinID]
					}
				}

				key := ns.node.ID + "|" + ref.nodeName + "|" + label + "|" + kindStr
				if seen[key] {
					continue
				}
				seen[key] = true
				edges = append(edges, Edge{
					From:  ns.node.ID,
					To:    ref.nodeName,
					Label: label,
					Kind:  kind,
				})
			}
		}
	}

	// Phase 4: keep only nodes that appear in at least one edge.
	connected := map[string]bool{}
	for _, e := range edges {
		connected[e.From] = true
		connected[e.To] = true
	}

	var filteredNodes []Node
	for _, ns := range nodes {
		if connected[ns.node.ID] {
			filteredNodes = append(filteredNodes, ns.node)
		}
	}

	return Graph{Nodes: filteredNodes, Edges: edges}, nil
}

func classToKind(className string) NodeKind {
	switch className {
	case "K2Node_FunctionEntry":
		return KindEntry
	case "K2Node_Event", "K2Node_CustomEvent":
		return KindEvent
	case "K2Node_IfThenElse":
		return KindBranch
	case "K2Node_VariableGet", "K2Node_VariableSet":
		return KindVariable
	default:
		return KindDefault
	}
}

func kindToClassSuffix(n Node) string {
	parts := strings.Split(n.ID, "_")
	if len(parts) >= 2 {
		last := parts[len(parts)-1]
		allDigits := true
		for _, c := range last {
			if c < '0' || c > '9' {
				allDigits = false
				break
			}
		}
		if allDigits && len(parts) > 1 {
			return strings.Join(parts[:len(parts)-1], "_")
		}
	}
	return n.ID
}

func parsePinLine(line string) pinInfo {
	var pin pinInfo

	if m := rePinId.FindStringSubmatch(line); m != nil {
		pin.pinID = m[1]
	}
	if m := rePinName.FindStringSubmatch(line); m != nil {
		pin.pinName = m[1]
	}
	if m := rePinCategory.FindStringSubmatch(line); m != nil {
		pin.isExec = m[1] == "exec"
	}
	if m := rePinDirection.FindStringSubmatch(line); m != nil {
		pin.isOutput = m[1] == "EGPD_Output"
	}
	if m := rePinHidden.FindStringSubmatch(line); m != nil {
		pin.isHidden = m[1] == "True"
	}
	if m := rePinFriendly.FindStringSubmatch(line); m != nil {
		label := m[1]
		pin.pinFriendly = label
		if label == "true" || label == "false" {
			pin.edgeLabel = label
		}
	}
	if m := reLinkedTo.FindStringSubmatch(line); m != nil {
		refs := reLinkedRef.FindAllStringSubmatch(m[1], -1)
		for _, r := range refs {
			pin.linkedTo = append(pin.linkedTo, linkedRef{nodeName: r[1], pinID: r[2]})
		}
	}

	return pin
}
