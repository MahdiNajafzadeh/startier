package easynode

func Run(configPath string) error {
	err := LoadConfig(configPath)
	if err != nil {
		return err
	}
	
	return err
}
