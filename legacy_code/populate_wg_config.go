package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
)

type Peer struct {
	IPAddr string
	PubKey string
}

func (p *Peer) formatPeer() []byte {
	return []byte("\n[Peer]\n" +
		"PublicKey = " + p.PubKey + "\n" +
		"AllowedIPs = " + p.IPAddr + "\n")
}

var peerList = []Peer{}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func populateWireGuardConfig(w http.ResponseWriter, r *http.Request) {

	// WG_SERVER_CONFIG must contain path to file holding server config
	serverConfig, err := ioutil.ReadFile(os.Getenv("WG_SERVER_CONFIG"))
	check(err)

	ipAddr := r.URL.Query().Get("ip")
	pubKey := r.URL.Query().Get("pubkey")

	// output ipAddr and pubkey, if they exist
	// uncomment this section if you need debug
	//		fmt.Fprintln(w, "processing IP address: ", ipAddr)
	//		fmt.Fprintln(w, "processing public key: ", pubKey)

	// check via regexp if the ip and pubkey are ok
	regexMatch := false
	matchedIP, _ := regexp.MatchString("^([0-9]{1,3}\\.){3,3}[0-9]{1,3}$", ipAddr)
	matchedPubKey, _ := regexp.MatchString("^[0-9a-zA-Z\\/=+]{43,43}=$", pubKey)
	if matchedIP && matchedPubKey {
		regexMatch = true
	}

	if regexMatch {
		// search for pubkeys and IPs
		// if pubkey is found: do nothing
		// if IP is found: overwrite pubkey
		// if nothing is found: add new peer
		rewriteRequired := false
		appendRequired := true
		for _, p := range peerList {
			if p.PubKey == pubKey {
				if p.IPAddr == ipAddr+"/32" {
					fmt.Fprintln(w, "This public key is already registered: ", p)
					appendRequired = false
				}
			} else {
				if p.IPAddr == ipAddr+"/32" {
					fmt.Fprintln(w, "Changing public key for IP ", p.IPAddr)
					p.PubKey = pubKey
					appendRequired = false
					rewriteRequired = true
				}
			}
		}
		if appendRequired {
			peerList = append(peerList, Peer{IPAddr: ipAddr + "/32", PubKey: pubKey})
			rewriteRequired = true
		}

		fmt.Println("Rewrite required: ", rewriteRequired)
		fmt.Println("Peers: ", peerList)
		if rewriteRequired {
			// hardcoded wireguard server config
			wireguardConfing := serverConfig
			// collect data about all peers
			for _, p := range peerList {
				wireguardConfing = append(wireguardConfing, p.formatPeer()...)
			}
			// write new config to file
			_ = ioutil.WriteFile("/etc/wireguard/"+os.Getenv("WG_CONFIG_NAME")+".conf", wireguardConfing, 0644)
			// restart wireguard interface
			// the WG_RESTART_SCRIPT env var should contain the path to the script
			//    that restart the wg interface (wg down / wg up)
			out, err := exec.Command("/bin/bash", os.Getenv("WG_RESTART_SCRIPT")).CombinedOutput()
			if err != nil {
				fmt.Println("Error: ", err)
			}
			fmt.Printf("output: %s", out)

			// tell the client everything is OK
			fmt.Fprintln(w, "OK")
		}
	} else {
		fmt.Fprintln(w, "ERROR! IP or PubKey failed regex check.")
	}

}

func main() {

	// create HTTP listener that parses IP and PubKey and adds them to the config
	// also restart wireguard interface at the end
	http.HandleFunc("/", populateWireGuardConfig)

	http.ListenAndServe(":8081", nil)

}