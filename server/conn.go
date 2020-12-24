package main

import (
	"net"
	"time"
)

type ConnService interface {
	Send()
	Received()
	Close()
}

type Conn struct {
	name       string
	conn       net.Conn
	rcvCh      chan []byte
	sendCh     chan []byte
	done       chan error
	hbTimer    *time.Timer
	hbInterval time.Duration
	hbTimeout  time.Duration
}
