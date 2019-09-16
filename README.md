# Image Proxy

### Configure service as a daemon

```
sudo vim /etc/systemd/system/image-proxy.service
```
```
[Unit]
Description=image-proxy
After=network.target

[Service]
#User=<changeMe>
#Group=<changeMe>
ExecStart=/opt/image-proxy/image-proxy -listen=localhost:8085
Restart=always

[Install]
WantedBy=multi-user.target
```
```
sudo systemctl enable image-proxy
sudo systemctl start image-proxy
sudo systemctl status image-proxy
```
