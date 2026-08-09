# Prometheus

Launch:
```sh
docker-compose up -d
```

Hosted prometheus web page is on port 9090

Copy service discovery files like the sd_node01.yml file into the docker container using:
```sh
docker cp *.yml prometheus-prometheus-1:/prometheus
```

Remove old service discovery files in the docker container using:
```sh
docker exec prometheus-prometheus-1 rm /prometheus/<file>
```

The node exporter may not be working because of firewall issues.
