// Implementation of the Adafruit VEML7700 Light/Lux Sensor.
// Product:https://www.adafruit.com/product/4162, https://www.vishay.com/docs/84286/veml7700.pdf https://www.vishay.com/docs/84323/designingveml7700.pdf
// https://github.com/adafruit/Adafruit_VEML7700/blob/master/Adafruit_VEML7700.cpp wsa used a reference implementation.

package luxsensor

import (
	"context"
	"fmt"
	"math"

	"go.viam.com/rdk/components/board/genericlinux/buses"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
)

var (
	LuxSensor = resource.NewModel("stevebriskin", "stemma", "lux-sensor")
)

const (
	defaultI2CAddress = 0x10
)

const (
	ALS_CONFIG        = 0x00 // Light configuration register
	ALS_THREHOLD_HIGH = 0x01 // Light high threshold for irq
	ALS_THREHOLD_LOW  = 0x02 // Light low threshold for irq
	ALS_POWER_SAVE    = 0x03 // Power save register
	ALS_DATA          = 0x04 // The light data output
	WHITE_DATA        = 0x05 // The white light data output
	INTERRUPTSTATUS   = 0x06 // What IRQ (if any)

	INTERRUPT_HIGH = 0x4000 // Interrupt status for high threshold
	INTERRUPT_LOW  = 0x8000 // Interrupt status for low threshold

	POWERSAVE_MODE1 = 0x00 // Power saving mode 1
	POWERSAVE_MODE2 = 0x01 // Power saving mode 2
	POWERSAVE_MODE3 = 0x02 // Power saving mode 3
	POWERSAVE_MODE4 = 0x03 // Power saving mode 4

	MAX_RES  float64 = 0.0042
	GAIN_MAX         = 2.0
	IT_MAX           = 800.0
)

type Persistence uint8

const (
	PERS_1 Persistence = 0x00 // ALS irq persistence 1 sample
	PERS_2 Persistence = 0x01 // ALS irq persistence 2 samples
	PERS_4 Persistence = 0x02 // ALS irq persistence 4 samples
	PERS_8 Persistence = 0x03 // ALS irq persistence 8 samples
)

func PersistenceFromValue(v int) Persistence {
	switch v {
	case 1:
		return PERS_1
	case 2:
		return PERS_2
	case 4:
		return PERS_4
	case 8:
		return PERS_8
	default:
		return 0xFF // invalid, no corresponding Persistence
	}
}

// Gain settings for VEML7700
type Gain uint8

const (
	GAIN_1   Gain = 0x00
	GAIN_2   Gain = 0x01
	GAIN_1_8 Gain = 0x02
	GAIN_1_4 Gain = 0x03
)

func (g Gain) Factor() float64 {
	switch g {
	case GAIN_1:
		return 1.0
	case GAIN_2:
		return 2.0
	case GAIN_1_4:
		return 0.25
	case GAIN_1_8:
		return 0.125
	default:
		return -1
	}
}

// Reverse function: returns Gain for the given factor, or 0xFF for error
func GainFromFactor(f string) Gain {
	switch f {
	case "1":
		return GAIN_1
	case "2":
		return GAIN_2
	case "1/4":
		return GAIN_1_4
	case "1/8":
		return GAIN_1_8
	default:
		return 0xFF // invalid, no corresponding Gain
	}
}

// Integration time settings for VEML7700
type IntegrationTime uint8

const (
	IT_100MS IntegrationTime = 0x00
	IT_200MS IntegrationTime = 0x01
	IT_400MS IntegrationTime = 0x02
	IT_800MS IntegrationTime = 0x03
	IT_50MS  IntegrationTime = 0x08
	IT_25MS  IntegrationTime = 0x0C
)

func (it IntegrationTime) TimeMs() int {
	switch it {
	case IT_100MS:
		return 100
	case IT_200MS:
		return 200
	case IT_400MS:
		return 400
	case IT_800MS:
		return 800
	case IT_50MS:
		return 50
	case IT_25MS:
		return 25
	default:
		return -1
	}
}

// Reverse function: returns IntegrationTime for the given ms value, or -1 for error
func IntegrationTimeFromMs(ms int) IntegrationTime {
	switch ms {
	case 100:
		return IT_100MS
	case 200:
		return IT_200MS
	case 400:
		return IT_400MS
	case 800:
		return IT_800MS
	case 50:
		return IT_50MS
	case 25:
		return IT_25MS
	default:
		return 0xFF // invalid, no corresponding IntegrationTime
	}
}

func init() {
	resource.RegisterComponent(sensor.API, LuxSensor,
		resource.Registration[sensor.Sensor, *Config]{
			Constructor: newLuxSensor,
		},
	)
}

type Config struct {
	I2CBus          string `json:"i2c_bus"`
	I2cAddr         int    `json:"i2c_addr,omitempty"`
	Gain            string `json:"gain,omitempty"`
	IntegrationTime int    `json:"integration_time,omitempty"`
	Persistence     int    `json:"persistence,omitempty"`
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

type luxSensor struct {
	resource.AlwaysRebuild
	name resource.Name

	logger logging.Logger
	cfg    *Config

	addr byte
	bus  buses.I2C

	gain        Gain
	integration IntegrationTime
	persistence Persistence
}

func newLuxSensor(ctx context.Context, deps resource.Dependencies, rawConf resource.Config, logger logging.Logger) (sensor.Sensor, error) {
	conf, err := resource.NativeConfig[*Config](rawConf)
	if err != nil {
		return nil, err
	}

	return NewLuxSensor(ctx, deps, rawConf.ResourceName(), conf, logger)
}

func NewLuxSensor(ctx context.Context, deps resource.Dependencies, name resource.Name, conf *Config, logger logging.Logger) (sensor.Sensor, error) {
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

	gain := GAIN_1_4
	if conf.Gain != "" {
		gain := GainFromFactor(conf.Gain)
		if gain == 0xFF {
			return nil, fmt.Errorf("Invalid gain value: %s. Valid values are 1, 2, 1/4, 1/8", conf.Gain)
		}
	}

	integration := IT_100MS
	if conf.IntegrationTime != 0 {
		integration := IntegrationTimeFromMs(conf.IntegrationTime)
		if integration == 0xFF {
			return nil, fmt.Errorf("Invalid integration time value: %d. Valid values are 100, 200, 400, 800, 50, 25", conf.IntegrationTime)
		}
	}

	persistence := PERS_1
	if conf.Persistence != 0 {
		persistence := PersistenceFromValue(conf.Persistence)
		if persistence == 0xFF {
			return nil, fmt.Errorf("Invalid persistence value: %d. Valid values are 1, 2, 4, 8", conf.Persistence)
		}
	}

	s := &luxSensor{
		name:        name,
		logger:      logger,
		cfg:         conf,
		bus:         i2cbus,
		addr:        byte(addr),
		gain:        gain,
		integration: integration,
		persistence: persistence,
	}

	if err := s.Initialize(ctx); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *luxSensor) Name() resource.Name {
	return s.name
}

func (s *luxSensor) Initialize(ctx context.Context) error {
	handle, err := s.bus.OpenHandle(s.addr)
	if err != nil {
		return fmt.Errorf("failed to open I2C handle at address 0x%X: %w", s.addr, err)
	}
	defer handle.Close()

	// Write configuration to ALS_CONF register (0x00)
	conf := packConfig(true, false, s.persistence, s.integration, s.gain)
	s.logger.Infof("Writing config to VEML7700: %016b", conf)
	if err := handle.WriteBlockData(ctx, ALS_CONFIG, []byte{byte(conf >> 8), byte(conf)}); err != nil {
		return fmt.Errorf("failed to write config to VEML7700: %w", err)
	}

	if err := handle.WriteBlockData(ctx, ALS_POWER_SAVE, []byte{0x00, 0x00}); err != nil {
		return fmt.Errorf("failed to write power save to VEML7700: %w", err)
	}
	return nil
}

// packConfig constructs a 16-bit VEML7700 config value according to register bit order.
func packConfig(enable bool, interrupt bool, persistence Persistence, integration IntegrationTime, gain Gain) uint16 {
	var conf uint16 = 0

	// bits 15-13: 000, already zero
	// bits 12-11: gain (2 bits)
	conf |= (uint16(gain&0x3) << 11)
	// bit 10: 0, already zero

	// bits 9-6: integration (4 bits)
	conf |= (uint16(integration&0xF) << 6)

	// bits 5-4: persistence (2 bits)
	conf |= (uint16(persistence&0x3) << 4)

	// bits 3-2: 00, already zero

	// bit 1: interrupt
	if interrupt {
		conf |= (1 << 1)
	}
	// bit 0: enable
	if enable {
		conf |= 1
	}

	return conf
}

func (s *luxSensor) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {

	als, err := s.readALS(ctx)
	if err != nil {
		return nil, err
	}

	lux := s.computeLux(ctx, als)

	// TODO: get other values and convert als to lux
	return map[string]interface{}{"als": als, "lux": lux}, nil

}

// Reference https://www.vishay.com/docs/84323/designingveml7700.pdf section "CALCULATING THE LUX LEVEL"
func (s *luxSensor) computeLux(ctx context.Context, als uint16) float64 {
	resolution := s.getResolution()
	lux := resolution * float64(als)

	// From the spec:
	// For illuminations > 1000 lx a correction formula needs to be applied.
	// When using GAIN level 1/4 and 1/8 the correction formula should be used.

	if s.gain == GAIN_1_4 || s.gain == GAIN_1_8 || lux > 1000 {
		// Apply correction formula for high lux values or low gain per VEML7700 datasheet
		// a*(lux^4) + b*(lux^3) + c*(lux^2) + d*lux
		// a = 6.0135e-13, b = -9.3924e-9, c = 8.1488e-5, d = 1.0023
		correctedLux := 6.0135e-13*math.Pow(lux, 4) +
			-9.3924e-9*math.Pow(lux, 3) +
			8.1488e-5*math.Pow(lux, 2) +
			1.0023*lux

		return correctedLux
	}

	return lux
}

func (s *luxSensor) getResolution() float64 {
	it_time := s.integration.TimeMs()
	gain_factor := s.gain.Factor()

	return MAX_RES * (IT_MAX / float64(it_time)) *
		(GAIN_MAX / gain_factor)
}

func (s *luxSensor) readALS(ctx context.Context) (uint16, error) {
	handle, err := s.bus.OpenHandle(s.addr)
	if err != nil {
		return 0, err
	}
	defer handle.Close()

	data, err := handle.ReadBlockData(ctx, ALS_DATA, 2)
	if err != nil {
		return 0, fmt.Errorf("failed to read from stemma while reading ALS data: %w", err)
	}

	// VEML7700 returns data in little-endian (LSB first) order.
	s.logger.Infof("Read ALS data from VEML7700. MSB: %02x, LSB: %02x, value: %04x", data[1], data[0], uint16(data[1])<<8|uint16(data[0]))
	return uint16(data[1])<<8 | uint16(data[0]), nil
}

func (s *luxSensor) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return nil, nil
}

func (s *luxSensor) Close(context.Context) error {
	return nil
}
