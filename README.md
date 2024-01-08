# gorawsock

这个库主要是通过gopacket库模拟了tcp/udp通讯，实现数据传输的功能，`examples` 目录下 `client.go` 和 `server.go` 就是两个测试程序
。测试这两个程序，是通过python socket模拟服务器和客户端测试的，两个程序都能正常使用。

测试的 `client.py` 代码如下：

```python
import socket
import time

client = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
client.connect(('192.168.1.3', 8000))

client.send(bytes('i am " + client.getsocketname()[0] + 'hello 192.168.1.3','utf-8'))
from_server = client.recv(4096)
print(str(from_server))
time.sleep(5)
client.close()
```
运行截图如下：

![image](https://github.com/xxxxnnxxxx/gorawsock/blob/main/images/example_server.png)

### 附加说明

在一些情况下，服务器端可能尝试端口复用，比如说，原有主机有某些程序运行在80端口，你可以在主机上运行此`rawsock`
构造的服务程序，也开启80端口，两个程序启动并不冲突，都可以监听80端口，但正常情况客户端不是 `rawsock` 实现的程序
尝试连接服务器的80端口，是不能正常与 `rawsock` 服务器通讯的，因为首先会连接到原有程序的80端口，那么这个时候，就可以
通过使用 `rawsock` 实现客户端，构造异形的 tcp连接系统，与服务器通讯，注意：服务器端也必须是和客户端同样的规则，比如说，
seq与ack的确认方式改变，这样客户端既不与原有程序连接，又能和 `rawsock` 构造的服务器保持稳定通讯。

以上只是思路，目前没有测试实现，后续有时间会上测试代码。

### 关于代码中的说明

在客户端连接服务器时有函数：

```go
Connect(targetIP net.IP, targetPort uint16, nexthopMAC net.HardwareAddr) (*Socket, error)
```
其中有参数 `nexthopMAC` 这个指定的是下一跳物理地址， 在tcp封包中， 网络接口层 `eth` 需要传递下一跳的物理地址，
在规则中，同一个局域网属于直连，那么下一跳地址就是目标机器的MAC地址，如果是不在同一个局域网，那么这个就是网关的物理地址。
注意： 我们这里排除掉回环的情况(127.0.0.1本地通讯)