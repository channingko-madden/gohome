# piwatcher

Using a PIR motion sensor and Pi camera module, take a picture and send it to Discord when motion is detected.

## Docker

```sh
docker build -t piwatcher:v1 .
```

```sh
docker run --privileged --tmpfs /dev/shm:exec -v /usr:/usr:ro -v /run/udev:/run/udev:ro -v /lib:/lib:ro -v /dev:/dev -d --name piwatcher-v1 --env DISCORD_WEBHOOK_URL=<url> --env RTSP_PATHS_CAM_SOURCE=rpiCamera --restart=always piwatcher:v1
```

- Run in privileged mode for access to GPIO and Pi camera.
- Mounts and volumes needed for accessing Pi camera and rpicam-jpeg.
