package service

import "github.com/mctofu/homekit/client/characteristic"

// TODO: generate these

type AccessoryInfo struct {
	FirmwareRevision characteristic.FirmwareRevision
	Manufacturer     characteristic.Manufacturer
	Model            characteristic.Model
	Name             characteristic.Name
	SerialNumber     characteristic.SerialNumber
	HardwareRevision *characteristic.HardwareRevision
	AccessoryFlags   *characteristic.AccessoryFlags
}

func ReadAccessoryInfo(chs []*characteristic.RawCharacteristic) *AccessoryInfo {
	var info AccessoryInfo
	for _, c := range chs {
		switch c.Type {
		case characteristic.TypeName:
			info.Name = *characteristic.ReadName(c)
		case characteristic.TypeFirmwareRevision:
			info.FirmwareRevision = *characteristic.ReadFirmwareRevision(c)
		case characteristic.TypeAccessoryFlags:
			info.AccessoryFlags = characteristic.ReadAccessoryFlags(c)
		case characteristic.TypeManufacturer:
			info.Manufacturer = *characteristic.ReadManufacturer(c)
		case characteristic.TypeModel:
			info.Model = *characteristic.ReadModel(c)
		case characteristic.TypeSerialNumber:
			info.SerialNumber = *characteristic.ReadSerialNumber(c)
		case characteristic.TypeHardwareRevision:
			info.HardwareRevision = characteristic.ReadHardwareRevision(c)
		}
	}

	return &info
}

type TemperatureSensor struct {
	Name               *characteristic.Name
	CurrentTemperature characteristic.CurrentTemperature
	StatusActive       *characteristic.StatusActive
	StatusFault        *characteristic.StatusFault
	StatusLowBattery   *characteristic.StatusLowBattery
	StatusTampered     *characteristic.StatusTampered
}

func ReadTemperatureSensor(chs []*characteristic.RawCharacteristic) *TemperatureSensor {
	var ts TemperatureSensor
	for _, ch := range chs {
		switch ch.Type {
		case characteristic.TypeCurrentTemperature:
			ts.CurrentTemperature = *characteristic.ReadCurrentTemperature(ch)
		case characteristic.TypeName:
			ts.Name = characteristic.ReadName(ch)
		case characteristic.TypeStatusActive:
			ts.StatusActive = characteristic.ReadStatusActive(ch)
		case characteristic.TypeStatusTampered:
			ts.StatusTampered = characteristic.ReadStatusTampered(ch)
		case characteristic.TypeStatusLowBattery:
			ts.StatusLowBattery = characteristic.ReadStatusLowBattery(ch)
		case characteristic.TypeStatusFault:
			ts.StatusFault = characteristic.ReadStatusFault(ch)
		}
	}
	return &ts
}

type HumiditySensor struct {
	Name                    *characteristic.Name
	CurrentRelativeHumidity characteristic.CurrentRelativeHumidity
	StatusActive            *characteristic.StatusActive
	StatusFault             *characteristic.StatusFault
	StatusLowBattery        *characteristic.StatusLowBattery
	StatusTampered          *characteristic.StatusTampered
}

func ReadHumiditySensor(chs []*characteristic.RawCharacteristic) *HumiditySensor {
	var hs HumiditySensor
	for _, ch := range chs {
		switch ch.Type {
		case characteristic.TypeCurrentRelativeHumidity:
			hs.CurrentRelativeHumidity = *characteristic.ReadCurrentRelativeHumidity(ch)
		case characteristic.TypeName:
			hs.Name = characteristic.ReadName(ch)
		case characteristic.TypeStatusActive:
			hs.StatusActive = characteristic.ReadStatusActive(ch)
		case characteristic.TypeStatusTampered:
			hs.StatusTampered = characteristic.ReadStatusTampered(ch)
		case characteristic.TypeStatusLowBattery:
			hs.StatusLowBattery = characteristic.ReadStatusLowBattery(ch)
		case characteristic.TypeStatusFault:
			hs.StatusFault = characteristic.ReadStatusFault(ch)
		}
	}
	return &hs
}

type CarbonDioxideSensor struct {
	Name                   *characteristic.Name
	CarbonDioxideDetected  characteristic.CarbonDioxideDetected
	CarbonDioxideLevel     *characteristic.CarbonDioxideLevel
	CarbonDioxidePeakLevel *characteristic.CarbonDioxidePeakLevel
	StatusActive           *characteristic.StatusActive
	StatusFault            *characteristic.StatusFault
	StatusLowBattery       *characteristic.StatusLowBattery
	StatusTampered         *characteristic.StatusTampered
}

func ReadCarbonDioxideSensor(chs []*characteristic.RawCharacteristic) *CarbonDioxideSensor {
	var cs CarbonDioxideSensor
	for _, ch := range chs {
		switch ch.Type {
		case characteristic.TypeCarbonDioxideDetected:
			cs.CarbonDioxideDetected = *characteristic.ReadCarbonDioxideDetected(ch)
		case characteristic.TypeCarbonDioxideLevel:
			cs.CarbonDioxideLevel = characteristic.ReadCarbonDioxideLevel(ch)
		case characteristic.TypeCarbonDioxidePeakLevel:
			cs.CarbonDioxidePeakLevel = characteristic.ReadCarbonDioxidePeakLevel(ch)
		case characteristic.TypeName:
			cs.Name = characteristic.ReadName(ch)
		case characteristic.TypeStatusActive:
			cs.StatusActive = characteristic.ReadStatusActive(ch)
		case characteristic.TypeStatusTampered:
			cs.StatusTampered = characteristic.ReadStatusTampered(ch)
		case characteristic.TypeStatusLowBattery:
			cs.StatusLowBattery = characteristic.ReadStatusLowBattery(ch)
		case characteristic.TypeStatusFault:
			cs.StatusFault = characteristic.ReadStatusFault(ch)
		}
	}
	return &cs
}

type Window struct {
	Name                *characteristic.Name
	CurrentPosition     characteristic.CurrentPosition
	TargetPosition      characteristic.TargetPosition
	PositionState       characteristic.PositionState
	ObstructionDetected *characteristic.ObstructionDetected
}

func ReadWindow(chs []*characteristic.RawCharacteristic) *Window {
	var w Window
	for _, ch := range chs {
		switch ch.Type {
		case characteristic.TypeCurrentPosition:
			w.CurrentPosition = *characteristic.ReadCurrentPosition(ch)
		case characteristic.TypeTargetPosition:
			w.TargetPosition = *characteristic.ReadTargetPosition(ch)
		case characteristic.TypePositionState:
			w.PositionState = *characteristic.ReadPositionState(ch)
		case characteristic.TypeName:
			w.Name = characteristic.ReadName(ch)
		case characteristic.TypeObstructionDetected:
			w.ObstructionDetected = characteristic.ReadObstructionDetected(ch)
		}
	}
	return &w
}

type WindowCovering struct {
	Name                *characteristic.Name
	CurrentPosition     characteristic.CurrentPosition
	TargetPosition      characteristic.TargetPosition
	PositionState       characteristic.PositionState
	ObstructionDetected *characteristic.ObstructionDetected
}

func ReadWindowCovering(chs []*characteristic.RawCharacteristic) *WindowCovering {
	var w WindowCovering
	for _, ch := range chs {
		switch ch.Type {
		case characteristic.TypeCurrentPosition:
			w.CurrentPosition = *characteristic.ReadCurrentPosition(ch)
		case characteristic.TypeTargetPosition:
			w.TargetPosition = *characteristic.ReadTargetPosition(ch)
		case characteristic.TypePositionState:
			w.PositionState = *characteristic.ReadPositionState(ch)
		case characteristic.TypeName:
			w.Name = characteristic.ReadName(ch)
		case characteristic.TypeObstructionDetected:
			w.ObstructionDetected = characteristic.ReadObstructionDetected(ch)
		}
	}
	return &w
}
