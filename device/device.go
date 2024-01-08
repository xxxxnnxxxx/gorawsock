package device

import (
	"github.com/google/gopacket/pcap"
	"net"
	"net/netip"
)

type IPAddress struct {
	IP          netip.Addr
	NetmaskBits int // 子网掩码位数
}

func (p *IPAddress) ToIP() net.IP {
	return net.ParseIP(p.IP.String())
}

func OpenPcapDevice(deviceLnkName string) (*pcap.Handle, error) {
	handle, err := pcap.OpenLive(deviceLnkName, 65536, true, pcap.BlockForever)
	if err != nil {
		return nil, err
	}

	return handle, nil
}

// 设备句柄管理
type DeviceHandle struct {
	Handle     *pcap.Handle
	IsLoopback bool
}

func NewDeviceHandle() *DeviceHandle {
	return &DeviceHandle{
		Handle:     nil,
		IsLoopback: false,
	}
}

func (p *DeviceHandle) Open(devLnkName string) error {
	var err error
	if p.Handle == nil {
		p.Handle, err = OpenPcapDevice(devLnkName)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *DeviceHandle) Close() {
	if p.Handle != nil {
		p.Handle.Close()
	}
}

func SendBuf(handle *pcap.Handle, buf []byte) error {
	err := handle.WritePacketData(buf)
	return err
}
