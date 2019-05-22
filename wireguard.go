package main

import (
	"errors"
	"fmt"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"net"
	"strconv"
)

// refreshing the device is required before searching IPs
func dRefresh() {
	var err error
	d, err = c.Device(wgInterface)
	if err != nil {
		fmt.Println("could not get wireguard device from env var WIREGUARD_INTERFACE: ", wgInterface)
		fmt.Println("ERROR: ", err)
		panic(err)
	}
}

func dGetPeersJSON() string {
	r := "{"
	dRefresh()
	for i, p := range d.Peers {
		if i > 0 {
			r += ","
		}
		r += "\n\n    {"
		r += "\n    peerLoopIndex: " + strconv.Itoa(i)
		r += ",\n    publicKey: \"" + p.PublicKey.String()
		r += "\",\n    AllowedIPs: \""
		for _, ipn := range p.AllowedIPs {
			r += ipn.String() + " "
		}
		r += "\",\n    endpoint: \"" + p.Endpoint.String()
		r += "\",\n    lastHandshake: \"" + p.LastHandshakeTime.String()
		r += "\",\n    bytesReceived: " + strconv.FormatInt(p.ReceiveBytes, 10)
		r += ",\n    bytesSent: " + strconv.FormatInt(p.TransmitBytes, 10)
		r += "\n    }"
	}
	r = r + "\n}"
	return r
}

func dDeletePeer(ks string) error {
	k, err := wgtypes.NewKey([]byte(ks))
	if err != nil {
		return err
	}
	dRefresh()
	for _, p := range d.Peers {
		if p.PublicKey == k {
			peers := []wgtypes.PeerConfig{
				{
					PublicKey: k,
					Remove:    true,
				},
			}
			newConfig := wgtypes.Config{
				ReplacePeers: false,
				Peers:        peers,
			}
			// apply config to interface
			var err error
			err = c.ConfigureDevice(wgInterface, newConfig)
			if err != nil {
				panic(err)
			}
			return nil
		} else {
			err := errors.New("key not found")
			return err
		}
	}
	return errors.New("something horrible happened")
}

func dAddPeer(ks string, ips string) error {
	k, err := wgtypes.NewKey([]byte(ks))
	if err != nil {
		return err
	}
	_, ip, err := net.ParseCIDR(ips)
	if err != nil {
		return err
	}
	dRefresh()
	ipList := []net.IPNet{
		*ip,
	}
	for _, p := range d.Peers {
		if p.PublicKey == k {
			ipList = append(ipList, p.AllowedIPs...)
		}
	}
	peers := []wgtypes.PeerConfig{
		{
			PublicKey:         k,
			ReplaceAllowedIPs: true,
			AllowedIPs:        ipList,
		},
	}
	// create config var
	newConfig := wgtypes.Config{
		ReplacePeers: false,
		Peers:        peers,
	}
	// apply config to interface
	err = c.ConfigureDevice(wgInterface, newConfig)
	if err != nil {
		return err
	}
	return nil
}
