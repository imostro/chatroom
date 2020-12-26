package server

import (
	"context"
	"errors"
	"fmt"
	"learn/chatroom/entity"
	"log"
	"net"
	"strconv"
	"strings"
)

var (
	broadCastCh = make(chan *entity.Message, 128)
)

type ChatServer interface {
	Start()
	Stop()
}

type ServiceSocket struct {
	listener   net.Listener
	serverAddr string          // 服务器地址
	stopCh     chan error      // 通知服务停止通道
	clients    map[string]Conn // 已连接的用户，key: ip:port
}

func NewServiceSocket(port int) (s *ServiceSocket, err error) {
	serverAddr := "localhost:" + strconv.Itoa(port)
	listener, err := net.Listen("tcp", serverAddr)
	if err != nil {
		return nil, err
	}
	return &ServiceSocket{
		listener:   listener,
		serverAddr: serverAddr,
		stopCh:     make(chan error),
		clients:    make(map[string]Conn),
	}, nil
}

func (s *ServiceSocket) handlerAccept(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				s.stopCh <- err
			}
			go s.handlerConn(ctx, conn)
		}
	}
}

func (s *ServiceSocket) handlerConn(ctx context.Context, rowConn net.Conn) {
	conn := NewConn(rowConn)
	s.clients[conn.name] = conn

	connCtx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
		delete(s.clients, conn.name)
	}()

	log.Printf("-> client %s join on chat.", conn.name)
	go conn.Send(connCtx)
	go conn.Received(connCtx)
	go conn.HeatBeat(connCtx)

	err := <-conn.done
	fmt.Printf("client %s close connect, reason: %v", conn.name, err)
}

func (s *ServiceSocket) Start() {
	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		cancel()
		s.listener.Close()
	}()
	go s.handlerAccept(ctx)
	go s.Broadcast(ctx)
	<-s.stopCh
}

func (s *ServiceSocket) Broadcast(ctx context.Context) {
	for true {
		select {
		case msg := <-broadCastCh:
			for name, conn := range s.clients {
				if strings.EqualFold(msg.ClientName, name) {
					continue
				}
				conn.sendCh <- msg
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *ServiceSocket) Stop() {
	s.stopCh <- errors.New("stop service")
}
