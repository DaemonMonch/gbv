package main

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

func (m *Meter) flattenTags() []string {
	arr := make([]string,len(m.AvailableTags))
	for _, tag := range m.AvailableTags {
		arr = append(arr,tag.flatten()...)
	}
	return arr
}

type Measurement struct {
	Statistic string
	Value float64
}

type Tag struct {
	Tag string
	Values []string
}

func (t Tag) flatten() []string {
	arr := make([]string,len(t.Values))
	for i := 0; i < len(t.Values); i++ {
		arr[i] = t.Tag  + ":"+ t.Values[i]
	}
	return arr
}

