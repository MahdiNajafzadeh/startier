package startier

func (db *DB) AddNode(node Node) (Node, bool) {
	db.nodes_mu.Lock()
	defer db.nodes_mu.Unlock()
	_, ok := db.nodes[node.ID]
	if ok {
		return Node{}, false
	}
	db.nodes[node.ID] = node
	return node, true
}

func (db *DB) GetNode(id string) (Node, bool) {
	db.nodes_mu.RLock()
	defer db.nodes_mu.RUnlock()
	node, ok := db.nodes[id]
	return node, ok
}

func (db *DB) GetAllNodes() []Node {
	db.nodes_mu.RLock()
	defer db.nodes_mu.RUnlock()
	nodes := []Node{}
	for _, v := range db.nodes {
		nodes = append(nodes, v)
	}
	return nodes
}

func (db *DB) UpdateNode(id string, node Node) bool {
	db.nodes_mu.Lock()
	defer db.nodes_mu.Unlock()
	_, ok := db.nodes[id]
	if ok {
		db.nodes[id] = node
		return ok
	}
	return false
}
