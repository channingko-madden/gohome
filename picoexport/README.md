# picoexport

For each Pico W running, launch an exporter container for it.

Update the sd_picoexporter.yml file with a new name, correct target, and correct port exposed by the exporter container.

Then upload into the running prometheus container for discovery.

## Docker

```shell
docker build -t picoexport:v1 .
```

```shell
docker run -d --expose <port> --name picoexport-<name> -p <port>:<port> --env PICO_SERVER_URL=http://<PICO_IP> --env PICO_NAME=<name> --env PORT=<port> --restart=always --net=prometheus_prom_net picoexport:v1
```
