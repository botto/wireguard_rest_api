package main

import (
	"errors"
	"fmt"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"net"
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
