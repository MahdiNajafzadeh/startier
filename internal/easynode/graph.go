package easynode

var _graph Graph[string]

func init() {
	_graph = NewGraph[string]()
}

type Graph[T comparable] struct {
	nodes map[T]struct{}
	edges map[T]map[T]struct{}
}

func NewGraph[T comparable]() Graph[T] {
	return Graph[T]{
		nodes: make(map[T]struct{}),
		edges: make(map[T]map[T]struct{}),
	}
}

func (g *Graph[T]) AddNode(id T) {
	g.nodes[id] = struct{}{}
}

func (g *Graph[T]) AddEdge(from, to T) {
	g.AddNode(from)
	g.AddNode(to)
	if g.edges[from] == nil {
		g.edges[from] = make(map[T]struct{})
	}
	if g.edges[to] == nil {
		g.edges[to] = make(map[T]struct{})
	}
	g.edges[from][to] = struct{}{}
	g.edges[to][from] = struct{}{}
}

func (g *Graph[T]) ShortestPath(from, to T) []T {
	if _, ok := g.nodes[from]; !ok {
		return nil
	}
	if _, ok := g.nodes[to]; !ok {
		return nil
	}
	visited := make(map[T]bool)
	prev := make(map[T]T)
	queue := []T{from}
	visited[from] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if current == to {
			path := []T{}
			for at := to; at != from; at = prev[at] {
				path = append([]T{at}, path...)
			}
			return append([]T{from}, path...)
		}
		for neighbor := range g.edges[current] {
			if !visited[neighbor] {
				visited[neighbor] = true
				prev[neighbor] = current
				queue = append(queue, neighbor)
			}
		}
	}
	return nil
}
