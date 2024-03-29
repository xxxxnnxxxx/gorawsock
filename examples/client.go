package main

import (
	"fmt"
	"github.com/xxxxnnxxxx/gorawsock"
	"net"
)

func main() {
	client := gorawsock.NewProtocolObject(gorawsock.SocketType_STREAM)
	err := client.InitAdapter(gorawsock.IFObtainType_DeviceLnkName, "\\Device\\NPF_{D50F087F-49E2-4423-B22F-DA7F46D42394}")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = client.Startup(0)
	if err != nil {
		fmt.Println(err)
		client.CloseAllofSockets()
		client.CloseDevice()
		return
	}
	socket, err := client.Connect(net.ParseIP("192.168.1.4"), 8000, net.HardwareAddr{0xe4, 0x5f, 0x01, 0x87, 0x5b, 0x1D})
	var result []byte
	recvLen := client.Recv(socket, &result)
	if recvLen == -1 {
		fmt.Println("连接已经断开")
	}
	fmt.Println(string(result))
	ret := client.Send(socket, []byte("hello"))
	if ret > 0 {
		fmt.Println("数据发送成功")
	}
}
