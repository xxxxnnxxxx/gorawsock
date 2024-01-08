package main

import (
	"fmt"
	"github.com/xxxxnnxxxx/gorawsock"
	"time"
)

func main() {
	udpserver := gorawsock.NewProtocolObject(gorawsock.SocketType_DGRAM)
	err := udpserver.InitAdapter(gorawsock.IFObtainType_DeviceLnkName, "\\Device\\NPF_{D50F087F-49E2-4423-B22F-DA7F46D42394}")
	if err != nil {
		fmt.Println(err)
		return
	}
	socket := gorawsock.CreateSocket(gorawsock.SocketType_DGRAM, 12000)
	udpserver.Bind(socket)
	err = udpserver.Startup()
	if err != nil {
		fmt.Println(err)
		udpserver.CloseAllofSockets()
		udpserver.CloseDevice()
		return
	}

	var result []byte
	recvLen := udpserver.Recv(socket, &result)
	if recvLen == -1 {
		fmt.Println("获取数据错误")
	}
	fmt.Println(string(result))

	udpserver.Send(socket, []byte("hello, client!!!"))

	time.Sleep(20 * time.Second)
}
