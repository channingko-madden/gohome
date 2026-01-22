//go:build aht20

package main

import (
	"log/slog"
	"machine"
	"tinygo.org/x/drivers/aht20"
)

type AHT20 struct {
	device aht20.Device
}

func (a *AHT20) ReadClimate() *climate {

	status := a.device.Status()
	if status == aht20.STATUS_BUSY {
		logger.Warn("AHT20 status busy")
		return nil
	}

	err := a.device.Read()
	if err != nil {
		logger.Error("Error reading AHT20", slog.String("err", err.Error()))
		return nil
	}

	celcius := float64(a.device.Celsius())

	f := (celcius * 9 / 5) + 32
	rh := float64(a.device.RelHumidity())

	return &climate{
		TempC: celcius,
		TempF: f,
		RH:    rh,
	}

}

func configureSensor() *AHT20 {
	i2c := machine.I2C1
	err := i2c.Configure(machine.I2CConfig{
		SCL:  machine.GP19,
		SDA:  machine.GP18,
		Mode: machine.I2CModeController,
	})
	if err != nil {
		logger.Error("Failed to configure I2C")
		panic("failed to configured I2C:" + err.Error())
	}

	logger.Info("Configured AHT20 sensor I2C")

	sensor := AHT20{
		device: aht20.New(i2c),
	}

	sensor.device.Configure()
	return &sensor
}
