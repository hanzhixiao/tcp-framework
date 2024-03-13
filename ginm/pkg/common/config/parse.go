package config

import (
	"mmo/ginm/pkg/utils"
	"os"
)

func readConfig(configFile string) ([]byte, error) {
	b, err := os.ReadFile(configFile)
	if err != nil {
		return nil, utils.Wrap(err, configFile)
	}
	return b, nil
	// cfgName := os.Getenv("CONFIG_NAME")
	// if len(cfgName) != 0 {
	// 	data, err := os.ReadFile(filepath.Join(cfgName, "config", "config.yaml"))
	// 	if err != nil {
	// 		data, err = os.ReadFile(filepath.Join(Root, "config", "config.yaml"))
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 	} else {
	// 		Root = cfgName
	// 	}
	// 	return data, nil
	// } else {
	// 	return os.ReadFile(fmt.Sprintf("../config/%s", "config.yaml"))
	// }
}
