package tests

import (
	"os"
	"testing"

	"UE_UML/internal/blueprint"
)

func TestParseEmpty(t *testing.T) {
	g, err := blueprint.ParseBlueprint("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(g.Nodes) != 0 {
		t.Errorf("expected 0 nodes, got %d", len(g.Nodes))
	}
	if len(g.Edges) != 0 {
		t.Errorf("expected 0 edges, got %d", len(g.Edges))
	}
}

func TestParseSingleNode(t *testing.T) {
	input := `Begin Object Class=/Script/BlueprintGraph.K2Node_CallFunction Name="K2Node_CallFunction_7"
   FunctionReference=(MemberName="myFunc")
End Object`
	g, err := blueprint.ParseBlueprint(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = g
}

func TestParseSingleNodeWithEdge(t *testing.T) {
	input := `Begin Object Class=/Script/BlueprintGraph.K2Node_FunctionEntry Name="K2Node_FunctionEntry_0"
   FunctionReference=(MemberName="myFunc")
   CustomProperties Pin (PinId=AAAA00000000000000000000000000AA,PinName="then",Direction="EGPD_Output",PinType.PinCategory="exec",LinkedTo=(K2Node_CallFunction_7 BBBB00000000000000000000000000BB,),)
End Object
Begin Object Class=/Script/BlueprintGraph.K2Node_CallFunction Name="K2Node_CallFunction_7"
   FunctionReference=(MemberName="doSomething")
   CustomProperties Pin (PinId=BBBB00000000000000000000000000BB,PinName="execute",PinType.PinCategory="exec",LinkedTo=(K2Node_FunctionEntry_0 AAAA00000000000000000000000000AA,),)
End Object`
	g, err := blueprint.ParseBlueprint(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(g.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(g.Nodes))
	}
	if len(g.Edges) != 1 {
		t.Errorf("expected 1 edge, got %d", len(g.Edges))
	}

	var entryNode *blueprint.Node
	for i := range g.Nodes {
		if g.Nodes[i].ID == "K2Node_FunctionEntry_0" {
			entryNode = &g.Nodes[i]
		}
	}
	if entryNode == nil {
		t.Fatal("K2Node_FunctionEntry_0 not found")
	}
	if entryNode.Kind != blueprint.KindEntry {
		t.Errorf("expected KindEntry, got %v", entryNode.Kind)
	}
	if entryNode.Label != "myFunc" {
		t.Errorf("expected label 'myFunc', got %q", entryNode.Label)
	}
}

func TestParseKindDetection(t *testing.T) {
	tests := []struct {
		name      string
		className string
		nodeID    string
		kind      blueprint.NodeKind
	}{
		{"FunctionEntry", "K2Node_FunctionEntry", "K2Node_FunctionEntry_0", blueprint.KindEntry},
		{"Event", "K2Node_Event", "K2Node_Event_0", blueprint.KindEvent},
		{"Branch", "K2Node_IfThenElse", "K2Node_IfThenElse_0", blueprint.KindBranch},
		{"Default", "K2Node_CallFunction", "K2Node_CallFunction_0", blueprint.KindDefault},
		{"VariableSet", "K2Node_VariableSet", "K2Node_VariableSet_0", blueprint.KindVariable},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			otherID := "K2Node_FunctionEntry_99"
			if tc.nodeID == otherID {
				otherID = "K2Node_CallFunction_99"
			}
			input := `Begin Object Class=/Script/BlueprintGraph.` + tc.className + ` Name="` + tc.nodeID + `"
   CustomProperties Pin (PinId=AAAA00000000000000000000000000AA,PinName="then",Direction="EGPD_Output",PinType.PinCategory="exec",LinkedTo=(` + otherID + ` BBBB00000000000000000000000000BB,),)
End Object
Begin Object Class=/Script/BlueprintGraph.K2Node_FunctionEntry Name="` + otherID + `"
   FunctionReference=(MemberName="x")
   CustomProperties Pin (PinId=BBBB00000000000000000000000000BB,PinName="execute",PinType.PinCategory="exec",LinkedTo=(` + tc.nodeID + ` AAAA00000000000000000000000000AA,),)
End Object`
			g, err := blueprint.ParseBlueprint(input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			var found *blueprint.Node
			for i := range g.Nodes {
				if g.Nodes[i].ID == tc.nodeID {
					found = &g.Nodes[i]
				}
			}
			if found == nil {
				t.Fatalf("node %s not found in graph", tc.nodeID)
			}
			if found.Kind != tc.kind {
				t.Errorf("expected kind %v, got %v", tc.kind, found.Kind)
			}
		})
	}
}

func TestParseFunctionLabelWithMemberParent(t *testing.T) {
	// MemberParent= comes before MemberName= — the old regex missed this case.
	input := `Begin Object Class=/Script/BlueprintGraph.K2Node_CallFunction Name="K2Node_CallFunction_11"
   FunctionReference=(MemberParent=Class'"/Script/Engine.KismetSystemLibrary"',MemberName="CapsuleTraceByChannel")
   CustomProperties Pin (PinId=AAAA00000000000000000000000000AA,PinName="then",Direction="EGPD_Output",PinType.PinCategory="exec",LinkedTo=(K2Node_FunctionEntry_0 BBBB00000000000000000000000000BB,),)
End Object
Begin Object Class=/Script/BlueprintGraph.K2Node_FunctionEntry Name="K2Node_FunctionEntry_0"
   FunctionReference=(MemberName="x")
   CustomProperties Pin (PinId=BBBB00000000000000000000000000BB,PinName="execute",PinType.PinCategory="exec",LinkedTo=(K2Node_CallFunction_11 AAAA00000000000000000000000000AA,),)
End Object`
	g, err := blueprint.ParseBlueprint(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var fn *blueprint.Node
	for i := range g.Nodes {
		if g.Nodes[i].ID == "K2Node_CallFunction_11" {
			fn = &g.Nodes[i]
		}
	}
	if fn == nil {
		t.Fatal("K2Node_CallFunction_11 not found")
	}
	if fn.Label != "CapsuleTraceByChannel" {
		t.Errorf("expected 'CapsuleTraceByChannel', got %q", fn.Label)
	}
}

func TestParseFunctionLabel(t *testing.T) {
	input := `Begin Object Class=/Script/BlueprintGraph.K2Node_CallFunction Name="K2Node_CallFunction_7"
   FunctionReference=(MemberName="CLIENT_getCameraPosByDist",MemberGuid=44882E1C458D2C810AB1BD95CB349DA9,bSelfContext=True)
   CustomProperties Pin (PinId=AAAA00000000000000000000000000AA,PinName="then",Direction="EGPD_Output",PinType.PinCategory="exec",LinkedTo=(K2Node_FunctionEntry_0 BBBB00000000000000000000000000BB,),)
End Object
Begin Object Class=/Script/BlueprintGraph.K2Node_FunctionEntry Name="K2Node_FunctionEntry_0"
   FunctionReference=(MemberName="myEntry")
   CustomProperties Pin (PinId=BBBB00000000000000000000000000BB,PinName="execute",PinType.PinCategory="exec",LinkedTo=(K2Node_CallFunction_7 AAAA00000000000000000000000000AA,),)
End Object`
	g, err := blueprint.ParseBlueprint(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var fn *blueprint.Node
	for i := range g.Nodes {
		if g.Nodes[i].ID == "K2Node_CallFunction_7" {
			fn = &g.Nodes[i]
		}
	}
	if fn == nil {
		t.Fatal("K2Node_CallFunction_7 not found")
	}
	if fn.Label != "CLIENT_getCameraPosByDist" {
		t.Errorf("expected label 'CLIENT_getCameraPosByDist', got %q", fn.Label)
	}
}

func TestParseVariableLabel(t *testing.T) {
	input := `Begin Object Class=/Script/BlueprintGraph.K2Node_VariableSet Name="K2Node_VariableSet_4"
   VariableReference=(MemberScope="traceTick",MemberName="hitActor",MemberGuid=1FD5CEB6440583A7C87C5FB9D505EFA0)
   CustomProperties Pin (PinId=AAAA00000000000000000000000000AA,PinName="then",Direction="EGPD_Output",PinType.PinCategory="exec",LinkedTo=(K2Node_FunctionEntry_0 BBBB00000000000000000000000000BB,),)
End Object
Begin Object Class=/Script/BlueprintGraph.K2Node_FunctionEntry Name="K2Node_FunctionEntry_0"
   FunctionReference=(MemberName="x")
   CustomProperties Pin (PinId=BBBB00000000000000000000000000BB,PinName="execute",PinType.PinCategory="exec",LinkedTo=(K2Node_VariableSet_4 AAAA00000000000000000000000000AA,),)
End Object`
	g, err := blueprint.ParseBlueprint(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var vs *blueprint.Node
	for i := range g.Nodes {
		if g.Nodes[i].ID == "K2Node_VariableSet_4" {
			vs = &g.Nodes[i]
		}
	}
	if vs == nil {
		t.Fatal("K2Node_VariableSet_4 not found")
	}
	if vs.Label != "Set hitActor" {
		t.Errorf("expected 'Set hitActor', got %q", vs.Label)
	}
}

func TestParseVariableGetLabelAndEdge(t *testing.T) {
	// VariableGet: label = just varName (no "Get " prefix), data edge label dropped.
	input := `Begin Object Class=/Script/BlueprintGraph.K2Node_VariableGet Name="K2Node_VariableGet_6"
   VariableReference=(MemberScope="foo",MemberName="hitActor",MemberGuid=AAAA00000000000000000000000000AA)
   CustomProperties Pin (PinId=BBBB00000000000000000000000000BB,PinName="hitActor",Direction="EGPD_Output",PinType.PinCategory="object",LinkedTo=(K2Node_CallFunction_7 CCCC00000000000000000000000000CC,),)
End Object
Begin Object Class=/Script/BlueprintGraph.K2Node_CallFunction Name="K2Node_CallFunction_7"
   FunctionReference=(MemberName="doThing")
   CustomProperties Pin (PinId=CCCC00000000000000000000000000CC,PinName="target",PinType.PinCategory="object",LinkedTo=(K2Node_VariableGet_6 BBBB00000000000000000000000000BB,),)
End Object`
	g, err := blueprint.ParseBlueprint(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var vg *blueprint.Node
	for i := range g.Nodes {
		if g.Nodes[i].ID == "K2Node_VariableGet_6" {
			vg = &g.Nodes[i]
		}
	}
	if vg == nil {
		t.Fatal("K2Node_VariableGet_6 not found")
	}
	if vg.Kind != blueprint.KindVariable {
		t.Errorf("expected KindVariable, got %v", vg.Kind)
	}
	if vg.Label != "hitActor" {
		t.Errorf("expected label 'hitActor' (no 'Get ' prefix), got %q", vg.Label)
	}

	// Data edge from VariableGet should have empty label (no noise)
	for _, e := range g.Edges {
		if e.From == "K2Node_VariableGet_6" && e.Kind == blueprint.DataEdge {
			if e.Label != "" {
				t.Errorf("expected empty edge label from VariableGet, got %q", e.Label)
			}
		}
	}
}

func TestParseEventLabel(t *testing.T) {
	input := `Begin Object Class=/Script/BlueprintGraph.K2Node_Event Name="K2Node_Event_0"
   EventReference=(MemberParent=Class'"/Script/sUtility.I_ActorComponentRedirector"',MemberName="OnPossessed")
   CustomProperties Pin (PinId=AAAA00000000000000000000000000AA,PinName="then",Direction="EGPD_Output",PinType.PinCategory="exec",LinkedTo=(K2Node_CallFunction_7 BBBB00000000000000000000000000BB,),)
End Object
Begin Object Class=/Script/BlueprintGraph.K2Node_CallFunction Name="K2Node_CallFunction_7"
   FunctionReference=(MemberName="doThing")
   CustomProperties Pin (PinId=BBBB00000000000000000000000000BB,PinName="execute",PinType.PinCategory="exec",LinkedTo=(K2Node_Event_0 AAAA00000000000000000000000000AA,),)
End Object`
	g, err := blueprint.ParseBlueprint(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var ev *blueprint.Node
	for i := range g.Nodes {
		if g.Nodes[i].ID == "K2Node_Event_0" {
			ev = &g.Nodes[i]
		}
	}
	if ev == nil {
		t.Fatal("K2Node_Event_0 not found")
	}
	if ev.Label != "OnPossessed" {
		t.Errorf("expected 'OnPossessed', got %q", ev.Label)
	}
}

func TestParseExecEdge(t *testing.T) {
	input := `Begin Object Class=/Script/BlueprintGraph.K2Node_FunctionEntry Name="K2Node_FunctionEntry_0"
   FunctionReference=(MemberName="myFunc")
   CustomProperties Pin (PinId=AAAA00000000000000000000000000AA,PinName="then",Direction="EGPD_Output",PinType.PinCategory="exec",LinkedTo=(K2Node_CallFunction_7 BBBB00000000000000000000000000BB,),)
End Object
Begin Object Class=/Script/BlueprintGraph.K2Node_CallFunction Name="K2Node_CallFunction_7"
   FunctionReference=(MemberName="doSomething")
   CustomProperties Pin (PinId=BBBB00000000000000000000000000BB,PinName="execute",PinType.PinCategory="exec",LinkedTo=(K2Node_FunctionEntry_0 AAAA00000000000000000000000000AA,),)
End Object`
	g, err := blueprint.ParseBlueprint(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(g.Edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(g.Edges))
	}
	e := g.Edges[0]
	if e.From != "K2Node_FunctionEntry_0" || e.To != "K2Node_CallFunction_7" {
		t.Errorf("unexpected edge: %+v", e)
	}
	if e.Label != "" {
		t.Errorf("expected empty label, got %q", e.Label)
	}
}

func TestParseBranchEdgeLabels(t *testing.T) {
	input := `Begin Object Class=/Script/BlueprintGraph.K2Node_IfThenElse Name="K2Node_IfThenElse_8"
   CustomProperties Pin (PinId=AAAA00000000000000000000000000AA,PinName="execute",PinType.PinCategory="exec",LinkedTo=(K2Node_FunctionEntry_0 CCCC00000000000000000000000000CC,),)
   CustomProperties Pin (PinId=BBBB00000000000000000000000000BB,PinName="then",PinFriendlyName=NSLOCTEXT("K2Node", "true", "true"),Direction="EGPD_Output",PinType.PinCategory="exec",LinkedTo=(K2Node_CallFunction_10 DDDD00000000000000000000000000DD,),)
   CustomProperties Pin (PinId=EEEE00000000000000000000000000EE,PinName="else",PinFriendlyName=NSLOCTEXT("K2Node", "false", "false"),Direction="EGPD_Output",PinType.PinCategory="exec",LinkedTo=(K2Node_CallFunction_20 FFFF00000000000000000000000000FF,),)
End Object
Begin Object Class=/Script/BlueprintGraph.K2Node_FunctionEntry Name="K2Node_FunctionEntry_0"
   FunctionReference=(MemberName="x")
   CustomProperties Pin (PinId=CCCC00000000000000000000000000CC,PinName="then",Direction="EGPD_Output",PinType.PinCategory="exec",LinkedTo=(K2Node_IfThenElse_8 AAAA00000000000000000000000000AA,),)
End Object
Begin Object Class=/Script/BlueprintGraph.K2Node_CallFunction Name="K2Node_CallFunction_10"
   FunctionReference=(MemberName="trueFunc")
   CustomProperties Pin (PinId=DDDD00000000000000000000000000DD,PinName="execute",PinType.PinCategory="exec",LinkedTo=(K2Node_IfThenElse_8 BBBB00000000000000000000000000BB,),)
End Object
Begin Object Class=/Script/BlueprintGraph.K2Node_CallFunction Name="K2Node_CallFunction_20"
   FunctionReference=(MemberName="falseFunc")
   CustomProperties Pin (PinId=FFFF00000000000000000000000000FF,PinName="execute",PinType.PinCategory="exec",LinkedTo=(K2Node_IfThenElse_8 EEEE00000000000000000000000000EE,),)
End Object`
	g, err := blueprint.ParseBlueprint(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	trueFound, falseFound := false, false
	for _, e := range g.Edges {
		if e.From == "K2Node_IfThenElse_8" && e.To == "K2Node_CallFunction_10" {
			if e.Label != "true" {
				t.Errorf("expected label 'true', got %q", e.Label)
			}
			trueFound = true
		}
		if e.From == "K2Node_IfThenElse_8" && e.To == "K2Node_CallFunction_20" {
			if e.Label != "false" {
				t.Errorf("expected label 'false', got %q", e.Label)
			}
			falseFound = true
		}
	}
	if !trueFound {
		t.Error("true branch edge not found")
	}
	if !falseFound {
		t.Error("false branch edge not found")
	}
}

func TestParseDataEdgesCreated(t *testing.T) {
	// Data output pins with LinkedTo should produce DataEdge entries.
	input := `Begin Object Class=/Script/BlueprintGraph.K2Node_FunctionEntry Name="K2Node_FunctionEntry_0"
   FunctionReference=(MemberName="myFunc")
   CustomProperties Pin (PinId=AAAA00000000000000000000000000AA,PinName="then",Direction="EGPD_Output",PinType.PinCategory="exec",LinkedTo=(K2Node_CallFunction_7 BBBB00000000000000000000000000BB,),)
End Object
Begin Object Class=/Script/BlueprintGraph.K2Node_CallFunction Name="K2Node_CallFunction_7"
   FunctionReference=(MemberName="doSomething")
   CustomProperties Pin (PinId=BBBB00000000000000000000000000BB,PinName="execute",PinType.PinCategory="exec",LinkedTo=(K2Node_FunctionEntry_0 AAAA00000000000000000000000000AA,),)
   CustomProperties Pin (PinId=CCCC00000000000000000000000000CC,PinName="dist",Direction="EGPD_Output",PinType.PinCategory="float",LinkedTo=(K2Node_CallFunction_99 DDDD00000000000000000000000000DD,),)
End Object
Begin Object Class=/Script/BlueprintGraph.K2Node_CallFunction Name="K2Node_CallFunction_99"
   FunctionReference=(MemberName="consumer")
   CustomProperties Pin (PinId=DDDD00000000000000000000000000DD,PinName="floatIn",PinType.PinCategory="float",LinkedTo=(K2Node_CallFunction_7 CCCC00000000000000000000000000CC,),)
End Object`
	g, err := blueprint.ParseBlueprint(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var dataEdge *blueprint.Edge
	for i := range g.Edges {
		e := &g.Edges[i]
		if e.From == "K2Node_CallFunction_7" && e.To == "K2Node_CallFunction_99" {
			dataEdge = e
		}
	}
	if dataEdge == nil {
		t.Fatal("expected data edge from K2Node_CallFunction_7 to K2Node_CallFunction_99")
	}
	if dataEdge.Kind != blueprint.DataEdge {
		t.Errorf("expected DataEdge kind, got %v", dataEdge.Kind)
	}
	if dataEdge.Label != "dist" {
		t.Errorf("expected label 'dist', got %q", dataEdge.Label)
	}
}

func TestParseDeduplication(t *testing.T) {
	input := `Begin Object Class=/Script/BlueprintGraph.K2Node_FunctionEntry Name="K2Node_FunctionEntry_0"
   FunctionReference=(MemberName="myFunc")
   CustomProperties Pin (PinId=AAAA00000000000000000000000000AA,PinName="then",Direction="EGPD_Output",PinType.PinCategory="exec",LinkedTo=(K2Node_CallFunction_7 BBBB00000000000000000000000000BB,),)
End Object
Begin Object Class=/Script/BlueprintGraph.K2Node_CallFunction Name="K2Node_CallFunction_7"
   FunctionReference=(MemberName="doSomething")
   CustomProperties Pin (PinId=BBBB00000000000000000000000000BB,PinName="execute",PinType.PinCategory="exec",LinkedTo=(K2Node_FunctionEntry_0 AAAA00000000000000000000000000AA,),)
End Object`
	g, err := blueprint.ParseBlueprint(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(g.Edges) != 1 {
		t.Errorf("expected 1 edge (deduped), got %d", len(g.Edges))
	}
}

func TestParseCase1Fixture(t *testing.T) {
	data, err := os.ReadFile("fixtures/case1.txt")
	if err != nil {
		t.Fatalf("cannot read fixture: %v", err)
	}
	g, err := blueprint.ParseBlueprint(string(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(g.Nodes) == 0 {
		t.Fatal("expected nodes from case1.txt")
	}

	var hasEntry bool
	for _, n := range g.Nodes {
		if n.Kind == blueprint.KindEntry {
			hasEntry = true
		}
	}
	if !hasEntry {
		t.Error("expected at least one KindEntry node")
	}

	var hasBranch bool
	for _, n := range g.Nodes {
		if n.Kind == blueprint.KindBranch {
			hasBranch = true
		}
	}
	if !hasBranch {
		t.Error("expected at least one KindBranch node")
	}

	var hasTrueEdge bool
	for _, e := range g.Edges {
		if e.Label == "true" {
			hasTrueEdge = true
		}
	}
	if !hasTrueEdge {
		t.Error("expected at least one edge with label 'true'")
	}
}

func TestParseCase3Fixture(t *testing.T) {
	data, err := os.ReadFile("fixtures/case3.txt")
	if err != nil {
		t.Fatalf("cannot read fixture: %v", err)
	}
	g, err := blueprint.ParseBlueprint(string(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 8 nodes: FunctionEntry, FunctionResult, CallFunction_1/2/4/5/6,
	// CommutativeAssociativeBinaryOperator_1
	if len(g.Nodes) != 8 {
		t.Errorf("expected 8 nodes, got %d", len(g.Nodes))
	}

	// FunctionEntry: KindEntry + correct label
	var entry *blueprint.Node
	for i := range g.Nodes {
		if g.Nodes[i].ID == "K2Node_FunctionEntry_0" {
			entry = &g.Nodes[i]
		}
	}
	if entry == nil {
		t.Fatal("K2Node_FunctionEntry_0 not found")
	}
	if entry.Kind != blueprint.KindEntry {
		t.Errorf("expected KindEntry, got %v", entry.Kind)
	}
	if entry.Label != "CLIENT_getCameraPosByDist" {
		t.Errorf("expected label 'CLIENT_getCameraPosByDist', got %q", entry.Label)
	}

	// FunctionResult: label must be "Return Node" (not the function name)
	var result *blueprint.Node
	for i := range g.Nodes {
		if g.Nodes[i].ID == "K2Node_FunctionResult_0" {
			result = &g.Nodes[i]
		}
	}
	if result == nil {
		t.Fatal("K2Node_FunctionResult_0 not found")
	}
	if result.Label != "Return Node" {
		t.Errorf("expected label 'Return Node', got %q", result.Label)
	}

	// Exec edge: FunctionEntry → FunctionResult (no label)
	var hasExecToResult bool
	for _, e := range g.Edges {
		if e.From == "K2Node_FunctionEntry_0" && e.To == "K2Node_FunctionResult_0" && e.Kind == blueprint.ExecEdge {
			hasExecToResult = true
		}
	}
	if !hasExecToResult {
		t.Error("expected exec edge FunctionEntry_0 → FunctionResult_0")
	}

	// Data edge: FunctionEntry → CallFunction_6 with label "dist"
	var distEdge *blueprint.Edge
	for i := range g.Edges {
		e := &g.Edges[i]
		if e.From == "K2Node_FunctionEntry_0" && e.To == "K2Node_CallFunction_6" {
			distEdge = e
		}
	}
	if distEdge == nil {
		t.Fatal("expected data edge FunctionEntry_0 → CallFunction_6")
	}
	if distEdge.Kind != blueprint.DataEdge {
		t.Errorf("expected DataEdge, got %v", distEdge.Kind)
	}
	if distEdge.Label != "dist" {
		t.Errorf("expected label 'dist', got %q", distEdge.Label)
	}

	// Data edge: CommutativeAssociativeBinaryOperator_1 → FunctionResult_0 with label "distance"
	var distanceEdge *blueprint.Edge
	for i := range g.Edges {
		e := &g.Edges[i]
		if e.From == "K2Node_CommutativeAssociativeBinaryOperator_1" && e.To == "K2Node_FunctionResult_0" {
			distanceEdge = e
		}
	}
	if distanceEdge == nil {
		t.Fatal("expected data edge CommutativeAssociativeBinaryOperator_1 → FunctionResult_0")
	}
	if distanceEdge.Kind != blueprint.DataEdge {
		t.Errorf("expected DataEdge, got %v", distanceEdge.Kind)
	}
	if distanceEdge.Label != "distance" {
		t.Errorf("expected label 'distance', got %q", distanceEdge.Label)
	}

}

func TestParseCase2Fixture(t *testing.T) {
	data, err := os.ReadFile("fixtures/case2.txt")
	if err != nil {
		t.Fatalf("cannot read fixture: %v", err)
	}
	g, err := blueprint.ParseBlueprint(string(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(g.Nodes) == 0 {
		t.Fatal("expected nodes from case2.txt")
	}

	var hasEvent bool
	for _, n := range g.Nodes {
		if n.Kind == blueprint.KindEvent {
			hasEvent = true
		}
	}
	if !hasEvent {
		t.Error("expected at least one KindEvent node in case2")
	}
}
