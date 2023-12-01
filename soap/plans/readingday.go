package plans

type ReadingDay struct {
	Day       int       `json:"day" bson:"day"`
	Passages  []Passage `json:"passages,omitempty" bson:"passages,omitempty"`
	Completed bool      `json:"completed" bson:"completed"`
}

type ByReadingDay []ReadingDay

func (c ByReadingDay) Len() int { return len(c) }
func (c ByReadingDay) Less(i, j int) bool {
	return c[i].Day < c[j].Day
}
func (c ByReadingDay) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
