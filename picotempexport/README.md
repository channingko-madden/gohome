# picotempexport

## Docker

```shell
docker build -t picotempexport:v1 .
```

```shell
docker run -d --name picotempexport-v1 -p 3030:3030 --env PICO_SERVER_URL=http://<PICO_IP>
--restart=always --net=prometheus_prom_net picotempexport:v1
```
