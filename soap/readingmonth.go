package soap

type ReadingMonth struct {
	Month int          `json:"month" bson:"month"`
	Days  []ReadingDay `json:"days,omitempty" bson:"days,omitempty"`
}

type ByReadingMonth []ReadingMonth

func (c ByReadingMonth) Len() int { return len(c) }
func (c ByReadingMonth) Less(i, j int) bool {
	return c[i].Month < c[j].Month
}
func (c ByReadingMonth) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (m *ReadingMonth) IsCompleted() bool {
	answer := true
	for _, day := range m.Days {
		if !day.Completed {
			answer = false
		}
	}
	return answer
}
