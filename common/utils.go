package common

import (
	"fmt"
	"net"
	"strings"
)
const CENSOR_FMT_IPV6 = "%02x%02x:%02x%02x:****:****:****:****:****:****"
const CENSOR_FMT_IPV4 = "%s.%s.***.***"
const BUF_SIZE = 0x2000

func CensorIP(ip net.IP) string {

	if ip == nil {
		panic("CensorIP: ip is nil!")
	}
	// only show part of ip
	is_ipv4 := ip.To4() != nil

	if !is_ipv4 {
		// IPV6
		return fmt.Sprintf(CENSOR_FMT_IPV6, ip[0],
			ip[1],
			ip[2],
			ip[3],
		)
	} else {
		ip_str := ip.To4().String()
		ip_split := strings.Split(ip_str, ".")
		return fmt.Sprintf(CENSOR_FMT_IPV4, ip_split[0],
			ip_split[1],
		)
	}
}

func FlushConn(conn net.Conn) (out []byte, err error) {
	temp := make([]byte, BUF_SIZE)

	n := BUF_SIZE

	for {
		n, err = conn.Read(temp)

		temp = temp[:n]
		if err != nil {
			return
		}
		out = append(out, temp...)

		if n != BUF_SIZE {
			return
		}
	}
}