package startier

func (db *DB) AddRemote(remote Remote) (Remote, bool) {
	db.remotes_mu.Lock()
	defer db.remotes_mu.Unlock()
	_, ok := db.remotes[remote.NodeID]
	if ok {
		return Remote{}, false
	}
	db.remotes[remote.NodeID] = remote
	return remote, true
}

func (db *DB) GetRemote(id string) (Remote, bool) {
	db.remotes_mu.RLock()
	defer db.remotes_mu.RUnlock()
	remote, ok := db.remotes[id]
	return remote, ok
}

func (db *DB) GetAllRemotes() []Remote {
	db.remotes_mu.RLock()
	defer db.remotes_mu.RUnlock()
	remotes := []Remote{}
	for _, v := range db.remotes {
		remotes = append(remotes, v)
	}
	return remotes
}

func (db *DB) UpdateRemote(id string, remote Remote) bool {
	db.remotes_mu.Lock()
	defer db.remotes_mu.Unlock()
	_, ok := db.remotes[id]
	if ok {
		db.remotes[id] = remote
		return ok
	}
	return false
}
