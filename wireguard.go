package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"net"
)

type PeerJSON struct {
	PeerLoopIndex int
	PublicKey     string
	AllowedIPs    string
	Endpoint      string
	LastHandshake string
	BytesReceived int64
	BytesSent     int64
}

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

func dGetPeersJSON() []byte {
	peersJSON := []PeerJSON{}
	dRefresh()
	for i, p := range d.Peers {
		ipString := ""
		for ipi, ipn := range p.AllowedIPs {
			if ipi > 0 {
				ipString += " "
			}
			ipString += ipn.String()
		}

		newJSON := PeerJSON{
			PeerLoopIndex: i,
			PublicKey:     p.PublicKey.String(),
			Endpoint:      p.Endpoint.String(),
			AllowedIPs:    ipString,
			LastHandshake: p.LastHandshakeTime.String(),
			BytesReceived: p.ReceiveBytes,
			BytesSent:     p.TransmitBytes,
		}
		peersJSON = append(peersJSON, newJSON)
	}

	r, _ := json.MarshalIndent(peersJSON, "", "    ")
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
