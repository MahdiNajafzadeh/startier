package startier

func (db *DB) AddAddress(address Address) (Address, bool) {
	db.addresses_mu.Lock()
	defer db.addresses_mu.Unlock()
	_, ok := db.addresses[address.Address]
	if ok {
		return Address{}, false
	}
	db.addresses[address.Address] = address
	return address, true
}

func (db *DB) GetAddress(id string) (Address, bool) {
	db.addresses_mu.RLock()
	defer db.addresses_mu.RUnlock()
	address, ok := db.addresses[id]
	return address, ok
}

func (db *DB) GetAllAddresss() []Address {
	db.addresses_mu.RLock()
	defer db.addresses_mu.RUnlock()
	addresses := []Address{}
	for _, v := range db.addresses {
		addresses = append(addresses, v)
	}
	return addresses
}

func (db *DB) UpdateAddress(id string, address Address) bool {
	db.addresses_mu.Lock()
	defer db.addresses_mu.Unlock()
	_, ok := db.addresses[id]
	if ok {
		db.addresses[id] = address
		return ok
	}
	return false
}
