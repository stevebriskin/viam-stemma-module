# Module Adafruit Stemma

Module for the the Adafruit Stemma components.

## Model stevebriskin:stemma:soil-sensor

The Stemma Soil Sensor (https://www.adafruit.com/product/4026)

### Configuration
The following attribute template can be used to configure this model:

```
{
  "i2c_bus": string,
  "i2c_addr": int
}

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
```json
{
  "temperature_c": 23.1285705,
  "moisture": 996
}
```