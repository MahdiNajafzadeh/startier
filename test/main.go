package main

import (
	"github.com/arnauddri/algorithms/data-structures/graph"
)

func main() {
	g := graph.NewUndirected()
	g.AddEdge(1, 2, 1)
	g.AddEdge(1, 2, 1)
	g.AddEdge(1, 2, 1)
}
