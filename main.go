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
var dString string

func peers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Write(dGetPeersJSON())
	case http.MethodPost:
		http.Error(w, "Available methods: GET, PUT, DELETE", http.StatusBadRequest)
	case http.MethodPut:
		err := dAddPeer(r.URL.Query().Get("pubkey"), r.URL.Query().Get("ip"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			w.Write([]byte("OK"))
		}
	case http.MethodDelete:
		err := dDeletePeer(r.URL.Query().Get("pubkey"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
		} else {
			w.Write([]byte("peer deleted"))
		}
	default:
		http.Error(w, "Available methods: GET, PUT, DELETE", http.StatusBadRequest)
	}
}

func privateKey(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPut:
		w.Write([]byte("This should actually work, but I didn't write the code yet."))
	default:
		http.Error(w, "You can only PUT the private key and GET public key.", http.StatusBadRequest)
	}
}

func publicKey(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Write([]byte(dPublicKey()))
	default:
		http.Error(w, "You can only PUT the private key and GET public key.", http.StatusBadRequest)
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
	dString = os.Getenv("WIREGUARD_INTERFACE")
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
