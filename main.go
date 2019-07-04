package main

import (
	"crypto/subtle"
	"fmt"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"net/http"
	"os"
)

var c = &wgctrl.Client{}
var d = &wgtypes.Device{}
var dString string
var dumpFile string

func healthz(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

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

func globalMiddleware(f http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recover Triggered: ", r)
				http.Error(w, "", 501)
			}
		}()

		w.Header().Add("Content-Type", "application/json")

		// use subtle.ConstantTimeCompare() to prevent timing attack
		user, pass, _ := r.BasicAuth()
		userResult := subtle.ConstantTimeCompare([]byte(user), []byte(os.Getenv("WIREGUARD_ADMIN")))
		passResult := subtle.ConstantTimeCompare([]byte(pass), []byte(os.Getenv("WIREGUARD_ADMIN_PASS")))
		authBool := (userResult == 1) && (passResult == 1)
		if !authBool {
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
		if dumpFile != "" {
			dumpToFileRoutine()
		}
	})
}

func main() {
	dString = os.Getenv("WIREGUARD_INTERFACE")
	dumpFile = os.Getenv("WIREGUARD_DUMP_FILE")
	if dumpFile != "" {
		bootstrapFromFile()
	}
	var wgctrlErr error
	c, wgctrlErr = wgctrl.New()
	if wgctrlErr != nil {
		fmt.Println("Wireguard error: ", wgctrlErr)
	}
	http.HandleFunc("/healthz", healthz)
	http.HandleFunc("/", globalMiddleware(rootDump))
	http.HandleFunc("/privateKey", globalMiddleware(privateKey))
	http.HandleFunc("/publicKey", globalMiddleware(publicKey))
	http.HandleFunc("/listenPort", globalMiddleware(listenPort))
	http.HandleFunc("/peers", globalMiddleware(peers))
	http.ListenAndServeTLS(os.Args[1], "server.crt", "server.key", nil)
}
