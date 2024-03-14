package grpcclientpool

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type CPool struct {
	mu         sync.Mutex
	queue      chan *qChan
	address    string
	maxOpen    int
	maxIdle    int
	credential credentials.TransportCredentials
	idleC      map[string]*Connection
	openCCount int
}

type Connection struct {
	id   string
	p    *CPool
	Conn *grpc.ClientConn
}

type qChan struct {
	c chan *Connection
	e chan error
}

type CPoolInterface interface {
	Release()
	Get() (*Connection, error)
	newConnection() (*Connection, error)
	connectionQueueProccess()
	update(conn *Connection)
}

func (cp *CPool) Release() {
	if len(cp.idleC) > 0 {
		for _, val := range cp.idleC {
			cp.mu.Lock()
			delete(cp.idleC, val.id)
			val.Conn.Close()
			cp.mu.Unlock()
		}
	}
}

func (cp *CPool) Get() (*Connection, error) {
	cp.mu.Lock()
	if len(cp.idleC) > 0 {
		for _, val := range cp.idleC {
			delete(cp.idleC, val.id)
			cp.openCCount++
			cp.mu.Unlock()
			return val, nil
		}
	}
	if cp.openCCount >= cp.maxOpen {
		qr := &qChan{
			c: make(chan *Connection),
			e: make(chan error),
		}
		cp.queue <- qr
		select {
		case conn := <-qr.c:
			cp.openCCount++
			cp.mu.Unlock()
			return conn, nil
		case err := <-qr.e:
			cp.mu.Unlock()
			return nil, err
		}
	}
	conn, err := cp.newConnection()
	if err != nil {
		return nil, err
	}

	cp.openCCount++
	cp.mu.Unlock()
	return conn, nil
}

func (cp *CPool) newConnection() (*Connection, error) {
	conn, err := grpc.Dial(cp.address, grpc.WithTransportCredentials(cp.credential))
	if err != nil {
		return nil, err
	}
	return &Connection{
		id:   fmt.Sprintf("%v", time.Now().Unix()),
		p:    cp,
		Conn: conn,
	}, nil
}

func (cp *CPool) connectionQueueProccess() {
	for rq := range cp.queue {
		var (
			isTimeout   = false
			isCompleted = false
			timeout     = time.After(time.Duration(3) * time.Second)
		)
		for {
			if isCompleted || isTimeout {
				break
			}
			select {
			case <-timeout:
				isTimeout = true
				rq.e <- errors.New("grpc: timed out when dialing")
			default:
				cp.mu.Lock()
				idles := len(cp.idleC)
				if idles > 0 {
					for _, val := range cp.idleC {
						delete(cp.idleC, val.id)
						cp.mu.Unlock()
						rq.c <- val
						isCompleted = true
						break
					}
				} else if cp.maxOpen > cp.openCCount {
					cp.openCCount++
					cp.mu.Unlock()

					conn, err := cp.newConnection()
					cp.mu.Lock()
					cp.openCCount--
					cp.mu.Unlock()
					if err == nil {
						rq.c <- conn
						isCompleted = true
					}

				} else {
					cp.mu.Unlock()
				}
			}
		}
	}
}

func (c *Connection) Release() {
	c.p.update(c)
}

func (cp *CPool) update(conn *Connection) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	if cp.maxIdle >= len(cp.idleC) {
		cp.idleC[conn.id] = conn
	} else {
		cp.openCCount--
		_ = conn.Conn.Close()
	}
}

type ClientPoolGRPC struct {
	MaxOpenConnection int    `mapstructure:"max_open_connection"`
	MaxIdleConnection int    `mapstructure:"max_idle_connection"`
	QueueTotal        int    `mapstructure:"queue_total"`
	Address           string `mapstructure:"address"`
	Credential        credentials.TransportCredentials
}

func New(c *ClientPoolGRPC) CPoolInterface {
	clientPool := &CPool{
		mu:         sync.Mutex{},
		address:    c.Address,
		maxOpen:    c.MaxOpenConnection,
		maxIdle:    c.MaxOpenConnection,
		credential: c.Credential,
		openCCount: 0,
		queue:      make(chan *qChan, c.QueueTotal),
		idleC:      make(map[string]*Connection, 0),
	}
	go clientPool.connectionQueueProccess()
	return clientPool
}
