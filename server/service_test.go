package server

import (
	"fmt"
	"testing"
)

func TestServer(t *testing.T) {
	service, err := NewServiceSocket(10086)
	if err != nil {
		fmt.Println("start error!")
		return
	}
	service.Start()
}
