package main

import (
	"encoding/json"
	"golang.zx2c4.com/wireguard/wgctrl"
	"net/http"
	"net/http/httptest"
	"testing"
)

var defaultTestConfig = `Use the following script to configure your device before testing:
ip link add dev testingrestapi type wireguard
wg set testingrestapi listen-port 1338 private-key <(echo "KL2U8h7HPit1vUMaTcMgnPwRgWwnYoOT4iWT3obYx0Y=")
ip link set up dev testingrestapi
`

func testMiddleware(t *testing.T, m, u string, handler http.HandlerFunc) (int, []byte) {
	dString = "testingrestapi"
	var wgctrlErr error
	c, wgctrlErr = wgctrl.New()
	if wgctrlErr != nil {
		t.Fatal("Wireguard error: ", wgctrlErr)
	}
	req, err := http.NewRequest(m, u, nil)
	t.Log(req)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr.Code, rr.Body.Bytes()
}

func TestHealthCheck(t *testing.T) {
	status, response := testMiddleware(t, "GET", "/healthz", http.HandlerFunc(healthz))
	if status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	if string(response) != "OK" {
		t.Errorf("handler returned unexpected body: got %v want %v",
			response, "OK")
	}
}

func TestDump(t *testing.T) {
	responseJSON := DeviceJSON{}
	status, response := testMiddleware(t, "GET", "/", globalMiddleware(rootDump))
	if status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	err := json.Unmarshal(response, &responseJSON)
	if err != nil {
		t.Error(defaultTestConfig)
		t.Fatal("Reponse failed unmarshal:", err, "\n the response was:\n,", response)
	}
	if responseJSON.PublicKey != "U2RqZbIkiJ4Hel9HvnX8oz8m1tDf2k2MBRrrZG0H5WY=" ||
		responseJSON.ListenPort != 1338 {
		t.Fatal(defaultTestConfig)
	}
}

func TestAuthentication(t *testing.T) {
	status, response := testMiddleware(t, "PUT", "/peers", http.HandlerFunc(globalMiddleware(peers)))
	if status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusUnauthorized)
		t.Error("This is the response:", string(response))
	}
}

func TestPutPeer(t *testing.T) {
}
