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
	message := "available commands: /peers /privateKey /publicKey /listenPort"
	w.Write(dDumpData(message))
}

func peers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Write(dGetPeersJSON())
	case http.MethodPut:
		w.Write(dAddPeer(r.URL.Query().Get("pubkey"), r.URL.Query().Get("ip")))
	case http.MethodDelete:
		w.Write(dDeletePeer(r.URL.Query().Get("pubkey")))
	default:
		http.Error(w, `{
			"status": "ERROR",
			"message": "Available methods for /peers are GET, PUT, DELETE"
			"error": "bad method"
		}`, http.StatusBadRequest)
	}
}

func privateKey(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodDelete:
		w.Write(dNewKeyPair())
	default:
		http.Error(w, `{
			"status": "ERROR",
			"message": "Use the DELETE request to generate a new key pair, or GET the /publicKey"
			"error": "bad method"
		}`, http.StatusBadRequest)
	}
}

func publicKey(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Write([]byte(`{"PublicKey": "` + dPublicKey() + `"}`))
	default:
		http.Error(w, `{
			"status": "ERROR",
			"message": "You can only GET the public key."
			"error": "bad method"
		}`, http.StatusBadRequest)
	}
}

func listenPort(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Write([]byte(`{"ListenPort": ` + dPort() + `}`))
	case http.MethodPut:
		w.Write(dSetPort(r.URL.Query().Get("pubkey")))
	default:
		http.Error(w, `{
			"status": "ERROR",
			"message": "GET current port, or PUT the new port."
			"error": "bad method"
		}`, http.StatusBadRequest)
	}
}

func authenticateAdmin(f http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); r != nil {
				fmt.Println("Recover Triggered: ", rec)
				http.Error(w, "", 501)
			}
		}()

		user, pass, _ := r.BasicAuth()
		if user != os.Getenv("WIREGUARD_ADMIN") || pass != os.Getenv("WIREGUARD_ADMIN_PASS") {
			if r.Method == http.MethodGet {
				f(w, r)
			} else {
				http.Error(w, `{
					"status": "ERROR",
					"message": "Only GET is allowed without authentication",
					"error": "bad credentials"
				}`, http.StatusUnauthorized)
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
