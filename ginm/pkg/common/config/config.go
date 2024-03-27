package config

import (
	"os"
	"testing"
	"time"
)

var cfg *config

func init() {
	//wd, _ := os.Getwd()
	//configFile := flag.String("config path", wd+"../../../config/config.yaml", "path of config.yaml")
	//*configFile
	cfg = newConfig()
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
	WorkerNum int  `yaml:"workerNum"`
	ChanSize  int  `yaml:"chanSize"`
	RobMode   bool `yaml:"robMode"`
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

func newConfig() *config {
	pwd, err := os.Getwd()
	if err != nil {
		pwd = "."
	}

	// Note: Prevent errors like "flag provided but not defined: -test.paniconexit0" from occurring in go test.
	// (防止 go test 出现"flag provided but not defined: -test.paniconexit0"等错误)
	testing.Init()

	// Initialize the GlobalObject variable and set some default values.
	// (初始化GlobalObject变量，设置一些默认值)
	return &config{
		Server: &server{
			Name:            "Server",
			IpVersion:       "tcp4",
			Port:            8999,
			Ip:              "127.0.0.1",
			GlobalQueueSize: 1024,
			MaxConn:         12000,
			IOReadBuffSize:  1024,
		},
		GlobalObject: &globalObject{
			RouterSlicesMode: false,
			HeartbeatMax:     5,
			Mode:             ServerModeTcp,
		},
		Logger: &logger{
			LogDir:            pwd + "/logger",
			LogFile:           "runningLog.log",
			LogIsolationLevel: 0,
		},
		Worker: &worker{
			WorkerNum: 10,
			ChanSize:  1024,
			RobMode:   false,
		},
	}
}
func GetConfig() *config {
	return cfg
}

func (g *config) HeartbeatMaxDuration() time.Duration {
	return time.Duration(g.GlobalObject.HeartbeatMax) * time.Second
}
