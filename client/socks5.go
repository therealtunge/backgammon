package client

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"

	"github.com/therealtunge/backgammon/common/logging"
)

const SOCKS_VERSION = 5

type Socks5Request struct {
	Version     uint8
	Command     uint8
	AddressType uint8
	DstAddr     net.Addr
	DstPort     uint16
}

const (
	SOCKS_CONNECT = 1
	SOCKS_BIND    = 2
	SOCKS_UDP     = 3
)

const (
	FMT_IPV6 = "%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x"
	FMT_IPV4 = "%d.%d.%d.%d"
)

var conn2port map[uint16]net.Conn = make(map[uint16]net.Conn)

func parsePacket(conn *bufio.Reader, connPort uint16) (error, []byte, bool, *string) {
	version, err := conn.ReadByte()

	if err != nil {
		return err, nil, false, nil
	}

	if version != SOCKS_VERSION {
		err = fmt.Errorf("socks version mismatch: expected %v, got %v", SOCKS_VERSION, version)
		return err, nil, false, nil
	}

	request, err := conn.ReadByte()

	if err != nil {
		return err, nil, false, nil
	}

	_, err = conn.ReadByte()
	if err != nil {
		return err, nil, false, nil
	}

	atyp, err := conn.ReadByte()
	if err != nil {
		return err, nil, false, nil
	}

	var _addr []byte
	var ipaddr net.IP
	var addr string

	switch atyp {
	case 0x01:
		// IPV4 :D
		_addr = make([]byte, 4)
		conn.Read(_addr)
		ipaddr = net.IP(_addr)
		logging.Info("connecting to Address: ", ipaddr)
		addr = ipaddr.String()
	case 0x03:
		// domain
		domainSize, _ := conn.ReadByte()
		var domain []byte = make([]byte, domainSize)
		conn.Read(domain)
		addr = string(domain)
	case 0x04:
		// IPV6 D:
		_addr = make([]byte, 16)
		conn.Read(_addr)
		ipaddr = net.IP(_addr)

		logging.Info("connecting to Address: ", ipaddr)

		addr = ipaddr.String()
	default:
		return fmt.Errorf("unknown address type"), nil, false, nil
	}

	_port := make([]byte, 2)
	conn.Read(_port)
	port := binary.BigEndian.Uint16(_port)
	switch request {
	case SOCKS_CONNECT:
		{
			if atyp == 0x04 {
				addr = fmt.Sprintf("[%s]:%d", addr, port)
			} else {
				addr = fmt.Sprintf("%s:%d", addr, port)
			}

			return nil, []byte{0x05, 0x00, 0x00, 0x01, 127, 0x0, 0x0, 0x1, 0x27, 0x18}, true, &addr // this is cursed as shit
		}
	default:
		return fmt.Errorf("socks: got invalid request :(\n"), nil, false, nil
	}
}
