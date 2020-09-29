# homekit

A go library and cli for interacting with HomeKit accessories. It can be used to read data from and control accessories.

For building HomeKit accessories with go see [brutella/hc](https://github.com/brutella/hc).

## Status

This project is currently in a prototype state. Many features are unimplemented (events, bluetooth, tests, etc) and implemented features may not work with all accessories. It does work with the devices I'm able to test with so it may work for you!

The API should be considered unstable and will probably change with future updates.

## CLI Usage

### Installing

```
$ go get github.com/mctofu/homekit/cmd/homekit
```

### Initialize
Create a new controller config file at `$USER_CONFIG_DIR/mctofu/homekit/default.json` and initialize it with random identity keys.
```
$ homekit createController
```

### Find a device to pair with

```
$ homekit discover
Detected Device: VELUX\ gateway
Model: VELUX
ID: AA:BB:CC:DD:EE:FF
IPs: [192.168.1.101]
Port: 5001
Feature flags: Supports HAP Pairing (1)
Status flags: Accessory is paired (0)

Found 1 devices
```

### Pair with the device
```
$ homekit pair --id AA:BB:CC:DD:EE:FF --pin XXX-XX-XXX --name alias
Accessory paired successfully!
```

### List attributes
```
$ homekit listCharacteristics --name alias
Accessory: 1 VELUX gateway (g123ab4)
  Service: 1 AccessoryInformation (3E)
    1.2: VELUX gateway / Name (23) [pr]
    1.3: VELUX / Manufacturer (20) [pr]
    1.4: VELUX Gateway / Model (21) [pr]
    1.5: g123ab4 / SerialNumber (30) [pr]
    1.6: false / Identify (14) [pw]
    1.7: 68 / FirmwareRevision (52) [pr]
  Service: 8 HAPProtocolInformation (A2)
    1.9: 1.1.0 / Version (37) [pr]
Accessory: 2 VELUX Sensor (p123456)
  Service: 1 AccessoryInformation (3E)
    2.2: VELUX Sensor / Name (23) [pr]
    2.3: VELUX / Manufacturer (20) [pr]
    2.4: VELUX Sensor / Model (21) [pr]
    2.5: p123456 / SerialNumber (30) [pr]
    2.7: false / Identify (14) [pw]
    2.6: 16 / FirmwareRevision (52) [pr]
  Service: 8 TemperatureSensor (8A)
    2.9: Temperature sensor / Name (23) [pr]
    2.10: 28.5 / CurrentTemperature (11) [pr ev]
  Service: 11 HumiditySensor (82)
    2.12: Humidity sensor / Name (23) [pr]
    2.13: 54 / CurrentRelativeHumidity (10) [pr ev]
  Service: 14 CarbonDioxideSensor (97)
    2.15: Carbon Dioxide sensor / Name (23) [pr]
    2.16: 0 / CarbonDioxideDetected (92) [pr ev]
    2.17: 1436 / CarbonDioxideLevel (93) [pr ev]
Accessory: 3 VELUX Window (123a4567890123b4)
  Service: 1 AccessoryInformation (3E)
    3.2: VELUX Window / Name (23) [pr]
    3.3: VELUX / Manufacturer (20) [pr]
    3.4: VELUX Window / Model (21) [pr]
    3.5: 123a4567890123b4 / SerialNumber (30) [pr]
    3.7: false / Identify (14) [pw]
    3.6: 12 / FirmwareRevision (52) [pr]
  Service: 8 Window (8B)
    3.9: Roof Window / Name (23) [pr]
    3.11: 0 / TargetPosition (7C) [pr pw ev]
    3.10: 0 / CurrentPosition (6D) [pr ev]
    3.12: 2 / PositionState (72) [pr ev]

... trimmed ...

Accessory: 10 VELUX Window (123a4567890123b5)
  Service: 1 AccessoryInformation (3E)
    10.2: VELUX Window / Name (23) [pr]
    10.3: VELUX / Manufacturer (20) [pr]
    10.4: VELUX Window / Model (21) [pr]
    10.5: 123a4567890123b5 / SerialNumber (30) [pr]
    10.7: false / Identify (14) [pw]
    10.6: 12 / FirmwareRevision (52) [pr]
  Service: 8 Window (8B)
    10.9: Roof Window / Name (23) [pr]
    10.11: 0 / TargetPosition (7C) [pr pw ev]
    10.10: 0 / CurrentPosition (6D) [pr ev]
    10.12: 2 / PositionState (72) [pr ev]
```

### Get particular characteristics
```
$ homekit getCharacteristics --name alias -c 2.10 -c 2.17
2.10: CurrentTemperature
Value: 28.5
2.17: CarbonDioxideLevel
Value: 1421
```

### Set characteristics
```
$ homekit setCharacteristics --name alias -c 3.11=60 -c 10.11=50
```

## Acknowlegments

- [brutella/hc](https://github.com/brutella/hc) provides much of the pairing and secure connection negotiation functionality.
- [jlusiardi/homekit_python](https://github.com/jlusiardi/homekit_python) was a great reference for building a HomeKit controller.
