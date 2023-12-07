package config

import (
	"errors"
	"fmt"
	"github.com/cloud-barista/cm-cicada/common"
	"github.com/jollaman999/utils/fileutil"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type cmCicadaConfig struct {
	CMCicada struct {
		Listen struct {
			Port string `yaml:"port"`
		} `yaml:"listen"`
	} `yaml:"cm-cicada"`
}

var CMCicadaConfig cmCicadaConfig
var cmCicadaConfigFile = "cm-cicada.yaml"

func checkCMCicadaConfigFile() error {
	if CMCicadaConfig.CMCicada.Listen.Port == "" {
		return errors.New("config error: cm-cicada.listen.port is empty")
	}
	port, err := strconv.Atoi(CMCicadaConfig.CMCicada.Listen.Port)
	if err != nil || port < 1 || port > 65535 {
		return errors.New("config error: cm-cicada.listen.port has invalid value")
	}

	return nil
}

func readCMCicadaConfigFile() error {
	common.RootPath = os.Getenv(common.ModuleROOT)
	if len(common.RootPath) == 0 {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		common.RootPath = homeDir + "/." + strings.ToLower(common.ModuleName)
	}

	err := fileutil.CreateDirIfNotExist(common.RootPath)
	if err != nil {
		return err
	}

	ex, err := os.Executable()
	if err != nil {
		return err
	}

	exPath := filepath.Dir(ex)
	configDir := exPath + "/conf"
	if !fileutil.IsExist(configDir) {
		configDir = common.RootPath + "/conf"
	}

	data, err := os.ReadFile(configDir + "/" + cmCicadaConfigFile)
	if err != nil {
		return errors.New("can't find the config file (" + cmCicadaConfigFile + ")" + fmt.Sprintln() +
			"Must be placed in '." + strings.ToLower(common.ModuleName) + "/conf' directory " +
			"under user's home directory or 'conf' directory where running the binary " +
			"or 'conf' directory where placed in the path of '" + common.ModuleROOT + "' environment variable")
	}

	err = yaml.Unmarshal(data, &CMCicadaConfig)
	if err != nil {
		return err
	}

	err = checkCMCicadaConfigFile()
	if err != nil {
		return err
	}

	return nil
}

func prepareCMCicadaConfig() error {
	err := readCMCicadaConfigFile()
	if err != nil {
		return err
	}

	return nil
}
