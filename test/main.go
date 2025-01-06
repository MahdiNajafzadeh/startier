package main

import (
	"fmt"

	"github.com/MahdiNajafzadeh/easynode/internal/easynode"
)

func main() {
	graph := easynode.NewGraph[int]()

	// Add nodes and edges
	graph.AddEdge(1, 2)
	graph.AddEdge(2, 3)
	graph.AddEdge(3, 4)
	graph.AddEdge(4, 5)
	graph.AddEdge(1, 6)

	tests := []struct {
		name     string
		from     int
		to       int
		expected []int
	}{
		{
			name:     "Shortest path exists",
			from:     1,
			to:       5,
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "Direct edge path",
			from:     1,
			to:       6,
			expected: []int{1, 6},
		},
		{
			name:     "No path exists",
			from:     1,
			to:       7,
			expected: nil,
		},
		{
			name:     "Path to self",
			from:     3,
			to:       3,
			expected: []int{3},
		},
		{
			name:     "Non-existent nodes",
			from:     8,
			to:       9,
			expected: nil,
		},
	}
	for _, test := range tests {
		result := graph.ShortestPath(test.from, test.to)
		fmt.Printf("result: %v\n", result)
	}
}
