package goserver

var (
	_ baseContext        = (*baseContextImpl)(nil)
	_ ConnectionContext  = (*connectionContextImpl)(nil)
	_ MessageContext     = (*messageContextImpl)(nil)
	_ ReadContext        = (*readContextImpl)(nil)
	_ CloseContext       = (*closeContextImpl)(nil)
	_ PanicContext       = (*panicContextImpl)(nil)
	_ RecvTimeOutContext = (*recvTimeOutContextImpl)(nil)
)

type (
	//baseContext .
	baseContext interface {
		Next(data ...interface{})
		Data() interface{}
		Error() error
	}

	//ConnectionContext .
	ConnectionContext interface {
		baseContext
		Conn() *Conn
	}

	//MessageContext .
	MessageContext interface {
		baseContext
		Conn() *Conn
		Packet() Packet
	}

	//ReadContext .
	ReadContext interface {
		baseContext
		Conn() *Conn
	}

	//CloseContext .
	CloseContext interface {
		baseContext
		State() *ConnState
	}

	//PanicContext .
	PanicContext interface {
		baseContext
		Conn() *Conn
		State() *ConnState
	}

	//RecvTimeOutContext .
	RecvTimeOutContext interface {
		baseContext
		Conn() *Conn
		State() *ConnState
	}
)

type (
	//baseContextImpl .
	baseContextImpl struct {
		conn *Conn
		next func()
		data interface{}
		err  error
	}

	connectionContextImpl struct {
		baseContextImpl
	}

	//messageContextImpl .
	messageContextImpl struct {
		baseContextImpl
		packet Packet
	}

	//readContextImpl .
	readContextImpl struct {
		baseContextImpl
	}

	//closeContextImpl .
	closeContextImpl struct {
		baseContextImpl
	}

	//panicContextImpl .
	panicContextImpl struct {
		baseContextImpl
	}

	recvTimeOutContextImpl struct {
		baseContextImpl
	}
)

func newBaseContext(conn *Conn, err error) baseContextImpl {
	return baseContextImpl{
		conn: conn,
		err:  err,
	}
}

func newMessageContext(conn *Conn, p Packet) MessageContext {
	return &messageContextImpl{
		baseContextImpl: newBaseContext(conn, nil),
		packet:          p,
	}
}

func newReadContext(conn *Conn) ReadContext {
	return &readContextImpl{
		baseContextImpl: newBaseContext(conn, nil),
	}
}

func newCloseContext(conn *Conn) CloseContext {
	return &closeContextImpl{
		baseContextImpl: newBaseContext(conn, nil),
	}
}

func newPanicContext(conn *Conn, err error) PanicContext {
	return &panicContextImpl{
		baseContextImpl: newBaseContext(conn, err),
	}
}

func newConnectionContext(conn *Conn) ConnectionContext {
	return &connectionContextImpl{
		baseContextImpl: newBaseContext(conn, nil),
	}
}

func newRecvTimeOutContext(conn *Conn) RecvTimeOutContext {
	return &recvTimeOutContextImpl{
		baseContextImpl: newBaseContext(conn, nil),
	}
}

//setNext .
func (c *baseContextImpl) setNext(next func()) {
	c.next = next
}

//Conn .
func (c *baseContextImpl) Conn() *Conn {
	return c.conn
}

//Next .
func (c *baseContextImpl) Next(data ...interface{}) {
	if len(data) > 0 {
		c.data = data[0]
	}
	if c.next != nil {
		c.next()
	}
}

//Data .
func (c *baseContextImpl) Data() interface{} {
	return c.data
}

//State .
func (c *baseContextImpl) State() *ConnState {
	return &c.Conn().state
}

//State .
func (c *baseContextImpl) Error() error {
	return c.err
}

//Packet .
func (c *messageContextImpl) Packet() Packet {
	return c.packet
}
