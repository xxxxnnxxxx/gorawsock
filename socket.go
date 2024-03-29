package gorawsock

import (
	"encoding/binary"
	"github.com/google/gopacket/layers"
	"github.com/xxxxnnxxxx/gorawsock/device"
	"github.com/xxxxnnxxxx/gorawsock/utils"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// 协议类型
const (
	SocketType_STREAM int = iota + 1
	SocketType_DGRAM
)

// TCP首部标记
const (
	TCP_SIGNAL_FIN = 0x01
	TCP_SIGNAL_SYN = 0x02
	TCP_SIGNAL_RST = 0x04
	TCP_SIGNAL_PSH = 0x08
	TCP_SIGNAL_ACK = 0x10
	TCP_SIGNAL_URG = 0x20
)

// tcp状态
const (
	TS_UNKNOWN      int = iota
	TCP_ESTABLISHED     // 连接建立：数据传送在进行
	TCP_SYN_SENT        // 已发送SYN:等待ACK
	TCP_SYN_RECV        // 已发送SYN+ACK: 等待ACK
	TCP_FIN_WAIT1       // 第一个FIN 已发送：等待ACK
	TCP_FIN_WAIT2       // 对第一个FIN 的ACK已收到：等待第二个FIN
	TCP_TIME_WAIT       // 收到第二个FIN, 已发送ACK: 等待2MSL超时
	TCP_CLOSE           // 没有连接
	TCP_CLOSE_WAIT      // 收到第一个FIN , 已发送ACK:等待应用程序关闭
	TCP_LAST_ACK        // 收到第二个FIN, 已发送ACK: 等待2MSL超时
	TCP_LISTEN          // 收到了被动打开：等待 SYN
	TCP_CLOSING         /* Now a valid state */ // 双发都已经决定同时关闭

	TCP_MAX_STATES /* Leave at the end! */
)

// tcp 状态字符串
var TCPStatusInfoMap = map[int]string{
	TCP_ESTABLISHED: "Estableshed",
	TCP_SYN_SENT:    "SynSent",
	TCP_SYN_RECV:    "SynRecv",
	TCP_FIN_WAIT1:   "FinWait1",
	TCP_FIN_WAIT2:   "FinWait2",
	TCP_TIME_WAIT:   "TimeWait",
	TCP_CLOSE:       "Close",
	TCP_CLOSE_WAIT:  "CloseWait",
	TCP_LAST_ACK:    "LastACK",
	TCP_LISTEN:      "Listening",
	TCP_CLOSING:     "Closing",
}

// socket 消息
// 主要用于控制接受数据使用
const (
	SocketMsg_Unknow int = iota
	SocketMsg_RecvData
	SocketMsg_Closed
)

// 最大报文段寿命
var MSL int = 30

func TimerCall(interval int, count int, callback func()) {
	for i := 0; i < count; i++ {
		timer := time.NewTimer(time.Duration(interval) * time.Second)
		<-timer.C
		go callback()
	}
}

func getCurrentTimestampBigEndian() []byte {
	now := time.Now()
	ts := now.Unix()

	result := make([]byte, 4)
	binary.BigEndian.PutUint32(result, uint32(ts))

	return result
}

// 生成随机端口
// 来自nmap 的源码
/*
#define PRIME_32K 32261
//Change base_port to a new number in a safe port range that is unlikely to
//     conflict with nearby past or future invocations of ultra_scan.
static u16 increment_base_port() {
static u16 g_base_port = 33000 + get_random_uint() % PRIME_32K;
	g_base_port = 33000 + (g_base_port - 33000 + 256) % PRIME_32K;
	return g_base_port;
}
*/
const PRIME_32K int = 32261

func GeneratePort() int {
	base_port := 33000 + int(utils.GenerateRandomUint())%PRIME_32K
	base_port = 33000 + (base_port-33000+256)%PRIME_32K

	return base_port
}

func generateRandowSeq() uint32 {
	// 使用当前时间的纳秒级别时间戳作为种子
	rand.Seed(time.Now().UnixNano())

	// 生成一个随机的 uint32 整数
	randomUint32 := rand.Uint32()

	return randomUint32
}

type TCPSock struct {
	Status int // TCP状态
	// TCP 信息
	SeqNum                                     uint32 // 顺序号
	AckNum                                     uint32 // 确认序列号
	RecvedSeqNum                               uint32 // 接收顺序号
	RecvedAckNum                               uint32 // 确认序列号
	DataOffset                                 uint8
	FIN, SYN, RST, PSH, ACK, URG, ECE, CWR, NS bool
	PreRecvedSignal                            int // 当前接收的信号  FIN/SYN/RST/PSH/ACK/URG/ECE/CWR/NS 分析后得到的信号
	PreSentSignal                              int // FIN/SYN/RST/PSH/ACK/URG/ECE/CWR/NS 分析后得到的信号
	Options                                    []layers.TCPOption
	IsSupportTimestamp                         bool   // 是否支持时间戳
	TsEcho                                     uint32 // 时间戳相关
	MSS                                        uint16 // 最大报文长度
	WinSize                                    uint16 // 窗口大小
	RecvdWinSize                               uint16 // 接收窗口大小
	Step                                       uint32 // seq/ack 步进
}

type UDPSock struct {
	Length uint16
}

type Socket struct {
	Lock       sync.RWMutex
	Family     layers.ProtocolFamily
	SocketType int // 数据类型
	Handle     *device.DeviceHandle

	RemoteIP   net.IP // 连接来源的IP
	RemotePort uint16 // 远程端口号
	Nexthop    net.HardwareAddr

	LocalIP   net.IP
	LocalPort uint16
	LocalMAC  net.HardwareAddr

	TCPSock
	UDPSock             // 保留
	PreLenOfSent uint32 // 上一个发送的数据包长度

	LenOfRecved   uint32 // 接收的数据包长度
	RecvedPayload []byte

	// 通知回调，触发通知
	IsTriggerNotify atomic.Bool
	NotifyCallback  func()

	DataBuf        *utils.Buffer
	lock_lastError sync.Mutex
	lastError      error
}

func NewSocket() *Socket {
	result := &Socket{
		RecvedPayload:  make([]byte, 0),
		DataBuf:        utils.NewBuffer(),
		NotifyCallback: nil,
	}
	result.TCPSock.Step = 1 // 步进默认为1(符合常规tcp协议规则)
	result.TCPSock.Status = TS_UNKNOWN
	result.Options = make([]layers.TCPOption, 0)
	return result
}

func CreateSocket(socketType int,
	localPort uint16) *Socket {
	result := NewSocket()

	result.SocketType = socketType
	result.LocalPort = localPort

	return result
}

func (p *Socket) SetStep(step uint32) {
	p.Step = step
}

func (p *Socket) Clone() *Socket {
	pSocket := NewSocket()
	pSocket.Family = p.Family
	pSocket.SocketType = p.SocketType
	pSocket.Handle = p.Handle

	pSocket.RemoteIP = append(pSocket.RemoteIP, p.RemoteIP...)
	pSocket.RemotePort = p.RemotePort
	pSocket.Nexthop = append(pSocket.Nexthop, p.Nexthop...)

	pSocket.LocalIP = append(pSocket.LocalIP, p.LocalIP...)
	pSocket.LocalPort = p.LocalPort
	pSocket.LocalMAC = append(pSocket.LocalMAC, p.LocalMAC...)

	pSocket.TCPSock.Status = p.TCPSock.Status
	pSocket.TCPSock.SeqNum = p.TCPSock.SeqNum
	pSocket.TCPSock.AckNum = p.TCPSock.AckNum
	pSocket.TCPSock.RecvedAckNum = p.TCPSock.RecvedAckNum
	pSocket.TCPSock.DataOffset = p.TCPSock.DataOffset
	pSocket.TCPSock.FIN = p.TCPSock.FIN
	pSocket.TCPSock.SYN = p.TCPSock.SYN
	pSocket.TCPSock.RST = p.TCPSock.RST
	pSocket.TCPSock.PSH = p.TCPSock.PSH
	pSocket.TCPSock.ACK = p.TCPSock.ACK
	pSocket.TCPSock.URG = p.TCPSock.URG
	pSocket.TCPSock.ECE = p.TCPSock.ECE
	pSocket.TCPSock.CWR = p.TCPSock.CWR
	pSocket.TCPSock.NS = p.TCPSock.NS
	pSocket.TCPSock.PreRecvedSignal = p.TCPSock.PreRecvedSignal
	pSocket.TCPSock.PreSentSignal = p.TCPSock.PreSentSignal
	pSocket.TCPSock.Options = append(pSocket.TCPSock.Options, p.TCPSock.Options...)
	pSocket.TCPSock.IsSupportTimestamp = p.TCPSock.IsSupportTimestamp
	pSocket.TCPSock.TsEcho = p.TCPSock.TsEcho
	pSocket.TCPSock.MSS = p.TCPSock.MSS
	pSocket.TCPSock.WinSize = p.TCPSock.WinSize
	pSocket.TCPSock.RecvdWinSize = p.TCPSock.RecvdWinSize

	pSocket.UDPSock.Length = p.UDPSock.Length

	pSocket.PreLenOfSent = p.PreLenOfSent

	pSocket.LenOfRecved = p.LenOfRecved
	pSocket.RecvedPayload = append(pSocket.RecvedPayload, p.RecvedPayload...)

	return pSocket
}

// 获取下一个包的seq,表示发送下一个包，这个值就是对应的seq
func (p *Socket) GetNextSeq() uint32 {
	if p.SocketType == SocketType_STREAM {
		// ack 不消费顺序号
		if p.PreSentSignal != TCP_SIGNAL_ACK {
			if p.PreLenOfSent == 0 {
				return p.SeqNum + p.Step
			} else {
				return p.SeqNum + p.PreLenOfSent
			}
		}
	} else if p.SocketType == SocketType_DGRAM {
	}

	return 0
}

// 更新序列号
func (p *Socket) UpdateSeqNum() {
	if p.SocketType == SocketType_STREAM {
		if p.PreSentSignal != TCP_SIGNAL_ACK {
			if p.PreLenOfSent > 0 {
				p.SeqNum += p.PreLenOfSent
			} else {
				p.SeqNum += p.Step
			}
		}

	}
}

func (p *Socket) UpdateAckNum() {
	if p.SocketType == SocketType_STREAM {
		p.AckNum = p.RecvedSeqNum
		if p.PreRecvedSignal != TCP_SIGNAL_ACK {
			if p.LenOfRecved > 0 {
				p.AckNum += p.LenOfRecved
			} else {
				p.AckNum += p.Step
			}
		}
	}
}

func (p *Socket) GetTsEcho() uint32 {
	var result uint32
	for _, option := range p.Options {
		if option.OptionType == layers.TCPOptionKindTimestamps {
			result = binary.BigEndian.Uint32(option.OptionData[7:])
			break
		}
	}

	return result
}

// 设置最近的错误
func (p *Socket) SetLastError(err error) {
	p.lock_lastError.Lock()
	defer p.lock_lastError.Unlock()

	p.lastError = err
}

// 获取最近的错误
func (p *Socket) GetLastError() error {
	return p.lastError
}
