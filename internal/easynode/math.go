package easynode

import (
	"sort"
)

type Graph struct {
	adjList map[string][]string
}

func NewGraph() *Graph {
	return &Graph{
		adjList: make(map[string][]string),
	}
}

func (g *Graph) AddEdge(from, to string) {
	g.adjList[from] = append(g.adjList[from], to)
}

func (g *Graph) FindAllPaths(start, end string) [][]string {
	var paths [][]string
	var currentPath []string
	visited := make(map[string]bool)
	g.dfs(start, end, visited, currentPath, &paths)
	// Sort paths by length (ascending)
	sort.Slice(paths, func(i, j int) bool {
		return len(paths[i]) < len(paths[j])
	})
	return paths
}

func (g *Graph) dfs(current, end string, visited map[string]bool, currentPath []string, paths *[][]string) {
	visited[current] = true
	currentPath = append(currentPath, current)

	if current == end {
		pathCopy := make([]string, len(currentPath))
		copy(pathCopy, currentPath)
		*paths = append(*paths, pathCopy)
	} else {
		for _, neighbor := range g.adjList[current] {
			if !visited[neighbor] {
				g.dfs(neighbor, end, visited, currentPath, paths)
			}
		}
	}

	visited[current] = false
	currentPath = currentPath[:len(currentPath)-1]
}

// func main() {
// 	graph := NewGraph()
// 	nodes := []string{"A", "B", "C", "D", "E"}
// 	for _, n1 := range nodes {
// 		for _, n2 := range nodes {
// 			if n1 != n2 {
// 				graph.AddEdge(n1, n2)
// 			}
// 		}
// 	}

// 	start, end := "A", "D"
// 	paths := graph.FindAllPaths(start, end)
// 	fmt.Printf("All paths from %s to %s (sorted by length):\n", start, end)
// 	for _, path := range paths {
// 		fmt.Println(path)
// 	}
// }
