package sql

import (
	"fmt"
	"strings"
)

type DependecyNode struct {
	Name  string
	Edges []*DependecyNode
}

func (d *DependecyNode) String() string {
	depNames := []string{}
	for _, e := range d.Edges {
		depNames = append(depNames, e.Name)
	}
	return fmt.Sprintf("%v -> (%v)", d.Name, strings.Join(depNames, ","))
}

func (d *DependecyNode) AddEdge(node *DependecyNode) {
	d.Edges = append(d.Edges, node)
}

type DependecyGraph []DependecyNode

func (g *DependecyGraph) Add(node *DependecyNode) *DependecyGraph {
	*g = append(*g, *node)
	return g
}

func (g *DependecyGraph) Remove(node *DependecyNode) *DependecyGraph {
	graph := DependecyGraph{}
	for _, n := range *g {
		if n.String() == node.String() {
			continue
		}
		graph = append(graph, n)
	}
	return &graph
}

func (g *DependecyGraph) Has(node *DependecyNode) bool {
	for _, n := range *g {
		if n.String() == node.String() {
			return true
		}
	}
	return false
}

func (g *DependecyGraph) resolveNode(node *DependecyNode, resolved, unresolved *DependecyGraph) (bool, error) {
	unresolved = unresolved.Add(node)
	for _, edge := range node.Edges {
		if !resolved.Has(edge) {
			if unresolved.Has(edge) {
				return true, fmt.Errorf("circular reference detected: %v -> %s", node.Name, edge.Name)
			}

			cycle, err := g.resolveNode(edge, resolved, unresolved)
			if cycle {
				return true, err
			}
		}
	}
	// prevent from being added multiple times
	if !resolved.Has(node) {
		resolved = resolved.Add(node)
	}
	unresolved = unresolved.Remove(node)
	return false, nil
}

func (g *DependecyGraph) Items() []DependecyNode {
	return *g
}

func (g *DependecyGraph) TopSort(resolved *DependecyGraph) (bool, error) {
	unresolved := &DependecyGraph{}
	for i := range *g {
		cycle, err := g.resolveNode(&(*g)[i], resolved, unresolved)
		if cycle {
			return true, err
		}
	}
	return false, nil
}

func (g *DependecyGraph) CreateNode(name string) {
	for _, n := range *g {
		if n.Name == name {
			return
		}
	}
	*g = append(*g, DependecyNode{Name: name})
}

func (g DependecyGraph) Node(name string) *DependecyNode {
	for i, n := range g {
		if n.Name == name {
			return &g[i]
		}
	}
	return nil
}

func (g DependecyGraph) PrintNames() {
	for _, node := range g {
		fmt.Printf("%v\n", node.Name)
	}
}

func (g DependecyGraph) Print() {
	for _, node := range g {
		fmt.Printf("%v\n", node.String())
	}
}
