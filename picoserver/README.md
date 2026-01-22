# picoserver

![Pico Board](pico_board.jpg)

## Build flags

Use a build flag to specify the type of sensor
- BME280 `-tags=bme280`
- AHT20 `-tags=aht20`

## tinygo build
```
tinygo build -target=pico -stack-size=8kb -size=short -tags={} -o main.uf2 .
```

Copy the uf2 to the Pico W.

OR

## tinygo flash
```
tinygo flash -target=pico -stack-size=8kb -size=short -tags={} .
```

See logging info using `tinygo monitor` or add -monitor when building/flashing to automatically
launch monitoring.
