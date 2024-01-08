package main

import (
	"fmt"
	"github.com/xxxxnnxxxx/gorawsock"
)

func main() {
	tcpserver := gorawsock.NewProtocolObject(gorawsock.SocketType_STREAM)
	err := tcpserver.InitAdapter(gorawsock.IFObtainType_DeviceLnkName, "\\Device\\NPF_{D50F087F-49E2-4423-B22F-DA7F46D42394}")
	if err != nil {
		fmt.Println(err)
		return
	}
	socket := gorawsock.CreateSocket(gorawsock.SocketType_STREAM, 8000)
	tcpserver.Bind(socket)
	err = tcpserver.Startup(0)
	if err != nil {
		fmt.Println(err)
		tcpserver.CloseAllofSockets()
		tcpserver.CloseDevice()
		return
	}

	client, ret := tcpserver.Accept()
	if ret == -1 {
		fmt.Println(err)
		tcpserver.CloseAllofSockets()
		tcpserver.CloseDevice()
		return
	}

	var result []byte
	recvLen := tcpserver.Recv(client, &result)
	if recvLen == -1 {
		fmt.Println("连接已经断开")
	}
	fmt.Println(string(result))
	ret = tcpserver.Send(client, []byte("i am server:"+socket.LocalIP.String()+"! hello "+client.RemoteIP.String()))
	if ret == -1 {
		fmt.Println(client.GetLastError())
		return
	}

	fmt.Println("发送成功")

}
