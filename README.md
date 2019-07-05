# [WireGuard REST API](https://gitlab.com/gun1x/wireguard_rest_api)

*If you need any modifications/improvements to this project, please let me know.*

This webserver allows you to control one wireguard interface located on the server. It allows you to get/set configuration of the WireGuard device, and also about the Peers configured on the device.

All the GET commands do not require authentication. All the PUT/DELETE commands require authentication.

Currently one server can manage only one wireguard device. This could change in the future, if anybody needs one server that manages multiple interfaces.

## Usage examples:

### get all information
```bash
$ curl -k "https://server:31337/"
{
    "Name": "internal",
    "Type": "Linux kernel",
    "PublicKey": "YTV4cu1ucfKDej2SqRX91ldU7uT9S+s96/RFJmuwyjg=",
    "FirewallMark": 0,
    "ListenPort": 1337,
    "Message": "available commands: /peers /privateKey /publicKey /listenPort"
}
```

### get all the peers
```bash
$ curl -k -G --user "user:pass" "https://server:31337/peers"
[
    {
        "PeerLoopIndex": 0,
        "PublicKey": "xzSmPlkbxCslHIZhon/fJZ7pjWrP4HSlSh2h1He/BCg=",
        "AllowedIPs": "10.100.0.3/32",
        "Endpoint": "192.168.121.39:58917",
        "LastHandshake": "2019-05-21 14:12:37.051697687 +0000 UTC",
        "BytesReceived": 264056,
        "BytesSent": 322256
    },
    {
        "PeerLoopIndex": 1,
        "PublicKey": "yKtnb6UgriaIm1Xi9DQ+BcTwPFITlQLQ9M2BmxMCrhs=",
        "AllowedIPs": "10.100.0.2/32",
        "Endpoint": "192.168.121.148:34790",
        "LastHandshake": "2019-05-21 13:31:12.104693013 +0000 UTC",
        "BytesReceived": 119780,
        "BytesSent": 143476
    },
    {
        "PeerLoopIndex": 2,
        "PublicKey": "+wOFGJo7bjuGhf/nMZ4IB9bNr475x2GURy6089UkJHM=",
        "AllowedIPs": "10.100.0.5/32",
        "Endpoint": "192.168.121.12:49446",
        "LastHandshake": "2019-05-21 13:31:12.186561752 +0000 UTC",
        "BytesReceived": 92740,
        "BytesSent": 97424
    }
]
```

### get only public key
```bash
 $ curl -k "https://server:31337/publicKey"
{"PublicKey": "YTV4cu1ucfKDej2SqRX91ldU7uT9S+s96/RFJmuwyjg="}
```

### change the private key
```bash
 $ curl -k "https://server:31337/privateKey"
{
    "status": "ERROR",
    "message": "Use the DELETE request to generate a new key pair, or GET the /publicKey",
    "error": "bad HTTP method"
}

 $ curl -X DELETE -k -G --user "user:pass"  "https://server:31337/privateKey"
{
    "status": "OK",
    "message": "GET public key at /publicKey"
}

 $ curl -k "https://server:31337/publicKey"           
{ "PublicKey": "daxw/ElaZgOypvBNbAL8es4DotrsxnQagwnFq7Ch5DU=" }
```

## Running with Docker

The server will run by default on port 31337, but you can change that in [docker start.sh](https://gitlab.com/gun1x/wireguard_rest_api/blob/master/start.sh).

```
docker pull "registry.gitlab.com/gun1x/wireguard_rest_api"
docker rm --force "wireguard_rest_api"
docker run \
  --rm \
  --net=host \
  --cap-add NET_ADMIN \
  --env WIREGUARD_INTERFACE=wg1337 \
  --env WIREGUARD_ADMIN=user \
  --env WIREGUARD_ADMIN_PASS=pass \
  --env WIREGUARD_DUMP_FILE="/root/wireguard_dump" \
  --volume "/root/wireguard_dump:/root/wireguard_dump:rw" \
  --name wireguard_rest_api \
  -it "registry.gitlab.com/gun1x/wireguard_rest_api"
```

`wireguard_dump` can be removed if you do not need reboot persistance. Be aware that `wireguard_dump` will save the private key on the disk.

## Running without docker

The server can run without Docker, as long as it has the environment variables. Let me know if you consider you need better documnetation for this.

The interface must be created before the server is started, as detailed in the [docker start.sh](https://gitlab.com/gun1x/wireguard_rest_api/blob/master/start.sh), which is a good example of how to run the server, and should work on any distribution.
