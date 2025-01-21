#!/bin/bash

docker service create                           \
  --name gate                                   \
  --network nano                                \
  --config source=gate,target=/etc/config.yaml  \
  --secret source=gate,target=/etc/secret.yaml  \
  -e CONFIG=/etc/config.yaml                    \
  -e SECRET=/etc/secret.yaml                    \
  ghcr.io/blabtm/emqx-gate:0.1.0-beta
