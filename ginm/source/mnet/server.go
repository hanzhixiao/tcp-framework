package mnet

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/xtaci/kcp-go"
	"gopkg.in/yaml.v3"
	"mmo/ginm/pkg/common/config"
	"mmo/ginm/source/decoder"
	"mmo/ginm/source/inter"
	"mmo/ginm/zlog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

type server struct {
	port          int
	name          string
	ipVersion     string
	ip            string
	msgHandler    inter.MsgHandler
	connManager   inter.ConnManager
	tcpListener   *net.TCPListener
	onStartConn   inter.Hook
	onStopConn    inter.Hook
	hc            inter.HeartbeatChecker
	upgrader      *websocket.Upgrader
	cid           uint32
	websocketAuth func(r *http.Request) error
	decoder       inter.Decoder
}

func (s *server) GetDecoder() inter.Decoder {
	return s.decoder
}

func (s *server) GetFieldLength() *inter.LengthField {
	if s.decoder == nil {
		return nil
	}
	return s.decoder.GetLengthField()
}

func (s *server) SetDecoder(decoder inter.Decoder) {
	s.decoder = decoder
}

func (s *server) AddInterceptor(interceptor inter.Interceptor) {
	s.msgHandler.AddInterceptor(interceptor)
}

func (s *server) AddRouterSlices(msgID uint32, router ...inter.RouterHandler) inter.RouterSlices {
	return s.msgHandler.AddRouterSlices(msgID, router...)
}

func (s *server) Use(handlers ...inter.RouterHandler) inter.RouterSlices {
	return s.msgHandler.Use(handlers...)
}

func (s *server) Group(start, end uint32, Handler ...inter.RouterHandler) inter.GroupRouterSlices {
	return s.msgHandler.Group(start, end, Handler...)
}

func (s *server) StartHeartBeat(interval time.Duration) {
	checker := NewHeartbeatChecker(interval)
	s.AddRouter(checker.MsgID(), checker.Router())
	s.hc = checker
}

func (s *server) GetOnStartConn() inter.Hook {
	return s.onStartConn
}

func (s *server) GetOnStopConn() inter.Hook {
	return s.onStopConn
}

func (s *server) SetOnStartConn(h inter.Hook) {
	s.onStartConn = h
}

func (s *server) SetOnStopConn(h inter.Hook) {
	s.onStopConn = h
}

func (s *server) AddRouter(msgType uint32, router inter.Router) {
	s.msgHandler.AddRouter(msgType, router)
}

func (s *server) Start(port ...int) {
	if s.decoder != nil {
		s.msgHandler.AddInterceptor(s.decoder)
	}
	cfg := config.GetConfig()
	s.msgHandler.StartAllWorker(cfg.Worker.ChanSize)
	if port != nil {
		s.port = port[0]
	}

	switch cfg.GlobalObject.Mode {
	case config.ServerModeTcp:
		go s.listenTcpConn()
	case config.ServerModeWebsocket:
		go s.listenWebsocketConn()
	case config.ServerModeKcp:
		go s.listenKcpConn()
	default:
		go s.listenTcpConn()
		go s.listenWebsocketConn()
	}
}

func (s *server) startConn(conn inter.Conn) {
	if s.hc != nil {
		heartBeatChecker := s.hc.Clone()
		heartBeatChecker.BindConn(conn)
	}
	conn.Start()
}

func (s *server) Serve(port ...int) {
	s.Start(port...)
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	sig := <-c
	zlog.Infof("[SERVE] Zinx server , name %s, Serve Interrupt, signal = %v", s.name, sig)
}

func (s *server) Stop() {
	s.connManager.ClearConn()
	fmt.Println("All conn closed")
	s.msgHandler.StopAllWorkers()
	fmt.Println("All worker stopped")
	s.tcpListener.Close()
	fmt.Println("Tcp Listener closed")
}

func (s *server) listenTcpConn() {
	fmt.Println("Starting Tcp server...")
	cfg := config.GetConfig()
	tcpAddr, err := net.ResolveTCPAddr(s.ipVersion, fmt.Sprintf("%s:%d", s.ip, s.port))
	if err != nil {
		fmt.Println("Resolve Tcp addr err:", err.Error())
		return
	}
	s.tcpListener, err = net.ListenTCP(s.ipVersion, tcpAddr)
	if err != nil {
		fmt.Println("Listen Tcp err:", err.Error())
		return
	}
	fmt.Println("Start Tcp Listener successfully...")
	for {
		conn, err := s.tcpListener.AcceptTCP()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				zlog.Error("Listener closed")
				return
			}
			fmt.Println("Accept client conn err: ", err.Error())
			continue
		}
		if s.connManager.GetConnNum() >= cfg.Server.MaxConn {
			fmt.Println("Conn number out of maxConnNumber")
			conn.Close()
			continue
		}
		fmt.Printf("client %s connected\n", conn.RemoteAddr().String())
		tcpConn := NewConn(s, conn, s.cid, s.msgHandler)
		atomic.AddUint32(&s.cid, 1)
		s.connManager.AddConn(tcpConn)
		s.startConn(tcpConn)
	}
}

func (s *server) listenWebsocketConn() {
	cfg := config.GetConfig()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if s.connManager.GetConnNum() >= cfg.Server.MaxConn {
			zlog.Infof("Exceeded the maxConnNum:%d", cfg.Server.MaxConn)
			return
		}
		// 2. If websocket authentication is required, set the authentication information
		// (如果需要 websocket 认证请设置认证信息)
		if s.websocketAuth != nil {
			err := s.websocketAuth(r)
			if err != nil {
				zlog.Errorf(" websocket auth err:%v", err)
				w.WriteHeader(401)
				return
			}
		}
		// 3. Check if there is a subprotocol specified in the header
		// (判断 header 里面是有子协议)
		if len(r.Header.Get("Sec-Websocket-Protocol")) > 0 {
			s.upgrader.Subprotocols = websocket.Subprotocols(r)
		}
		// 4. Upgrade the connection to a websocket connection
		// (升级成 websocket 连接)
		conn, err := s.upgrader.Upgrade(w, r, nil)
		if err != nil {
			zlog.Ins().ErrorF("new websocket err:%v", err)
			w.WriteHeader(500)
			return
		}
		newCid := atomic.AddUint32(&s.cid, 1)
		wsConn := NewWsConnection(s, conn, newCid, s.msgHandler, s.onStartConn, s.onStopConn)
		go s.startConn(wsConn)
	})

	err := http.ListenAndServe(fmt.Sprintf("%s:%d", s.ip, s.port), nil)
	if err != nil {
		panic(err)
	}
}

func (s *server) SetWebsocketAuth(f func(r *http.Request) error) {
	s.websocketAuth = f
}

func (s *server) listenKcpConn() {
	conf := config.GetConfig()
	listener, err := kcp.Listen(fmt.Sprintf("%s:%d", s.ip, s.port))
	if err != nil {
		zlog.Errorf("[START] resolve KCP addr err: %v\n", err)
		return
	}
	zlog.Infof("[START] KCP server listening at IP: %s, Port %d, Addr %s", s.ip, s.port, listener.Addr().String())
	go func() {
		for {
			if s.connManager.GetConnNum() >= conf.Server.MaxConn {
				zlog.Infof("Exceeded the maxConnNum:%d", conf.Server.MaxConn)
				continue
			}
			conn, err := listener.Accept()
			if err != nil {
				zlog.Errorf("Accept KCP err: %v", err)
				continue
			}
			newCid := atomic.AddUint32(&s.cid, 1)
			dealConn := NewKcpServerConn(s, conn.(*kcp.UDPSession), newCid, s.onStartConn, s.onStopConn, s.msgHandler)

			go s.startConn(dealConn)
		}
	}()
}

func NewServer() inter.Server {
	cfg := config.GetConfig()
	configData, _ := yaml.Marshal(cfg)
	fmt.Println("Server config\n", string(configData))
	return &server{
		port:        cfg.Server.Port,
		name:        cfg.Server.Name,
		ipVersion:   cfg.Server.IpVersion,
		ip:          cfg.Server.Ip,
		decoder:     decoder.NewTLVDecoder(),
		msgHandler:  NewMessageHandler(cfg.Worker.WorkerNum),
		connManager: NewConnManager(),
		upgrader: &websocket.Upgrader{
			ReadBufferSize: cfg.Server.IOReadBuffSize,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}
