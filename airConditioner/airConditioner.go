package airConditioner

import (
	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/service"
	"log"
)

const (
	Name         = "Air Conditioner"
	SerialNumber = "8848"
	Model        = "钛合金限量版"
	Manufacturer = "Xiggua Technology"
)

type airConditioner struct {
	Accessory *accessory.Accessory

	Thermostat         *service.Thermostat
	ACRemoteController *ACRemoteController
}

func initAirConditioner() airConditioner {
	info := accessory.Info{
		Name:         Name,
		SerialNumber: SerialNumber,
		Manufacturer: Manufacturer,
		Model:        Model,
	}

	acc := airConditioner{}
	acc.ACRemoteController = initDefaultACRemoteController()
	acc.Accessory = accessory.New(info, accessory.TypeAirConditioner)

	// init thermostat and fanV2
	acc.Thermostat = service.NewThermostat()
	fanV2 := service.NewFanV2()

	acc.ACRemoteController.temperatureCallBack = func(temperature float64) {
		acc.Thermostat.CurrentTemperature.SetValue(temperature)
	}
	acc.ACRemoteController.modeCallback = func(mode int) {
		acc.Thermostat.TargetHeatingCoolingState.SetValue(mode)
	}

	// config thermostat
	acc.Thermostat.CurrentTemperature.SetValue(DefaultTemperature)
	acc.Thermostat.TargetTemperature.SetMinValue(MinTemperature)
	acc.Thermostat.TargetTemperature.SetMaxValue(MaxTemperature)
	acc.Thermostat.TargetTemperature.SetValue(DefaultTemperature)
	acc.Thermostat.TargetTemperature.SetStepValue(1)
	acc.Thermostat.TargetHeatingCoolingState.SetValue(OFF)
	acc.Thermostat.CurrentHeatingCoolingState.SetValue(OFF)

	acc.Thermostat.TargetTemperature.OnValueRemoteUpdate(
		acc.ACRemoteController.updateTargetTemperature)
	acc.Thermostat.TargetHeatingCoolingState.OnValueRemoteUpdate(
		acc.ACRemoteController.updateTargetHeatingCoolingState)

	// config fanV2
	speed := characteristic.NewRotationSpeed()
	speed.SetMinValue(MinSpeed)
	speed.SetMaxValue(MaxSpeed)
	speed.SetStepValue(1)
	speed.SetValue(DefaultSpeed)
	fanV2.AddCharacteristic(speed.Characteristic)
	speed.OnValueRemoteUpdate(
		acc.ACRemoteController.updateFanSpeed)
	acc.Accessory.AddService(acc.Thermostat.Service)
	acc.Thermostat.AddLinkedService(fanV2.Service)
	acc.Accessory.AddService(fanV2.Service)
	return acc
}

func StartAirConditioner(storage string, pin string) {
	acc := initAirConditioner()
	log.Println("successfully init")
	t, err := hc.NewIPTransport(hc.Config{Pin: pin, StoragePath: storage}, acc.Accessory)
	if err != nil {
		log.Fatal(err)
	}

	hc.OnTermination(func() {
		<-t.Stop()
	})

	t.Start()
}
