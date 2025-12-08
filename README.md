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