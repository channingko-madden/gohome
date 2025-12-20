package main

import (
	"machine"
	"math"
	"time"

	"tinygo.org/x/drivers/aht20"
	//"tinygo.org/x/drivers/hd44780"
)

func configureAHT20() *aht20.Device {
	i2c := machine.I2C0
	err := i2c.Configure(machine.I2CConfig{})
	if err != nil {
		panic("failed to configured I2C:" + err.Error())
	}

	sensor := aht20.New(i2c)
	sensor.Configure()
	return &sensor
}

// Return Temp in Fahrenheit and Relative Humitidy
func readSensor(sensor *aht20.Device) (float64, float64) {

	sensor.Read()
	fahrenheight := math.Round((float64(sensor.Celsius()) * 9 / 5) + 32)
	rh := math.Round(float64(sensor.RelHumidity())*100) / 100

	return fahrenheight, rh
}

func main() {

	sensor := configureAHT20()

	for {
		f, rh := readSensor(sensor)
		println("F: ", f, "RH: ", rh)
		time.Sleep(1 * time.Second)
	}

}
