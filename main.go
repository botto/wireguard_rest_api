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

func rootDump(w http.ResponseWriter, r *http.Request) {
	message := "available methods: /peers /privateKey /publicKey /listenPort"
	w.Write(dDumpData(message))
}

func peers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Write(dGetPeersJSON())
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
	case http.MethodDelete:
		err := dNewKeyPair()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			w.Write([]byte("OK; GET /publicKey "))
		}
	default:
		http.Error(w, "Use the DELETE request to generate a new key pair, or GET the /publicKey", http.StatusBadRequest)
	}
}

func publicKey(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Write([]byte(dPublicKey()))
	default:
		http.Error(w, "You can only GET the public key.", http.StatusBadRequest)
	}
}

func listenPort(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Write([]byte(dPort()))
	case http.MethodPut:
		err := dSetPort(r.URL.Query().Get("pubkey"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			w.Write([]byte("port changed"))
		}
	default:
		http.Error(w, "GET current port, or PUT the new port.", http.StatusBadRequest)
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
	http.HandleFunc("/", rootDump)
	http.HandleFunc("/privateKey", authenticateAdmin(privateKey))
	http.HandleFunc("/publicKey", authenticateAdmin(publicKey))
	http.HandleFunc("/listenPort", authenticateAdmin(listenPort))
	http.HandleFunc("/peers", authenticateAdmin(peers))
	http.ListenAndServeTLS(os.Args[1], "server.crt", "server.key", nil)
}
