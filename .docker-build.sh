#!/bin/bash

# I use this to build and push.
CGO_ENABLED=0 go build
docker build -t registry.gitlab.com/gun1x/wireguard_go_api .
docker push registry.gitlab.com/gun1x/wireguard_go_api

