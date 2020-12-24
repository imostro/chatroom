package main

import (
	"net"
)

type Service interface {
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
	serverAddr := "localhost:" + string(port)
	listener, err := net.Listen("tcp", serverAddr)
	if err != nil {
		return nil, err
	}
	return &ServiceSocket{
		listener:   listener,
		serverAddr: serverAddr,
		stopCh:     make(chan error),
	}, nil
}

func (s *ServiceSocket) Start() {

}

func (s *ServiceSocket) Stop() {

}
