# Get Started 

The Docker image is hosted on the GitHub container registry and is ready to be used:

```
docker pull ghcr.io/blabtm/emqx-gate:latest
```

The following configuration properties are required in order to properly setup extension:

- PORT - host port on which service will be started (default: 9001)
- EMQX_ADAPTER_HOST - EMQX hostname at which ConnectionAdapter will be started (default: emqx)
- EMQX_ADAPTER_PORT - EMQX port number at which ConnectionAdapter will be started (default: 9100)

Below is a minimum viable stack file (example/compose.yaml):

```yaml
networks:
  stage:
    name: stage
    driver: bridge

services:
  emqx:
    image: ghcr.io/blabtm/emqx:latest
    ports:
      - 1883:1883
      - 18083:18083
      - 20041:20041
    networks:
      - stage
  emqx-gate:
    image: ghcr.io/blabtm/emqx-gate:latest
    environment:
      PORT: 9001
      EMQX_ADAPTER_HOST: emqx
      EMQX_ADAPTER_PORT: 9100
    networks:
      - stage
```

> When the server starts, see the log output to determine the `MACHINE` address.

To add gateway instance:

1. Login to EMQX Dashboard
2. Go to Management -> Gateways
3. Click `Setup` opposite the ExProto
4. Configure the gateway:
    - gRPC ConnectionAdapter - Bind: `0.0.0.0:{EMQX_ADAPTER_PORT}`
    - gRPC ConnectionHandler - Server: `http://{MACHINE}:{PORT}`
5. Go `Next`
6. Setup `default` listener:
    - Type: `tcp`
    - Bind: `20041`
7. Go `Update` -> `Next` -> `Enable`

Enjoy!
