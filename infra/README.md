# Infrastructure

## Grafana

```
docker run -d --name=grafana01 --restart=always --net=prometheus_prom_net -p 3000:3000 grafana/grafana
```
