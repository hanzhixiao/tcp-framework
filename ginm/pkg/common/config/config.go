package config

import (
	"bytes"
	"flag"
	"gopkg.in/yaml.v3"
	"time"
)

var cfg *config

func init() {

	configFile := flag.String("config path", "../ginm/config/config.yaml", "path of config.yaml")
	cfg = newConfig(*configFile)
}

const (
	ServerModeTcp       = "tcp"
	ServerModeWebsocket = "websocket"
	ServerModeKcp       = "kcp"
)

type config struct {
	Server       *server       `yaml:"server"`
	Worker       *worker       `yaml:"worker"`
	Logger       *logger       `yaml:"logger"`
	GlobalObject *globalObject `yaml:"globalObject"`
}

type worker struct {
	WorkerNum int `yaml:"workerNum"`
	ChanSize  int `yaml:"chanSize"`
}

type server struct {
	Name            string `yaml:"name"`
	Port            int    `yaml:"port"`
	IpVersion       string `yaml:"ipVersion"`
	Ip              string `yaml:"ip"`
	GlobalQueueSize int    `yaml:"globalQueueSize"`
	MaxConn         int    `yaml:"maxConn"`
	IOReadBuffSize  int    `yaml:"ioReadBuffSize"`
}

type globalObject struct {
	RouterSlicesMode bool   `yaml:"routerSlicesMode"`
	HeartbeatMax     int    `yaml:"heartbeatMax"`
	Mode             string `yaml:"mode"`
}

type logger struct {
	LogDir            string `json:"logDir"`
	LogFile           string `json:"logFile"`
	LogIsolationLevel int    `json:"logIsolationLevel"`
}

func newConfig(configFile string) *config {
	conf := &config{}
	data, err := readConfig(configFile)
	if err != nil {
		panic(err)
		return nil
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(conf); err != nil {
		panic("read config file error")
		return nil
	}
	return conf
}
func GetConfig() *config {
	return cfg
}

func (g *config) HeartbeatMaxDuration() time.Duration {
	return time.Duration(g.GlobalObject.HeartbeatMax) * time.Second
}
