package main

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

type DumpFileJSON struct {
	PrivateKey string
	Peers      []DumpFilePeerJSON
}

type DumpFilePeerJSON struct {
	PublicKey  string
	AllowedIPs string
}

var dumpToFileTime int64

func dumpToFileRoutine() {
	go func() {
		mutex.Lock()
		dumpToFileTime = time.Now().UnixNano()
		mutex.Unlock()
		expectedTimestamp := dumpToFileTime
		time.Sleep(3 * time.Second)
		if expectedTimestamp == dumpToFileTime {
			log.Println("calling dumpToFile()")
			dumpToFile()
		} else {
			log.Println("not calling dumpToFile()")
		}
	}()
}

func dumpToFile() error {
	dumpFilePeersJSON := []DumpFilePeerJSON{}
	dumpFileJSON := DumpFileJSON{}

	dRefresh()

	for _, p := range d.Peers {
		ipString := ""

		for ipi, ipn := range p.AllowedIPs {
			if ipi > 0 {
				ipString += " "
			}
			ipString += ipn.String()
		}

		newJSON := DumpFilePeerJSON{
			PublicKey:  p.PublicKey.String(),
			AllowedIPs: ipString,
		}

		dumpFilePeersJSON = append(dumpFilePeersJSON, newJSON)
	}
	dumpFileJSON.PrivateKey = d.PrivateKey.String()
	dumpFileJSON.Peers = dumpFilePeersJSON

	f, err := os.Create(dumpFile)
	if err != nil {
		return err
	}
	o, _ := json.MarshalIndent(dumpFileJSON, "", "  ")
	_, err = f.Write(o)
	if err != nil {
		f.Close()
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}

	log.Println("wrote to file", dumpFile)

	return nil
}
