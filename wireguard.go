package main

import (
	"encoding/json"
	"fmt"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"net"
	"strconv"
	"sync"
)

var mutex = &sync.RWMutex{}

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

type ClientOutput struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

func (o *ClientOutput) bytes() []byte {
	jsonData, err := json.MarshalIndent(o, "", "    ")
	if err != nil {
		fmt.Println("Error parsing JSON: ", o)
		panic(err)
	}
	return jsonData
}

// refreshing the device is required before searching IPs
func dRefresh() {
	var err error
	d, err = c.Device(dString)
	if err != nil {
		fmt.Println("could not get wireguard device from env var WIREGUARD_INTERFACE:", dString)
		fmt.Println("ERROR: ", err)
		panic(err)
	}
}

func dDumpData(message string) []byte {
	mutex.RLock()
	dRefresh()
	deviceJSON := DeviceJSON{
		Name:         d.Name,
		Type:         d.Type.String(),
		PublicKey:    d.PublicKey.String(),
		ListenPort:   d.ListenPort,
		FirewallMark: d.FirewallMark,
		Message:      message,
	}
	mutex.RUnlock()
	r, _ := json.MarshalIndent(deviceJSON, "", "    ")
	return r
}

func dGetPeersJSON() []byte {
	peersJSON := []PeerJSON{}
	mutex.RLock()
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
	mutex.RUnlock()

	r, _ := json.MarshalIndent(peersJSON, "", "    ")
	return r
}

func dDeletePeer(ks string) []byte {
	o := ClientOutput{
		Status:  "OK",
		Message: "peer " + ks + " deleted",
	}
	k, err := wgtypes.ParseKey(ks)
	if err != nil {
		o.Status = "ERROR"
		o.Message = "bad public key"
		o.Error = err.Error()
		return o.bytes()
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
	mutex.Lock()
	err = c.ConfigureDevice(dString, newConfig)
	mutex.Unlock()

	if err != nil {
		o.Status = "ERROR"
		o.Message = "Peer deletion failed"
		o.Error = err.Error()
	}
	return o.bytes()
}

func dAddPeer(ks string, ips string) []byte {

	o := ClientOutput{
		Status:  "OK",
		Message: "peer " + ks + " added",
	}
	k, err := wgtypes.ParseKey(ks)
	if err != nil {
		o.Status = "ERROR"
		o.Message = "bad public key"
		o.Error = err.Error()
		return o.bytes()
	}
	_, ip, err := net.ParseCIDR(ips)
	if err != nil {
		o.Status = "ERROR"
		o.Message = "bad CIDR; use GET variable ip"
		o.Error = err.Error()
		return o.bytes()
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
	mutex.Lock()
	err = c.ConfigureDevice(dString, newConfig)
	mutex.Unlock()

	if err != nil {
		o.Status = "ERROR"
		o.Message = "wg ConfigureDevice failed"
		o.Error = err.Error()
	}
	return o.bytes()
}

func dPublicKey() string {
	mutex.RLock()
	defer mutex.RUnlock()
	dRefresh()
	return d.PublicKey.String()
}

func dPort() string {
	mutex.RLock()
	defer mutex.RUnlock()
	dRefresh()
	return strconv.Itoa(d.ListenPort)
}

func dNewKeyPair() []byte {
	o := ClientOutput{
		Status:  "OK",
		Message: "GET public key at /publicKey",
	}

	privateKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		o.Status = "ERROR"
		o.Message = "wgtypes.GeneratePrivateKey() failed"
		o.Error = err.Error()
		return o.bytes()
	}

	newConfig := wgtypes.Config{
		PrivateKey: &privateKey,
	}

	// apply config to interface
	mutex.Lock()
	err = c.ConfigureDevice(dString, newConfig)
	mutex.Unlock()
	if err != nil {
		o.Status = "ERROR"
		o.Message = "wg ConfigureDevice failed"
		o.Error = err.Error()
	}
	return o.bytes()
}

func dSetPort(ps string) []byte {

	o := ClientOutput{
		Status:  "OK",
		Message: "port set to " + ps,
	}
	p, err := strconv.Atoi(ps)
	if err != nil {
		o.Status = "ERROR"
		o.Message = "was your port an int?"
		o.Error = err.Error()
		return o.bytes()
	}

	newConfig := wgtypes.Config{
		ListenPort: &p,
	}

	// apply config to interface
	mutex.Lock()
	err = c.ConfigureDevice(dString, newConfig)
	mutex.Unlock()
	if err != nil {
		o.Status = "ERROR"
		o.Message = "wg ConfigureDevice failed"
		o.Error = err.Error()
	}
	return o.bytes()
}

func bootstrapFromFile() {
	dumpFileJSON := getFromFile()
	privateKey, err := wgtypes.ParseKey(dumpFileJSON.PrivateKey)
	if err != nil {
		fmt.Println("Could not parse private key")
		panic(err)
	}
	peers := []wgtypes.PeerConfig{}
	for _, dumpFilePeerJSON := range dumpFileJSON.Peers {
		publicKey, err := wgtypes.ParseKey(dumpFilePeerJSON.PublicKey)
		if err != nil {
			fmt.Println("Could not parse the public key of one peer")
			panic(err)
		}
		_, allowedIP, err := net.ParseCIDR(dumpFilePeerJSON.AllowedIPs)
		if err != nil {
			fmt.Println("One public key not added to peer list.")
			fmt.Println("Could not parse the IP of one peer. This code does not accept multiple subnets.")
			fmt.Println("If you need this functionality, fix the code at https://gitlab.com/gun1x/wireguard_rest_api")
			fmt.Println("I will accept your push request.")
		} else {
			peers = append(peers, wgtypes.PeerConfig{
				PublicKey: publicKey,
				AllowedIPs: []net.IPNet{
					*allowedIP,
				},
			})
		}
	}
	newConfig := wgtypes.Config{
		PrivateKey: &privateKey,
		Peers:      peers,
	}
	// apply config to interface
	err = c.ConfigureDevice(dString, newConfig)
	if err != nil {
		fmt.Println("could not apply wireguard config")
		panic(err)
	}
}
