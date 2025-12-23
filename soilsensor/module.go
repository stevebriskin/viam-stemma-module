// Implementation of the adafruit stemma soil moisture sensor. The sensor relies on the seesaw chip and library and i2c communication to read values.
// https://github.com/adafruit/Adafruit_Seesaw/blob/master/Adafruit_seesaw.cpp was used as the reference implementation.
// Product: https://www.adafruit.com/product/4026

package soilsensor

import (
	"context"
	"fmt"
	"time"

	"go.viam.com/rdk/components/board/genericlinux/buses"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
)

var (
	StemmaSoilSensor = resource.NewModel("stevebriskin", "stemma", "soil-sensor")
)

const (
	defaultI2CAddress = 0x36
)

func init() {
	resource.RegisterComponent(sensor.API, StemmaSoilSensor,
		resource.Registration[sensor.Sensor, *Config]{
			Constructor: newStemmaSoilSensorStemmaSoilSensor,
		},
	)
}

type Config struct {
	I2CBus  string `json:"i2c_bus"`
	I2cAddr int    `json:"i2c_addr,omitempty"`
}

// Validate ensures all parts of the config are valid and important fields exist.
// Returns implicit dependencies based on the config.
// The path is the JSON path in your robot's config (not the `Config` struct) to the
// resource being validated; e.g. "components.0".
func (cfg *Config) Validate(path string) ([]string, []string, error) {
	var deps []string
	if len(cfg.I2CBus) == 0 {
		return nil, nil, resource.NewConfigValidationFieldRequiredError(path, "i2c_bus")
	}
	return deps, nil, nil
}

type stemmaSoilSensorStemmaSoilSensor struct {
	resource.AlwaysRebuild
	name resource.Name

	logger logging.Logger
	cfg    *Config

	addr byte
	bus  buses.I2C
}

func newStemmaSoilSensorStemmaSoilSensor(ctx context.Context, deps resource.Dependencies, rawConf resource.Config, logger logging.Logger) (sensor.Sensor, error) {
	conf, err := resource.NativeConfig[*Config](rawConf)
	if err != nil {
		return nil, err
	}

	return NewStemmaSoilSensor(ctx, deps, rawConf.ResourceName(), conf, logger)
}

func NewStemmaSoilSensor(ctx context.Context, deps resource.Dependencies, name resource.Name, conf *Config, logger logging.Logger) (sensor.Sensor, error) {
	// Note: the compiler will flag this as a missing function on non-linux systems.
	i2cbus, err := buses.NewI2cBus(conf.I2CBus)
	if err != nil {
		return nil, fmt.Errorf("failed to open i2c bus %s: %w",
			conf.I2CBus, err)
	}

	addr := conf.I2cAddr
	if addr == 0 {
		addr = defaultI2CAddress
	}

	s := &stemmaSoilSensorStemmaSoilSensor{
		name:   name,
		logger: logger,
		cfg:    conf,
		bus:    i2cbus,
		addr:   byte(addr),
	}
	return s, nil
}

func (s *stemmaSoilSensorStemmaSoilSensor) Name() resource.Name {
	return s.name
}

func (s *stemmaSoilSensorStemmaSoilSensor) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {

	temperature, err := s.GetTemperature(ctx)
	if err != nil {
		return nil, err
	}

	moisture, err := s.GetMoisture(ctx)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{"temperature_c": temperature, "moisture": moisture}, nil
}

func (s *stemmaSoilSensorStemmaSoilSensor) GetTemperature(ctx context.Context) (float32, error) {
	handle, err := s.bus.OpenHandle(s.addr)
	if err != nil {
		return 0, err
	}
	defer handle.Close()

	err = handle.Write(ctx, []byte{0x00, 0x04})
	if err != nil {
		return 0, fmt.Errorf("failed to write to stemma while reading temperature: %w", err)
	}

	// Sleep for a moment to allow the sensor to process the command
	time.Sleep(5 * time.Millisecond)

	data, err := handle.ReadBlockData(ctx, 0x00, 4)
	if err != nil {
		return 0, fmt.Errorf("failed to read from stemma while reading temperature: %w", err)
	}

	value := float32((uint32(data[0])<<24)|(uint32(data[1])<<16)|
		(uint32(data[2])<<8)|uint32(data[3])) * float32(1.0/(1<<16))

	return value, nil
}

func (s *stemmaSoilSensorStemmaSoilSensor) GetMoisture(ctx context.Context) (int32, error) {
	handle, err := s.bus.OpenHandle(s.addr)
	if err != nil {
		return 0, err
	}
	defer handle.Close()

	err = handle.Write(ctx, []byte{0x0F, 0x10})
	if err != nil {
		return 0, fmt.Errorf("failed to write to stemma while reading temperature: %w", err)
	}

	// Sleep for a moment to allow the sensor to process the command
	time.Sleep(5 * time.Millisecond)

	data, err := handle.ReadBlockData(ctx, 0x0F, 2)
	if err != nil {
		return 0, fmt.Errorf("failed to read from stemma while reading temperature: %w", err)
	}

	value := int32((uint32(data[0]) << 8) | uint32(data[1]))

	return value, nil
}

func (s *stemmaSoilSensorStemmaSoilSensor) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return nil, nil
}

func (s *stemmaSoilSensorStemmaSoilSensor) Close(context.Context) error {
	return nil
}
