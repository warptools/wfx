package wfx

import (
	"github.com/dominikbraun/graph"
)

func toposort(targets []*Target) ([]string, error) {
	g := graph.New(
		func(t *Target) string { return t.Name() },
		graph.Directed(),
		graph.PermitCycles(), // n.b., this does not do what the name suggests.
	)
	for _, t := range targets {
		g.AddVertex(t)
		for _, e := range t.dependsOn {
			g.AddEdge(t.Name(), e)
		}
	}
	order, err := graph.TopologicalSort(g)
	return order, err
}
