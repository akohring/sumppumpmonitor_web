# Local setup
```
go build && ./sumppumpmonitor
```

# Build for raspberry pi
```
sudo apt-get install gcc-arm*
env GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=1 CC=arm-linux-gnueabi-gcc go build
```

# Startup script
```
sudo chmod +x /etc/init.d/sumppumpmonitorweb
sudo update-rc.d sumppumpmonitorweb defaults 100
sudo service sumppumpmonitorweb start
```

# sudo crontab -e
```
*/1 * * * * /opt/networkmonitor/run.sh > /dev/null 2>&1
*/1 * * * * /opt/sumppumpmonitor/run.sh > /dev/null 2>&1
*/1 * * * * /opt/sumppumpmonitor/service.py > /dev/null 2>&1
*/1 * * * * /opt/powerboostmonitor/powerboost.sh > /dev/null 2>&1
0 0 * * * curl -X DELETE http://localhost:8080/pitlevels > /dev/null 2>&1
```