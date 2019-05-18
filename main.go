package main

import (
	"fmt"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"net/http"
	"os"
	"sync"
)

var userKeyMap = make(map[string]wgtypes.Key)
var mutex = &sync.Mutex{}
var c = &wgctrl.Client{}
var d = &wgtypes.Device{}
var wgInterface string
var err error

func dumpAllData(w http.ResponseWriter, r *http.Request) {
}

func addPublicKey(w http.ResponseWriter, r *http.Request) {
}

func main() {
	wgInterface = os.Getenv("WIREGUARD_INTERFACE")
	c, err = wgctrl.New()
	if err != nil {
		fmt.Println("Wireguard error: ", err)
	}
	http.HandleFunc("/addPublicKey", addPublicKey)
	http.HandleFunc("/dumpAllData", dumpAllData)
	http.ListenAndServeTLS(os.Args[1], "server.crt", "server.key", nil)
}
