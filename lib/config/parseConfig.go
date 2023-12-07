package config

func PrepareConfigs() error {
	err := prepareCMCicadaConfig()
	if err != nil {
		return err
	}

	return nil
}
