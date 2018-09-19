package main

import (
	"fmt"
	"github.com/pkg/errors"
	"math/big"
	"net"
	"strconv"
	"strings"
)

func inetToN(ip string) (uint64, error) {
	if net.ParseIP(ip).To4() == nil {
		return 0, errors.Errorf("%s not a valid ipv4 address", ip)
	}
	ret := big.NewInt(0)
	ret.SetBytes(net.ParseIP(ip).To4())

	return ret.Uint64(), nil
}

func inetNToA(ip uint64) string {
	return fmt.Sprintf("%d.%d.%d.%d", byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip))
}

func encodeAddr(addr string) (uint64, error) {
	splitAddr := strings.Split(addr, ":")
	ip := splitAddr[0]
	port, err := strconv.ParseUint(splitAddr[1], 10, 64)
	if err != nil {
		return 0, err
	}
	ipn, err := inetToN(ip)
	return ipn*100000 + port, nil
}

func decodeAddr(addr uint64) string {
	ip := addr / 100000
	port := addr - ip*10000
	return fmt.Sprintf("%s:%d", inetNToA(ip), port)
}
