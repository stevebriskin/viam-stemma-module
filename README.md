# Module Adafruit Stemma

Module for the the Adafruit Stemma components.

## Model stevebriskin:stemma:soil-sensor

The Stemma Soil Sensor: Adafruit STEMMA Soil Sensor - I2C Capacitive Moisture Sensor (https://www.adafruit.com/product/4026). From documentation: "[the sensor] will give you a reading ranging from about 200 (very dry) to 2000 (very wet). As a bonus, we also give you the ambient temperature from the internal temperature sensor on the microcontroller, it's not high precision, maybe good to + or - 2 degrees Celsius."

### Configuration
The following attribute template can be used to configure this model:

```json
{
  "i2c_bus": "string",
  "i2c_addr": "int"
}
```

#### Attributes

The following attributes are available for this model:

| Name          | Type   | Inclusion | Description                |
|---------------|--------|-----------|----------------------------|
| `i2c_bus` | string  | Required  | I2C bus to use |
| `i2c_address` | string | Optional  | I2C address of the soil sensor. Default: 0x36 |

#### Example Configuration

```json
{
  "i2c_bus": "1"
}
```

### Data

The `Readings` method returns the temperature (in Celcius) and capacitative moisture values. Example:

| Field           | Type    | Description                                     |
|-----------------|---------|-------------------------------------------------|
| `temperature_c` | float   | Temperature in Celsius, +/- 2c accuracy.        |
| `moisture`      | int     | Moisture value (200 is dry, 2000 is very wet)   |

Example:
```json
{
  "temperature_c": 23.1285705,
  "moisture": 996
}
```

## Model stevebriskin:stemma:lux-sensor

Implementation of the Adafruit Vishay VEML7700 Light/Lux Sensor.
Product:https://www.adafruit.com/product/4162, https://www.vishay.com/docs/84286/veml7700.pdf https://www.vishay.com/docs/84323/designingveml7700.pdf

Per spec, a correction factor is automatically applied for lux readings over 1000lx or if gain is either 1/4 or 1/8.

### Configuration
The following attribute template can be used to configure this model:

```json
{
  "i2c_bus": "string",
  "gain": "string",
  "integration_time": int,
  "persistence": int  
}
```

#### Attributes

The following attributes are available for this model:
See Vishay model specs for configuration details as they control reading ranges.

| Name               | Type    | Required | Description                                                                             |
|--------------------|---------|----------|-----------------------------------------------------------------------------------------|
| `i2c_bus`          | string  | Yes      | I2C bus to use.                                                                         |
| `gain`             | string  | No       | Sensor gain. Allowed: "1", "2", "1/4", "1/8". Default: "1/8".                           |
| `integration_time` | int     | No       | Integration time in milliseconds. Allowed: 100, 200, 400, 800. Default: 100.            |
| `persistence`      | int     | No       | Persistence setting. Allowed: 1, 2, 4, 8. Default: 1.                                   |



#### Example Configuration

```json
{
  "i2c_bus": "1",
  "gain": "1/4"
}
```

### Data

The `Readings` method returns the raw als readings and computed lux.

| Field           | Type    | Description                                     |
|-----------------|---------|-------------------------------------------------|
| `als` | int   | Raw ALS reading from the sensor. |
| `lux`      | float     | Computed lux value, with a correction factor applied if needed   |
