package database

import "startier/internal/node/database/models"


func (d *Database) RegisterNode(node *models.Node) error {
	return d.db.Create(node).Error
}

func (d *Database) RegisterAddress(address *models.Address) error {
	// ...
}

func (d *Database) RegisterRoute(route *models.Route) error {
	// ...
}

func (d *Database) RegisterConnection(connection *models.Connection) error {
	// ...
}

func (d *Database) RegisterRemote(remote *models.Remote) error {
	// ...
}

// write query to get best route
func (d *Database) GetBestRoute(from, to string) ([]models.Node, error) {
	var nodes []models.Node
	err := d.db.Raw("SELECT * FROM nodes WHERE id IN (SELECT to_id FROM routes WHERE from_id IN (SELECT id FROM nodes WHERE hostname = ?) AND to_id IN (SELECT id FROM nodes WHERE hostname = ?))", from, to).Scan(&nodes).Error
	return nodes, err
}