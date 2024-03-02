#!/usr/bin/sh

cp /openapp/frp/frpc.toml /etc/frp/frpc.toml

/usr/bin/frpc -c /etc/frp/frpc.toml
