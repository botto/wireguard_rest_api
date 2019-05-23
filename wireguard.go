package main

import (
	"encoding/json"
	"fmt"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"net"
	"strconv"
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

type DeviceJSON struct {
	Name         string
	Type         string
	PublicKey    string
	FirewallMark int
	ListenPort   int
	Message      string
}

// refreshing the device is required before searching IPs
func dRefresh() {
	var err error
	d, err = c.Device(dString)
	if err != nil {
		fmt.Println("could not get wireguard device from env var WIREGUARD_INTERFACE: ", dString)
		fmt.Println("ERROR: ", err)
		panic(err)
	}
}

func dDumpData(message string) []byte {
	dRefresh()
	deviceJSON := DeviceJSON{
		Name:         d.Name,
		Type:         d.Type.String(),
		PublicKey:    d.PublicKey.String(),
		ListenPort:   d.ListenPort,
		FirewallMark: d.FirewallMark,
		Message:      message,
	}
	r, _ := json.MarshalIndent(deviceJSON, "", "    ")
	return r
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
	k, err := wgtypes.ParseKey(ks)
	if err != nil {
		return err
	}
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
	err = c.ConfigureDevice(dString, newConfig)
	return err
}

func dAddPeer(ks string, ips string) error {

	k, err := wgtypes.ParseKey(ks)
	if err != nil {
		return err
	}
	_, ip, err := net.ParseCIDR(ips)
	if err != nil {
		return err
	}

	// create config var
	peers := []wgtypes.PeerConfig{
		{
			PublicKey:         k,
			ReplaceAllowedIPs: true,
			AllowedIPs: []net.IPNet{
				*ip,
			},
		},
	}
	newConfig := wgtypes.Config{
		ReplacePeers: false,
		Peers:        peers,
	}

	// apply config to interface
	err = c.ConfigureDevice(dString, newConfig)

	return err
}

func dPublicKey() string {
	dRefresh()
	return d.PublicKey.String()
}

func dPort() string {
	dRefresh()
	return string(d.ListenPort)
}

func dNewKeyPair() error {
	dRefresh()

	privateKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return err
	}

	newConfig := wgtypes.Config{
		PrivateKey: &privateKey,
	}

	// apply config to interface
	err = c.ConfigureDevice(dString, newConfig)
	return err
}

func dSetPort(ps string) error {

	p, err := strconv.Atoi(ps)
	if err != nil {
		return err
	}

	newConfig := wgtypes.Config{
		ListenPort: &p,
	}

	// apply config to interface
	err = c.ConfigureDevice(dString, newConfig)
	return err
}
