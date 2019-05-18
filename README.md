__This project is absolute crap at this point. DO NOT USE THIS!__

# WireGuard Go API `work in progress`

GoLang API that helps you manage WireGuard access on nodes. This project is really young, but I would love to work with somebody on it.

The `populate_wg_config` [binary can be downloaded](http://85.9.20.186/wireguard_go_api/populate_wg_config).

## Usage

In order to get `populate_wg_config.go` to work, you will need the following environment variables:

* `WG_SERVER_CONFIG`: path to the file that contains server configuration. You should be able to cat the config like this:

```
$ cat "$WG_SERVER_CONFIG"
[Interface]
Address = 10.10.10.100/24
ListenPort = 31337
PrivateKey = EKmMopd+xmskI9dXNtCHqS4TM0GQRmMkYh4Gs6Svm2k=
```

* `WG_CONFIG_NAME`: the name of the configuration file (the config will get written to `/etc/wireguard/${WG_CONFIG_NAME}.conf`)

* `WG_RESTART_SCRIPT`: the path to the `wg-quick` script you want to use to restart wireguard. If running wireguard in a specific network namespace, the script should look like this:
```

$ cat wireguard-restart-example.sh
ip netns exec your_namespace wg-quick down "$WG_CONFIG_NAME"
ip netns exec your_namespace wg-quick up "$WG_CONFIG_NAME"
```
