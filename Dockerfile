FROM alpine:edge
RUN echo "http://dl-cdn.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories
RUN apk add --no-cache bash iptables wireguard-tools iproute2 openssl

COPY wireguard_rest_api /opt/
COPY start.sh /start.sh

ENTRYPOINT ["/bin/bash"]
CMD ["/start.sh"]
