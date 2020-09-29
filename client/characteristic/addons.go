package characteristic

// Celcius returns the temperature in degrees Celcius
func (c *CurrentTemperature) Celcius() float64 {
	return c.Value
}

// Fahrenheit returns the temperature in degrees Fahrenheit
func (c *CurrentTemperature) Fahrenheit() float64 {
	return c.Value*9/5 + 32
}
