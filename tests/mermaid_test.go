package tests

import (
	"os"
	"strings"
	"testing"

	"UE_UML/internal/blueprint"
)

func TestGenerateMermaidHeader(t *testing.T) {
	g := blueprint.Graph{}
	out := blueprint.GenerateMermaid(g)
	if !strings.Contains(out, "flowchart LR") {
		t.Errorf("expected output to contain 'flowchart LR', got: %q", out)
	}
	if !strings.Contains(out, "curve") {
		t.Errorf("expected init directive with curve setting, got: %q", out)
	}
}

func TestGenerateMermaidNodeShapes(t *testing.T) {
	tests := []struct {
		name     string
		kind     blueprint.NodeKind
		id       string
		label    string
		contains string
	}{
		{"Entry", blueprint.KindEntry, "MyEntry", "myFunc", `MyEntry(["myFunc"])`},
		{"Event", blueprint.KindEvent, "MyEvent", "OnTick", `MyEvent(["OnTick"])`},
		{"Branch", blueprint.KindBranch, "MyBranch", "Branch", `MyBranch{"Branch"}`},
		{"Default", blueprint.KindDefault, "MyNode", "doStuff", `MyNode["doStuff"]`},
		{"Variable", blueprint.KindVariable, "MyVar", "hitActor", `MyVar(["hitActor"]):::variable`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := blueprint.Graph{
				Nodes: []blueprint.Node{{ID: tc.id, Kind: tc.kind, Label: tc.label}},
			}
			out := blueprint.GenerateMermaid(g)
			if !strings.Contains(out, tc.contains) {
				t.Errorf("expected output to contain %q\ngot:\n%s", tc.contains, out)
			}
		})
	}
}

func TestGenerateMermaidEdgePlain(t *testing.T) {
	g := blueprint.Graph{
		Nodes: []blueprint.Node{
			{ID: "A", Kind: blueprint.KindDefault, Label: "NodeA"},
			{ID: "B", Kind: blueprint.KindDefault, Label: "NodeB"},
		},
		Edges: []blueprint.Edge{{From: "A", To: "B", Label: ""}},
	}
	out := blueprint.GenerateMermaid(g)
	if !strings.Contains(out, "A --> B") {
		t.Errorf("expected 'A --> B', got:\n%s", out)
	}
}

func TestGenerateMermaidEdgeLabeled(t *testing.T) {
	g := blueprint.Graph{
		Nodes: []blueprint.Node{
			{ID: "A", Kind: blueprint.KindBranch, Label: "Branch"},
			{ID: "B", Kind: blueprint.KindDefault, Label: "TrueNode"},
		},
		Edges: []blueprint.Edge{{From: "A", To: "B", Label: "true"}},
	}
	out := blueprint.GenerateMermaid(g)
	if !strings.Contains(out, `A -- "true" --> B`) {
		t.Errorf("expected 'A -- \"true\" --> B', got:\n%s", out)
	}
}

func TestGenerateMermaidLabelSanitization(t *testing.T) {
	t.Run("QuotesEscaped", func(t *testing.T) {
		g := blueprint.Graph{
			Nodes: []blueprint.Node{{ID: "X", Kind: blueprint.KindDefault, Label: `say "hello"`}},
		}
		out := blueprint.GenerateMermaid(g)
		if strings.Contains(out, `"hello"`) {
			t.Errorf("raw quotes should be escaped, got:\n%s", out)
		}
		if !strings.Contains(out, "#quot;") {
			t.Errorf("expected #quot; in output, got:\n%s", out)
		}
	})

	t.Run("LongLabelTruncated", func(t *testing.T) {
		longLabel := "ThisIsAVeryLongLabelThatExceedsFortyCharactersDefinitely"
		g := blueprint.Graph{
			Nodes: []blueprint.Node{{ID: "Y", Kind: blueprint.KindDefault, Label: longLabel}},
		}
		out := blueprint.GenerateMermaid(g)
		if strings.Contains(out, longLabel) {
			t.Errorf("long label should be truncated, got:\n%s", out)
		}
		if !strings.Contains(out, "...") {
			t.Errorf("expected '...' in truncated label, got:\n%s", out)
		}
	})
}

func TestGenerateMermaidDataEdge(t *testing.T) {
	g := blueprint.Graph{
		Nodes: []blueprint.Node{
			{ID: "A", Kind: blueprint.KindDefault, Label: "GetCamera"},
			{ID: "B", Kind: blueprint.KindDefault, Label: "Trace"},
		},
		Edges: []blueprint.Edge{
			{From: "A", To: "B", Label: "ReturnValue", Kind: blueprint.DataEdge},
		},
	}
	out := blueprint.GenerateMermaid(g)
	if !strings.Contains(out, `A -. "ReturnValue" .-> B`) {
		t.Errorf("expected dashed data edge, got:\n%s", out)
	}
}

func TestGenerateMermaidRoundTrip(t *testing.T) {
	data, err := os.ReadFile("fixtures/case1.txt")
	if err != nil {
		t.Fatalf("cannot read fixture: %v", err)
	}
	g, err := blueprint.ParseBlueprint(string(data))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	out := blueprint.GenerateMermaid(g)

	if !strings.Contains(out, "flowchart LR") {
		t.Error("expected flowchart LR header")
	}
	if !strings.Contains(out, "{") {
		t.Error("expected at least one diamond shape in output")
	}
	if !strings.Contains(out, `(["`) {
		t.Error("expected at least one pill shape in output")
	}
}
