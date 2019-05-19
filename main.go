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
		w.Write([]byte("get peers"))
	case http.MethodPost:
		w.Write([]byte("post peers"))
	case http.MethodPut:
		w.Write([]byte("put peers"))
	case http.MethodDelete:
		w.Write([]byte("delete peers"))
	default:
		http.Error(w, "wat", 401)
	}
}

func privateKey(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Write([]byte("get private key"))
	case http.MethodPost:
		w.Write([]byte("post private key"))
	case http.MethodPut:
		w.Write([]byte("put private key"))
	case http.MethodDelete:
		w.Write([]byte("delete private key"))
	default:
		http.Error(w, "wat", 401)
	}
}

func publicKey(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Write([]byte("get public key"))
	case http.MethodPost:
		w.Write([]byte("post public key"))
	case http.MethodPut:
		w.Write([]byte("put public key"))
	case http.MethodDelete:
		w.Write([]byte("delete public key"))
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
