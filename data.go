package gbv

type MeterNames struct {
	Names []string
}

type Meter struct {
	Name string
	BaseUnit string
	Description string
	Measurements []Measurement
	AvailableTags []Tag
}

type Measurement struct {
	Statistic string
	Value float64
}

type Tag struct {
	Tag string
	Values []string
}

