package ip

import (
	"net"
	"strconv"
	"strings"
)

// External get external ip.
func External() (res []string) {
	inters, err := net.Interfaces()
	if err != nil {
		return
	}
	for _, inter := range inters {
		if !strings.HasPrefix(inter.Name, "lo") {
			addrs, err := inter.Addrs()
			if err != nil {
				continue
			}
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok {
					if ipnet.IP.IsLoopback() || ipnet.IP.IsLinkLocalMulticast() || ipnet.IP.IsLinkLocalUnicast() {
						continue
					}
					if ip4 := ipnet.IP.To4(); ip4 != nil {
						switch true {
						case ip4[0] == 10:
							continue
						case ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31:
							continue
						case ip4[0] == 192 && ip4[1] == 168:
							continue
						default:
							res = append(res, ipnet.IP.String())
						}
					}
				}
			}
		}
	}
	return
}

// Internal get internal ip.
func Internal() string {
	inters, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, inter := range inters {
		if !isUp(inter.Flags) {
			continue
		}
		if !strings.HasPrefix(inter.Name, "lo") {
			addrs, err := inter.Addrs()
			if err != nil {
				continue
			}
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						return ipnet.IP.String()
					}
				}
			}
		}
	}
	return ""
}

// isUp Interface is up
func isUp(v net.Flags) bool {
	return v&net.FlagUp == net.FlagUp
}

// IPToUInt32 .
func IPToUInt32(ip net.IP) uint32 {
	bits := strings.Split(ip.String(), ".")
	if len(bits) < 4 {
		return 0
	}
	a, _ := strconv.Atoi(bits[0])
	b, _ := strconv.Atoi(bits[1])
	c, _ := strconv.Atoi(bits[2])
	d, _ := strconv.Atoi(bits[3])
	u := uint32(a) << 24
	u |= uint32(b) << 16
	u |= uint32(c) << 8
	u |= uint32(d)
	return u
}

// UInt32ToIP .
func UInt32ToIP(uip uint32) net.IP {
	var bytes [4]byte
	bytes[0] = byte(uip & 0xFF)
	bytes[1] = byte((uip >> 8) & 0xFF)
	bytes[2] = byte((uip >> 16) & 0xFF)
	bytes[3] = byte((uip >> 24) & 0xFF)
	return net.IPv4(bytes[3], bytes[2], bytes[1], bytes[0])
}
