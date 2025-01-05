package easynode

import "gonum.org/v1/gonum/graph/simple"

var _graph *simple.UndirectedGraph

func init() {
	_graph = simple.NewUndirectedGraph()
}

// func GetPathsFrom2Node() {
// 	var edges []Edge
// 	_db.Find(&edges)
// 	_graph.NewNode().ID()
// }
