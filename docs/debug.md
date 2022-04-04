# Debugging

When debugging image configuration it is hard to interate fast.
One of the options is this:

1. SSH into device which is connected to the monitor
2. Run test container:
```
docker run -it --privileged -v /tmp:/tmp --entrypoint bash -p 8081:8081 quay.io/unikiosk/unikiosk 
```
3. In the container modify `/root/start` to one of these:
```
su  -c "/tmp/screen"
```
or if you working with chromium issue:
```
su -c "chromium --no-sandbox  --disable-gpu-sandbox --no-sanbox --app=https://synpse.net"
```
4. Open new terminal close with unikiosk project dir:
```
make build-screen-arm64 ; synpse cp ./release/screen-arm64 device_name:/tmp/screen
```

This allows, build, rerun, build behavior.