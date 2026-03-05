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
		{"Entry", blueprint.KindEntry, "MyEntry", "myFunc", `MyEntry(["myFunc<br/><small>MyEntry</small>"])`},
		{"Event", blueprint.KindEvent, "MyEvent", "OnTick", `MyEvent(["OnTick<br/><small>MyEvent</small>"])`},
		{"Branch", blueprint.KindBranch, "MyBranch", "Branch", `MyBranch{"Branch<br/><small>MyBranch</small>"}`},
		{"Default", blueprint.KindDefault, "MyNode", "doStuff", `MyNode["doStuff<br/><small>MyNode</small>"]`},
		{"Variable", blueprint.KindVariable, "MyVar", "hitActor", `MyVar(["hitActor<br/><small>MyVar</small>"]):::variable`},
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

func TestGenerateMermaidSubgraphs(t *testing.T) {
	g := blueprint.Graph{
		Nodes: []blueprint.Node{
			{ID: "E1", Kind: blueprint.KindEvent, Label: "Event1"},
			{ID: "N1", Kind: blueprint.KindDefault, Label: "Node1"},
			{ID: "E2", Kind: blueprint.KindEvent, Label: "Event2"},
			{ID: "N2", Kind: blueprint.KindDefault, Label: "Node2"},
		},
		Edges: []blueprint.Edge{
			{From: "E1", To: "N1", Kind: blueprint.ExecEdge},
			{From: "E2", To: "N2", Kind: blueprint.ExecEdge},
		},
	}
	out := blueprint.GenerateMermaid(g)
	if !strings.Contains(out, "subgraph sg_0 [\"Event1\"]") {
		t.Errorf("expected first subgraph with label Event1, got:\n%s", out)
	}
	if !strings.Contains(out, "subgraph sg_1 [\"Event2\"]") {
		t.Errorf("expected second subgraph with label Event2, got:\n%s", out)
	}
	if !strings.Contains(out, "end") {
		t.Error("expected 'end' for subgraphs")
	}
}

func TestGenerateMermaidNodeSubtitle(t *testing.T) {
	g := blueprint.Graph{
		Nodes: []blueprint.Node{
			{ID: "K2Node_CallFunction_7", Kind: blueprint.KindDefault, Label: "GetActorLocation"},
			{ID: "K2Node_IfThenElse_8", Kind: blueprint.KindBranch, Label: "Branch"},
			{ID: "K2Node_Event_0", Kind: blueprint.KindEvent, Label: "OnTick"},
			{ID: "K2Node_VariableGet_3", Kind: blueprint.KindVariable, Label: "myVar"},
		},
		Edges: []blueprint.Edge{
			{From: "K2Node_CallFunction_7", To: "K2Node_IfThenElse_8", Kind: blueprint.ExecEdge},
			{From: "K2Node_Event_0", To: "K2Node_VariableGet_3", Kind: blueprint.ExecEdge},
		},
	}
	out := blueprint.GenerateMermaid(g)

	cases := []string{
		`K2Node_CallFunction_7["GetActorLocation<br/><small>K2Node_CallFunction_7</small>"]`,
		`K2Node_IfThenElse_8{"Branch<br/><small>K2Node_IfThenElse_8</small>"}`,
		`K2Node_Event_0(["OnTick<br/><small>K2Node_Event_0</small>"])`,
		`K2Node_VariableGet_3(["myVar<br/><small>K2Node_VariableGet_3</small>"])`,
	}
	for _, want := range cases {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output\ngot:\n%s", want, out)
		}
	}
}

func TestGenerateMermaidSubgraphNameNoSubtitle(t *testing.T) {
	g := blueprint.Graph{
		Nodes: []blueprint.Node{
			{ID: "K2Node_CustomEvent_5", Kind: blueprint.KindEvent, Label: "OnFire"},
			{ID: "K2Node_CallFunction_2", Kind: blueprint.KindDefault, Label: "SpawnActor"},
		},
		Edges: []blueprint.Edge{
			{From: "K2Node_CustomEvent_5", To: "K2Node_CallFunction_2", Kind: blueprint.ExecEdge},
		},
	}
	out := blueprint.GenerateMermaid(g)

	if !strings.Contains(out, `subgraph sg_0 ["OnFire"]`) {
		t.Errorf("subgraph header should be display label only, got:\n%s", out)
	}
	if strings.Contains(out, `subgraph sg_0 ["OnFire<br/>`) {
		t.Errorf("subgraph header must not include graph name subtitle, got:\n%s", out)
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

	subtitleChecks := []string{
		`<br/><small>K2Node_FunctionEntry_0</small>`,
		`<br/><small>K2Node_CallFunction_7</small>`,
		`<br/><small>K2Node_IfThenElse_8</small>`,
	}
	for _, want := range subtitleChecks {
		if !strings.Contains(out, want) {
			t.Errorf("expected subtitle %q in output", want)
		}
	}
}

func TestGenerateMermaidUnconnectedNode(t *testing.T) {
	g := blueprint.Graph{
		Nodes: []blueprint.Node{
			{ID: "K2Node_CallFunction_1", Kind: blueprint.KindDefault, Label: "GetActorLocation"},
			{ID: "K2Node_CallFunction_2", Kind: blueprint.KindDefault, Label: "SetActorLocation"},
			{ID: "K2Node_VariableSet_0", Kind: blueprint.KindVariable, Label: "myVar"}, // no edges
		},
		Edges: []blueprint.Edge{
			{From: "K2Node_CallFunction_1", To: "K2Node_CallFunction_2", Kind: blueprint.ExecEdge},
		},
	}
	out := blueprint.GenerateMermaid(g)
	if !strings.Contains(out, `K2Node_VariableSet_0`) {
		t.Errorf("unconnected node should appear in output, got:\n%s", out)
	}
}

func TestParseCase4UnconnectedNode(t *testing.T) {
	data, err := os.ReadFile("fixtures/case4.txt")
	if err != nil {
		t.Fatalf("cannot read fixture: %v", err)
	}
	g, err := blueprint.ParseBlueprint(string(data))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	found := false
	for _, n := range g.Nodes {
		if n.ID == "K2Node_VariableSet_0" {
			found = true
			if n.Scope != "testFunctionB" {
				t.Errorf("expected Scope=%q, got %q", "testFunctionB", n.Scope)
			}
			break
		}
	}
	if !found {
		t.Fatal("expected K2Node_VariableSet_0 to be present in parsed graph")
	}

	out := blueprint.GenerateMermaid(g)

	// Only one subgraph should exist — no orphan "Graph N".
	if strings.Count(out, "subgraph") != 1 {
		t.Errorf("expected exactly 1 subgraph, got:\n%s", out)
	}
	if !strings.Contains(out, "K2Node_VariableSet_0") {
		t.Errorf("unconnected node must appear in output:\n%s", out)
	}
	if strings.Contains(out, `"Graph 2"`) {
		t.Errorf("disconnected node should be merged into its scope subgraph, not 'Graph 2':\n%s", out)
	}
}

func TestGenerateMermaidScopeMerge(t *testing.T) {
	g := blueprint.Graph{
		Nodes: []blueprint.Node{
			{ID: "K2Node_FunctionEntry_0", Kind: blueprint.KindEntry, Label: "myFunc"},
			{ID: "K2Node_CallFunction_1", Kind: blueprint.KindDefault, Label: "DoThing"},
			{ID: "K2Node_VariableSet_0", Kind: blueprint.KindVariable, Label: "Set myVar", Scope: "myFunc"},
		},
		Edges: []blueprint.Edge{
			{From: "K2Node_FunctionEntry_0", To: "K2Node_CallFunction_1", Kind: blueprint.ExecEdge},
		},
	}
	out := blueprint.GenerateMermaid(g)

	if strings.Count(out, "subgraph") != 1 {
		t.Errorf("expected 1 subgraph after scope merge, got:\n%s", out)
	}
	if !strings.Contains(out, "K2Node_VariableSet_0") {
		t.Errorf("scoped isolated node must appear in output:\n%s", out)
	}
}
