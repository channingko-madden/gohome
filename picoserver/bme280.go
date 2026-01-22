//go:build bme280

package main

import (
	"log/slog"
	"machine"
	"tinygo.org/x/drivers/bme280"
)

type BME280 struct {
	device *bme280.Device
}

func (b *BME280) ReadClimate() *climate {
	connected := b.device.Connected()
	if !connected {
		logger.Error("failed to detect BME280 with I2C")
		return nil
	}
	logger.Info("BME280 detected with I2C")

	curTemp, err := b.device.ReadTemperature()
	if err != nil {
		logger.Error("error reading BME280 temperature", slog.String("err", err.Error()))
		return nil
	}

	curRH, err := b.device.ReadHumidity()
	if err != nil {
		logger.Error("error reading BME280 relative humidity", slog.String("err", err.Error()))
		return nil
	}

	return &climate{
		TempC: float64(curTemp) / 1000,
		TempF: ((float64(curTemp) / 1000) * 9 / 5) + 32,
		RH:    float64(curRH) / 100,
	}
}

// Configure BME280
func configureSensor() *BME280 {
	i2c := machine.I2C1
	err := i2c.Configure(machine.I2CConfig{
		SCL: machine.GP19,
		SDA: machine.GP18,
	})
	if err != nil {
		panic("failed to configured I2C:" + err.Error())
	}

	sensor := bme280.New(i2c)
	sensor.Configure()

	return &BME280{
		device: &sensor}
}
