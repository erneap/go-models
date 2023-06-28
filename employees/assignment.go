package employees

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

type Assignment struct {
	ID           uint       `json:"id" bson:"id"`
	Site         string     `json:"site" bson:"site"`
	Workcenter   string     `json:"workcenter" bson:"workcenter"`
	StartDate    time.Time  `json:"startDate" bson:"startDate"`
	EndDate      time.Time  `json:"endDate" bson:"endDate"`
	Schedules    []Schedule `json:"schedules" bson:"schedules"`
	RotationDate time.Time  `json:"rotationdate" bson:"rotationdate"`
	RotationDays int        `json:"rotationdays" bson:"rotationdays"`
}

type ByAssignment []Assignment

func (c ByAssignment) Len() int { return len(c) }
func (c ByAssignment) Less(i, j int) bool {
	if c[i].StartDate.Equal(c[j].StartDate) {
		return c[i].EndDate.Before(c[j].EndDate)
	}
	return c[i].StartDate.Before(c[j].StartDate)
}
func (c ByAssignment) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (a *Assignment) UseAssignment(site string, date time.Time) bool {
	return strings.EqualFold(a.Site, site) &&
		(a.StartDate.Equal(date) || a.StartDate.Before(date)) &&
		(a.EndDate.Equal(date) || a.EndDate.After(date))
}

func (a *Assignment) GetWorkday(date time.Time, offset float64) *Workday {
	// get the site utc offset
	zoneID := "UTC"
	if offset > 0 {
		zoneID += "+" + fmt.Sprintf("%0.1f", offset)
	} else if offset < 0 {
		zoneID += fmt.Sprintf("%0.1f", offset)
	}
	timeZone := time.FixedZone(zoneID, int(offset*60*60))
	start := time.Date(a.StartDate.Year(), a.StartDate.Month(), a.StartDate.Day(),
		0, 0, 0, 0, timeZone)
	for start.Weekday() != time.Sunday {
		start = start.AddDate(0, 0, -1)
	}
	days := int(math.Floor((date.Sub(start).Hours()) / 24))
	if len(a.Schedules) == 1 || a.RotationDays <= 0 {
		iDay := days % len(a.Schedules[0].Workdays)
		return a.Schedules[0].GetWorkday(uint(iDay))
	} else if len(a.Schedules) > 1 {
		schID := (days / a.RotationDays) % len(a.Schedules)
		iDay := days % len(a.Schedules[schID].Workdays)
		return a.Schedules[schID].GetWorkday(uint(iDay))
	}
	return nil
}

func (a *Assignment) AddSchedule(days int) {
	sch := Schedule{
		ID: uint(len(a.Schedules)),
	}
	for i := 0; i < days; i++ {
		wd := Workday{
			ID:    uint(i),
			Hours: 0.0,
		}
		sch.Workdays = append(sch.Workdays, wd)
	}
	a.Schedules = append(a.Schedules, sch)
}

func (a *Assignment) ChangeScheduleDays(schedID uint, days int) {
	for i, sch := range a.Schedules {
		if sch.ID == schedID {
			sort.Sort(ByWorkday(sch.Workdays))
			if len(sch.Workdays) > days {
				sch.Workdays = sch.Workdays[:days]
			} else if len(sch.Workdays) < days {
				for j := len(sch.Workdays); j < days; j++ {
					wd := Workday{
						ID:    uint(j),
						Hours: 0.0,
					}
					sch.Workdays = append(sch.Workdays, wd)
				}
			}
			a.Schedules[i] = sch
		}
	}
}

func (a *Assignment) UpdateWorkday(schID, wdID uint, wkctr, code string,
	hours float64) error {
	for _, sch := range a.Schedules {
		if sch.ID == schID {
			sch.UpdateWorkday(wdID, wkctr, code, hours)
			return nil
		}
	}
	return errors.New("either schedule not found")
}

func (a *Assignment) RemoveSchedule(schID uint) {
	if len(a.Schedules) > 1 {
		found := false
		sort.Sort(BySchedule(a.Schedules))
		for i := 0; i < len(a.Schedules) && !found; i++ {
			if a.Schedules[i].ID == schID {
				a.Schedules = append(a.Schedules[:i], a.Schedules[i+1:]...)
			}
		}
		for i, sch := range a.Schedules {
			sch.ID = uint(i)
		}
		a.RotationDate = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
		a.RotationDays = 0
	} else {
		for _, wd := range a.Schedules[0].Workdays {
			wd.Workcenter = ""
			wd.Code = ""
			wd.Hours = 0.0
		}
	}
}

type Workday struct {
	ID         uint    `json:"id" bson:"id"`
	Workcenter string  `json:"workcenter" bson:"workcenter"`
	Code       string  `json:"code" bson:"code"`
	Hours      float64 `json:"hours" bson:"hours"`
}

type ByWorkday []Workday

func (c ByWorkday) Len() int { return len(c) }
func (c ByWorkday) Less(i, j int) bool {
	return c[i].ID < c[j].ID
}
func (c ByWorkday) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

type Schedule struct {
	ID       uint      `json:"id" bson:"id"`
	Workdays []Workday `json:"workdays" bson:"workdays"`
}

type BySchedule []Schedule

func (c BySchedule) Len() int { return len(c) }
func (c BySchedule) Less(i, j int) bool {
	return c[i].ID < c[j].ID
}
func (c BySchedule) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (sc *Schedule) GetWorkday(id uint) *Workday {
	for _, day := range sc.Workdays {
		if day.ID == id {
			return &day
		}
	}
	return nil
}

func (sc *Schedule) UpdateWorkday(id uint, wkctr, code string, hours float64) {
	found := false
	for i, day := range sc.Workdays {
		if day.ID == id {
			found = true
			day.Hours = hours
			day.Code = code
			day.Workcenter = wkctr
			sc.Workdays[i] = day
		}
	}
	if !found {
		wd := Workday{
			ID:         id,
			Workcenter: wkctr,
			Code:       code,
			Hours:      hours,
		}
		sc.Workdays = append(sc.Workdays, wd)
		sort.Sort(ByWorkday(sc.Workdays))
	}
}

func (sc *Schedule) SetScheduleDays(days int) error {
	if days == 0 || days%7 != 0 {
		return errors.New("new days value must be greater than zero and a mulitple of seven")
	}
	sort.Sort(ByWorkday(sc.Workdays))
	if days > len(sc.Workdays) {
		for i := len(sc.Workdays); i < days; i++ {
			wd := Workday{
				ID: uint(i),
			}
			sc.Workdays = append(sc.Workdays, wd)
		}
	} else if days < len(sc.Workdays) {
		sc.Workdays = sc.Workdays[:days]
	}
	return nil
}

type Variation struct {
	ID        uint      `json:"id" bson:"id"`
	Site      string    `json:"site" bson:"site"`
	IsMids    bool      `json:"mids" bson:"mids"`
	StartDate time.Time `json:"startdate" bson:"startdate"`
	EndDate   time.Time `json:"enddate" bson:"enddate"`
	Schedule  Schedule  `json:"schedule" bson:"schedule"`
}

type ByVariation []Variation

func (c ByVariation) Len() int { return len(c) }
func (c ByVariation) Less(i, j int) bool {
	if c[i].StartDate.Equal(c[j].StartDate) {
		return c[i].EndDate.Before(c[j].EndDate)
	}
	return c[i].StartDate.Before(c[j].StartDate)
}
func (c ByVariation) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (a *Variation) UseVariation(site string, date time.Time) bool {
	return strings.EqualFold(a.Site, site) &&
		(a.StartDate.Equal(date) || a.StartDate.Before(date)) &&
		(a.EndDate.Equal(date) || a.EndDate.After(date))
}

func (a *Variation) SetScheduleDays() {
	start := time.Date(a.StartDate.Year(), a.StartDate.Month(), a.StartDate.Day(),
		0, 0, 0, 0, time.UTC)
	end := time.Date(a.EndDate.Year(), a.EndDate.Month(), a.EndDate.Day(), 0, 0,
		0, 0, time.UTC)
	for start.Weekday() != time.Sunday {
		start = start.AddDate(0, 0, -1)
	}
	for end.Weekday() != time.Saturday {
		end = end.AddDate(0, 0, 1)
	}

	a.Schedule.Workdays = a.Schedule.Workdays[:0]
	count := uint(0)
	for start.Before(end) || start.Equal(end) {
		count++
		wd := Workday{
			ID: count,
		}
		a.Schedule.Workdays = append(a.Schedule.Workdays, wd)
		start = start.AddDate(0, 0, 1)
	}
}

func (a *Variation) GetWorkday(site string, date time.Time) *Workday {
	start := time.Date(a.StartDate.Year(), a.StartDate.Month(), a.StartDate.Day(),
		0, 0, 0, 0, time.UTC)
	for start.Weekday() != time.Sunday {
		start = start.AddDate(0, 0, -1)
	}
	days := int(math.Floor(date.Sub(start).Hours() / 24))
	iDay := days % len(a.Schedule.Workdays)
	return a.Schedule.GetWorkday(uint(iDay))
}

func (a *Variation) UpdateWorkday(wdID uint, wkctr, code string, hours float64) {
	a.Schedule.UpdateWorkday(wdID, wkctr, code, hours)
}
