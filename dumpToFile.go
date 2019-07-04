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

// this makes the script wait for 3 seconds before writing
// this will reduce the number of writes, in case of multiple requests
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

func getFromFile() DumpFileJSON {
	dumpFileJSON := DumpFileJSON{}

	f, err := os.Open(dumpFile)
	if err != nil {
		log.Println("error opening", dumpFile)
		panic(err)
	}
	fi, err := f.Stat()
	if err != nil {
		log.Println("Could not obtain stat, handle error")
		panic(err)
	}

	dataFromFile := make([]byte, fi.Size())
	_, err = f.Read(dataFromFile)
	if err != nil {
		log.Println("error reading", dumpFile, err)
		panic(err)
	}
	json.Unmarshal(dataFromFile, &dumpFileJSON)

	return dumpFileJSON
}
