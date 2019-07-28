package airConditioner

import (
	"fmt"
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
	Manufacturer = "xiggua"
)

type airConditioner struct {
	Accessory *accessory.Accessory

	Thermost           *service.Thermostat
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

	// init thermost and fanv2
	acc.Thermost = service.NewThermostat()
	fanv2 := service.NewFanV2()

	acc.ACRemoteController.temperatureCallBack = func(temperature float64) {
		acc.Thermost.CurrentTemperature.SetValue(temperature)
	}
	acc.ACRemoteController.modeCallback = func(mode int) {
		acc.Thermost.TargetHeatingCoolingState.SetValue(mode)
	}

	// config thermost
	acc.Thermost.CurrentTemperature.SetValue(DefaultTemperature)
	acc.Thermost.TargetTemperature.SetMinValue(MinTemperature)
	acc.Thermost.TargetTemperature.SetMaxValue(MaxTemperature)
	acc.Thermost.TargetTemperature.SetValue(DefaultTemperature)
	acc.Thermost.TargetTemperature.SetStepValue(1)
	acc.Thermost.TargetHeatingCoolingState.SetValue(COOL)

	acc.Thermost.TargetTemperature.OnValueRemoteUpdate(
		acc.ACRemoteController.updateTargetTemperature)
	acc.Thermost.TargetHeatingCoolingState.OnValueRemoteUpdate(
		acc.ACRemoteController.updateTargetHeatingCoolingState)

	// config fanv2
	speed := characteristic.NewRotationSpeed()
	speed.SetMinValue(MinSpeed)
	speed.SetMaxValue(MaxSpeed)
	speed.SetStepValue(1)
	speed.SetValue(DefaultSpeed)
	fanv2.AddCharacteristic(speed.Characteristic)
	speed.OnValueRemoteUpdate(
		acc.ACRemoteController.updateFanSpeed)
	acc.Accessory.AddService(acc.Thermost.Service)
	acc.Thermost.AddLinkedService(fanv2.Service)
	acc.Accessory.AddService(fanv2.Service)
	return acc
}

func AirConditionerMain(storage string, pin string) {
	acc := initAirConditioner()
	fmt.Println("successfully init")
	t, err := hc.NewIPTransport(hc.Config{Pin: pin, StoragePath: storage}, acc.Accessory)
	if err != nil {
		log.Fatal(err)
	}

	hc.OnTermination(func() {
		<-t.Stop()
	})

	t.Start()
}
