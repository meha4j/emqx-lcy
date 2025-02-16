import http.client
import netifaces as net
import ipaddress as ips
import http
import os
import base64
import json

adapterPort = os.environ['EMQX_ADAPTER_PORT']
gatePort = os.environ['PORT']
network = os.environ['NETWORK']
emqxHost = os.environ['EMQX_HOST']
emqxPort = os.environ['EMQX_PORT']
emqxUser = os.environ['EMQX_USER']
emqxPass = os.environ['EMQX_PASS']

n = ips.ip_network(network)

for iface in net.interfaces():
    if net.AF_INET in net.ifaddresses(iface):
        for link in net.ifaddresses(iface)[net.AF_INET]:
            addr = ips.ip_address(link['addr'])

            if addr in n.hosts():
                req = {
                    'name': 'exproto',
                    'server': {
                        'bind': adapterPort
                    },
                    'handler': {
                        'address': f'http://{addr}:{gatePort}'
                    }
                }

                http.client.HTTPConnection(f'http://{emqxHost}:{emqxPort}').request(
                    'PUT', '/gateways/exproto', json.dumps(req), {
                        'Content-Type': 'application/json', 
                        'Authorization': f'Basic {base64.b64encode(f'{emqxUser}:{emqxPass}')}'
                    }
                )

                exit(0)

exit(-1)
