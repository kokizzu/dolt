package env

import (
	"github.com/liquidata-inc/ld/dolt/go/libraries/doltcore/doltdb"
	"os/user"
	"path/filepath"
)

const (
	credsDir = "creds"

	configFile   = "config.json"
	globalConfig = "config_global.json"

	repoStateFile = "repo_state.json"
)

type HomeDirProvider func() (string, error)

func GetCurrentUserHomeDir() (string, error) {
	if usr, err := user.Current(); err != nil {
		return "", err
	} else {
		return usr.HomeDir, nil
	}
}

func getCredsDir(hdp HomeDirProvider) (string, error) {
	homeDir, err := hdp()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, doltdb.DoltDir, credsDir), nil
}

func getGlobalCfgPath(hdp HomeDirProvider) (string, error) {
	homeDir, err := hdp()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, doltdb.DoltDir, globalConfig), nil
}

func getLocalConfigPath() string {
	return filepath.Join(doltdb.DoltDir, configFile)
}

func getRepoStateFile() string {
	return filepath.Join(doltdb.DoltDir, repoStateFile)
}
