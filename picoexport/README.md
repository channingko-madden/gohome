# picoexport

For each Pico W running, launch an exporter container for it.

Update the sd_picoexporter.yml file with a new name and the correct target, then
upload into the running prometheus container for discovery.

## Docker

```shell
docker build -t picoexport:<name> .
```

```shell
docker run -d --name picoexport-<name> -p 3030:3030 --env PICO_SERVER_URL=http://<PICO_IP> PICO_NAME=<name> --restart=always --net=prometheus_prom_net picoexport:<name>
```
