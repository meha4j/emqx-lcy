#!/bin/sh

ADDRESS=$(ip route | grep "${NETWORK}" | grep -oE "[0-9]+.[0-9]+.[0-9]+.[0-9]+" | tail -n1)

if [ -z "${ADDRESS}" ]; then
    echo "We are not at the right place..."
else
    res=$(curl http://${EMQX_HOST}:${EMQX_PORT}/api/v5/gateways/exproto \
        -X PUT \
        --write-out "%{http_code}" \
        --user "${EMQX_USER}:${EMQX_PASS}" \
        --header "Content-Type: application/json" \
        --data-binary '{
            "name": "exproto",
            "server": {
                "bind": "'"${EMQX_ADAPTER_PORT}"'"
            },
            "handler": {
                "address": "http://'"${ADDRESS}"':'"${PORT}"'"
            }
        }')

    echo $res

    if [ "$res" -eq "204" ]; then
        haproxy -f /usr/local/etc/haproxy/haproxy.cfg
    fi
fi
