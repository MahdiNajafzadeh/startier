package startier

import "log"

func Run(configPath string) error {
	c, err := LoadConfig(configPath)
	if err != nil {
		return err
	}
	log.Printf("CONFIG : %+v", c)
	ch := make(chan error)
	go RunDatabase(ch)
	go RunTun(ch)
	go RunNetwork(ch)
	return <-ch
}