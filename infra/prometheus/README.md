# Prometheus

Copy service discovery files into the docker container using:
```sh
docker cp *.yml prometheus-prometheus-1:/prometheus
```

Remove old service discovery files in the docker container using:
```sh
docker exec prometheus-prometheus-1 rm /prometheus/<file>
```
