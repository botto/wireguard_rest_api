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

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {

	// WG_SERVER_CONFIG should contain server config
	serverConfig, err := ioutil.ReadFile(os.Getenv("WG_SERVER_CONFIG"))
	check(err)

	peerList := []Peer{}

	// create HTTP listener that parses IP and PubKey and adds them to the config
	// also restart wireguard interface at the end
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		ipAddr := r.URL.Query().Get("ip")
		pubKey := r.URL.Query().Get("pubkey")

		// output ipAddr and pubkey, if they exist
		// uncomment this section if you need debug
		//		fmt.Fprintln(w, "processing IP address: ", ipAddr)
		//		fmt.Fprintln(w, "processing public key: ", pubKey)

		// check via regexp if the ip and pubkey are ok
		regexMatch := false
		matchedIP, _ := regexp.MatchString("^([0-9]{1,3}\\.){3,3}[0-9]{1,3}$", ipAddr)
		matchedPubKey, _ := regexp.MatchString("^[0-9a-zA-Z\\/=]{43,43}=$", pubKey)
		if matchedIP && matchedPubKey {
			regexMatch = true
		} else {
			fmt.Fprintln(w, "ERROR! IP or PubKey failed regex check.")
		}

		// find if that combination if the IP or the pubkey already exist.
		// if not, inject them
		// if yes, return message to user
		peerExists := false
		for _, p := range peerList {
			if p.IPAddr == ipAddr+"/24" || p.PubKey == pubKey {
				peerExists = true
				fmt.Fprintln(w, "ERROR! Peer already registered: ", p)
			}
		}

		if !peerExists && regexMatch {
			peerList = append(peerList, Peer{IPAddr: ipAddr + "/24", PubKey: pubKey})

			// hardcoded wireguard server config
			wireguardConfing := serverConfig

			// collect data about all peers
			for _, p := range peerList {
				wireguardConfing = append(wireguardConfing, p.formatPeer()...)
			}

			// write new config to file
			_ = ioutil.WriteFile(os.Getenv("WG_CONFIG_NAME"), wireguardConfing, 0644)

			// restart wireguard interface
			// the WG_RESTART_SCRIPT env var should contain the path to the script
			//    that restart the wg interface (wg down / wg up)
			out, err := exec.Command(os.Getenv("WG_RESTART_SCRIPT")).Output()
			if err != nil {
				fmt.Println("Error: ", err)
			}
			fmt.Printf("output: %s", out)

			// tell the client everything is OK
			fmt.Fprintln(w, "OK")
		}

	})

	http.ListenAndServe(":8081", nil)

}
