package database

import (
	"errors"
	"startier/internal/node/database/models"
)

func (d *Database) RegisterNode(node *models.Node) error {
	return d.db.Create(node).Error
}

func (d *Database) RegisterAddress(node *models.Node, address *models.Address) error {
	if node.ID == 0 {
		return errors.New("invalid node id")
	}
	address.NodeID = node.ID
	return d.db.Create(address).Error
}

func (d *Database) RegisterRemote(node *models.Node, remote *models.Remote) error {
	if node.ID == 0 {
		return errors.New("invalid node id")
	}
	remote.NodeID = node.ID
	return d.db.Create(remote).Error
}
