package main

import (
	"adafruitstemma/luxsensor"
	"adafruitstemma/soilsensor"

	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/module"
	"go.viam.com/rdk/resource"
)

func main() {
	// ModularMain can take multiple APIModel arguments, if your module implements multiple models.
	module.ModularMain(resource.APIModel{sensor.API, soilsensor.StemmaSoilSensor},
		resource.APIModel{sensor.API, luxsensor.LuxSensor},
	)
}
