# `WORK IN PROGRESS`

for a teaser, head over to https://gitlab.com/gun1x/wireguard-mariadb-auth

## routing

Currently this app will support only ONE wg interface per server. This might change in the future, but I am under time pressure to deliver a project so I have time to add this flexibility. This means that the device `d` is global at the moment.

However in the future the router should be changed with something like mux so that we can use something like `/peers/{id:[a-z0-9]+}` to get the interface name, and `d` will change on every request according to device name.


## run example

```
docker pull "registry.gitlab.com/gun1x/wireguard_go_api"
docker rm --force "wireguard_go_api"
docker run \
  --rm \
  --net=host \
  --cap-add NET_ADMIN \
  --env WIREGUARD_INTERFACE=wg1337 \
  --env WIREGUARD_ADMIN=user \
  --env WIREGUARD_ADMIN_PASS=pass \
  --name wireguard_go_api \
  -it "registry.gitlab.com/gun1x/wireguard_go_api"
```
