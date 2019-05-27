package main

import (
	"encoding/json"
	"golang.zx2c4.com/wireguard/wgctrl"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var defaultTestConfig = `Use the following script to configure your device before testing:
sudo ip link add dev testingrestapi type wireguard
sudo bash -c 'wg set testingrestapi listen-port 1338 private-key <(echo "KL2U8h7HPit1vUMaTcMgnPwRgWwnYoOT4iWT3obYx0Y=")'
sudo ip link set up dev testingrestapi
`

func testMiddleware(t *testing.T, m, u string, handler http.HandlerFunc, auth bool) (int, []byte) {
	dString = "testingrestapi"
	var wgctrlErr error
	c, wgctrlErr = wgctrl.New()
	if wgctrlErr != nil {
		t.Fatal("Wireguard error: ", wgctrlErr)
	}
	req, err := http.NewRequest(m, u, nil)
	if auth {
		os.Setenv("WIREGUARD_ADMIN", "userFoo")
		os.Setenv("WIREGUARD_ADMIN", "userBar")
		req.Header.Set("Authorization", "userFoo passBar")
	}
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr.Code, rr.Body.Bytes()
}

func TestHealthCheck(t *testing.T) {
	status, response := testMiddleware(t, "GET", "/healthz", http.HandlerFunc(healthz), false)
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
	status, response := testMiddleware(t, "GET", "/", globalMiddleware(rootDump), false)
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

	// test bad authentication
	status, response := testMiddleware(t, "PUT", "/healthz", http.HandlerFunc(globalMiddleware(healthz)), false)
	if status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusUnauthorized)
		t.Error("This is the response:", string(response))
	}
	responseJSON := ClientOutput{}
	err := json.Unmarshal(response, &responseJSON)
	if err != nil {
		t.Fatal("Reponse failed unmarshal:", err, "\n the response was:\n,", response)
	}
	if responseJSON.Status != "ERROR" {
		t.Error("The status of the JSON is not ERROR, it is", responseJSON.Status)
	}

	// test good authentication
	status, response = testMiddleware(t, "PUT", "/healthz", http.HandlerFunc(globalMiddleware(healthz)), true)
	if status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
		t.Error("This is the response:", string(response))
	}

}

func TestKeyGeneration(t *testing.T) {
	_, response := testMiddleware(t, "DELETE", "/privateKey", http.HandlerFunc(globalMiddleware(privateKey)), true)
	responseJSON := ClientOutput{}
	err := json.Unmarshal(response, &responseJSON)
	if err != nil {
		t.Fatal("Reponse failed unmarshal:", err, "\n the response was:\n,", response)
	}
	if responseJSON.Status != "OK" {
		t.Error("Failed to generate new private key. Response:\n", string(response))
	}
	_, response = testMiddleware(t, "GET", "/publicKey", http.HandlerFunc(globalMiddleware(publicKey)), true)
	if string(response) == `{ "PublicKey": "U2RqZbIkiJ4Hel9HvnX8oz8m1tDf2k2MBRrrZG0H5WY=" }` {
		t.Fatal("The public key is still the same:\n", string(response))
	}
}

func TestPutPeer(t *testing.T) {
	url := `peers?pubkey=W4GUbzUthA5atPF1aFB1dWTZYT6hrNko1qxRu1GcqjM%3D&ip=10.200.123.123%2F32`
	status, response := testMiddleware(t, "PUT", url, http.HandlerFunc(globalMiddleware(peers)), true)
	if status != http.StatusOK {
		t.Error("This is the response:", string(response))
		t.Fatalf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	responseJSON := ClientOutput{}
	err := json.Unmarshal(response, &responseJSON)
	if err != nil {
		t.Fatal("Reponse failed unmarshal:", err, "\n the response was:\n,", response)
	}
	if responseJSON.Status != "OK" {
		t.Error("The response status is not OK:\n", string(response))
	}
}

func TestGetPeers(t *testing.T) {
	status, response := testMiddleware(t, "GET", "/peers", http.HandlerFunc(globalMiddleware(peers)), false)
	if status != http.StatusOK {
		t.Error("This is the response:", string(response))
		t.Fatalf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	responseJSON := []PeerJSON{}
	err := json.Unmarshal(response, &responseJSON)
	if err != nil {
		t.Fatal("Reponse failed unmarshal:", err, "\n the response was:\n,", response)
	}
	foundPeer := false
	for _, p := range responseJSON {
		if p.PublicKey == "W4GUbzUthA5atPF1aFB1dWTZYT6hrNko1qxRu1GcqjM=" &&
			p.AllowedIPs == "10.200.123.123/32" {
			foundPeer = true
		}
	}
	if !foundPeer {
		t.Fatal("Didn't find the Peer:\n", responseJSON)
	}

}

func TestDeletePeer(t *testing.T) {
	url := `peers?pubkey=W4GUbzUthA5atPF1aFB1dWTZYT6hrNko1qxRu1GcqjM%3D`
	status, response := testMiddleware(t, "DELETE", url, http.HandlerFunc(globalMiddleware(peers)), true)
	if status != http.StatusOK {
		t.Error("This is the response:", string(response))
		t.Fatalf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	responseJSON := ClientOutput{}
	err := json.Unmarshal(response, &responseJSON)
	if err != nil {
		t.Fatal("Reponse failed unmarshal:", err, "\n the response was:\n,", response)
	}
	if responseJSON.Status != "OK" {
		t.Error("The response status is not OK:\n", string(response))
	}
}
