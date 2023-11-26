package teams

import (
	"strings"
	"time"
)

type CompanyHoliday struct {
	ID          string      `json:"id" bson:"id"`
	Name        string      `json:"name" bson:"name"`
	SortID      uint        `json:"sort" bson:"sort"`
	ActualDates []time.Time `json:"actualdates,omitempty" bson:"actualdates,omitempty"`
}

type ByCompanyHoliday []CompanyHoliday

func (c ByCompanyHoliday) Len() int { return len(c) }
func (c ByCompanyHoliday) Less(i, j int) bool {
	if c[i].ID == c[j].ID {
		return c[i].SortID < c[j].SortID
	}
	return strings.EqualFold(c[i].ID, "H")
}
func (c ByCompanyHoliday) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (ch *CompanyHoliday) GetActual(year int) *time.Time {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, actual := range ch.ActualDates {
		if (actual.Equal(start) || actual.After(start)) && actual.Before(end) {
			return &actual
		}
	}
	return nil
}

func (ch *CompanyHoliday) Purge(date time.Time) {
	for i := len(ch.ActualDates) - 1; i >= 0; i-- {
		if ch.ActualDates[i].Before(date) {
			ch.ActualDates = append(ch.ActualDates[:i], ch.ActualDates[i+1:]...)
		}
	}
}

type Company struct {
	ID             string           `json:"id" bson:"id"`
	Name           string           `json:"name" bson:"name"`
	IngestType     string           `json:"ingest" bson:"ingest"`
	IngestPeriod   int              `json:"ingestPeriod,omitempty" bson:"ingestPeriod,omitempty"`
	IngestStartDay int              `json:"startDay,omitempty" bson:"startDay,omitempty"`
	IngestPwd      string           `json:"ingestPwd" bson:"ingestPwd"`
	Holidays       []CompanyHoliday `json:"holidays,omitempty" bson:"holidays,omitempty"`
}

type ByCompany []Company

func (c ByCompany) Len() int { return len(c) }
func (c ByCompany) Less(i, j int) bool {
	return c[i].Name < c[j].Name
}
func (c ByCompany) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (c *Company) Purge(date time.Time) {
	for h, hol := range c.Holidays {
		hol.Purge(date)
		c.Holidays[h] = hol
	}
}
