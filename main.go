package main

import (
	"fmt"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"net/http"
	"os"
	"subtle"
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
		o := ClientOutput{
			Status:  "ERROR",
			Message: "Available methods for /peers are GET, PUT, DELETE",
			Error:   "bad HTTP method",
		}
		http.Error(w, string(o.bytes()), http.StatusBadRequest)
	}
}

func privateKey(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodDelete:
		w.Write(dNewKeyPair())
	default:
		o := ClientOutput{
			Status: "ERROR",
			Message: "Use the DELETE request to generate a new key pair" +
				", or GET the /publicKey",
			Error: "bad HTTP method",
		}
		http.Error(w, string(o.bytes()), http.StatusBadRequest)
	}
}

func publicKey(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Write([]byte(`{ "PublicKey": "` + dPublicKey() + `" }`))
	default:
		o := ClientOutput{
			Status:  "ERROR",
			Message: "You can only GET the public key.",
			Error:   "bad HTTP method",
		}
		http.Error(w, string(o.bytes()), http.StatusBadRequest)
	}
}

func listenPort(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Write([]byte(`{ "ListenPort": ` + dPort() + ` }`))
	case http.MethodPut:
		w.Write(dSetPort(r.URL.Query().Get("port")))
	default:
		o := ClientOutput{
			Status:  "ERROR",
			Message: "GET current port, or PUT the new port.",
			Error:   "bad HTTP method",
		}
		http.Error(w, string(o.bytes()), http.StatusBadRequest)
	}
}

func authenticateAdmin(f http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recover Triggered: ", r)
				http.Error(w, "", 501)
			}
		}()

		// use subtle.ConstantTimeCompare() to prevent timing attack
		user, pass, _ := r.BasicAuth()
		userEnv := []byte(os.Getenv("WIREGUARD_ADMIN"))
		passEnv := []byte(os.Getenv("WIREGUARD_ADMIN_PASS"))
		authBool := (subtle.ConstantTimeCompare([]byte(user), userEnv)) && (subtle.ConstantTimeCompare([]byte(pass), passEnv))
		if authBool {
			if r.Method == http.MethodGet {
				f(w, r)
			} else {
				o := ClientOutput{
					Status:  "ERROR",
					Message: "Only GET is allowed without authentication",
					Error:   "bad credentials",
				}
				http.Error(w, string(o.bytes()), http.StatusUnauthorized)
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
