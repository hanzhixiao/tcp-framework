package inter

type ConnManager interface {
	ClearConn()
	GetConn(connId uint32) (Conn, error)
	AddConn(conn Conn)
	RemoveConn(connId uint32) error
	GetConnNum() int
}
