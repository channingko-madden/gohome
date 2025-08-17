# gohome

## Docker

### Build
```
docker build -t restapi:v1 .
```

### Run
```
docker run -d -p 4000:4000 --name restapi-v1 --restart=always restapi:v1
```
