package arping

import (
	"errors"
	"fmt"
	"net"
)

// findIPInNetworkFromIface find an ip from iface as src
func findIPInNetworkFromIface(dstIP net.IP, iface net.Interface, ignoreNet bool) (net.IP, error) {
	addrs, err := iface.Addrs()

	if err != nil {
		return nil, err
	}

	if len(addrs) == 0 {
		return nil, fmt.Errorf("iface: '%s' do not contains any ip", iface.Name)
	}

	var firstIP net.IP
	for i, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok {
			if ipnet.Contains(dstIP) {
				return ipnet.IP, nil
			}
			if i == 0 {
				firstIP = ipnet.IP
			}
		}
	}

	if ignoreNet {
		return firstIP, nil
	}
	return nil, fmt.Errorf("iface: '%s' can't reach ip: '%s'", iface.Name, dstIP)
}

func findUsableInterfaceForNetwork(dstIP net.IP) (*net.Interface, error) {
	ifaces, err := net.Interfaces()

	if err != nil {
		return nil, err
	}

	isDown := func(iface net.Interface) bool {
		return iface.Flags&1 == 0
	}

	hasAddressInNetwork := func(iface net.Interface) bool {
		if _, err := findIPInNetworkFromIface(dstIP, iface, false); err != nil {
			return false
		}
		return true
	}

	verboseLog.Println("search usable interface")
	logIfaceResult := func(msg string, iface net.Interface) {
		verboseLog.Printf("%10s: %6s %18s  %s", msg, iface.Name, iface.HardwareAddr, iface.Flags)
	}

	for _, iface := range ifaces {
		if isDown(iface) {
			logIfaceResult("DOWN", iface)
			continue
		}

		if !hasAddressInNetwork(iface) {
			logIfaceResult("OTHER NET", iface)
			continue
		}

		logIfaceResult("USABLE", iface)
		return &iface, nil
	}
	return nil, errors.New("no usable interface found")
}
