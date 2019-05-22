package main

import (
	"fmt"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"net/http"
	"os"
)

var c = &wgctrl.Client{}
var d = &wgtypes.Device{}
var wgInterface string

func peers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Write(dGetPeersJSON())
	case http.MethodPost:
		w.Write([]byte("post peers"))
	case http.MethodPut:
		err := dAddPeer(r.URL.Query().Get("pubkey"), r.URL.Query().Get("ip"))
		if err != nil {
			w.Write([]byte("public key added"))
		} else {
			http.Error(w, err.Error(), http.StatusForbidden)
		}
	case http.MethodDelete:
		err := dDeletePeer(r.URL.Query().Get("pubkey"))
		if err != nil {
			w.Write([]byte("ublic key deleted"))
		} else {
			http.Error(w, err.Error(), http.StatusForbidden)
		}
	default:
		http.Error(w, "wat", 401)
	}
}

func privateKey(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		http.Error(w, "You can only PUT the private key and get public key.", http.StatusBadRequest)
	case http.MethodPost:
		http.Error(w, "Use methods GET or PUT for the private key", http.StatusBadRequest)
	case http.MethodPut:
		w.Write([]byte("put private key"))
	case http.MethodDelete:
		http.Error(w, "PUT another private key instead of just deleting it.", http.StatusBadRequest)
	default:
		http.Error(w, "wat", 401)
	}
}

func publicKey(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Write([]byte("get public key"))
	case http.MethodPost:
		http.Error(w, "You can only PUT the private key and get public key.", http.StatusBadRequest)
	case http.MethodPut:
		http.Error(w, "You can only PUT the private key and get public key.", http.StatusBadRequest)
	case http.MethodDelete:
		http.Error(w, "What do you think this command is supposed to do?", http.StatusBadRequest)
	default:
		http.Error(w, "wat", 401)
	}
}

func authenticateAdmin(f http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, _ := r.BasicAuth()
		if user != os.Getenv("WIREGUARD_ADMIN") || pass != os.Getenv("WIREGUARD_ADMIN_PASS") {
			if r.Method == http.MethodGet {
				f(w, r)
			} else {
				http.Error(w, "Method not authorized", 401)
			}
		} else {
			f(w, r)
		}
	})
}

func main() {
	wgInterface = os.Getenv("WIREGUARD_INTERFACE")
	var wgctrlErr error
	c, wgctrlErr = wgctrl.New()
	if wgctrlErr != nil {
		fmt.Println("Wireguard error: ", wgctrlErr)
	}
	http.HandleFunc("/privateKey", authenticateAdmin(privateKey))
	http.HandleFunc("/publicKey", authenticateAdmin(publicKey))
	http.HandleFunc("/peers", authenticateAdmin(peers))
	http.ListenAndServeTLS(os.Args[1], "server.crt", "server.key", nil)
}
