package employees

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/erneap/go-models/users"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Employee struct {
	ID          primitive.ObjectID  `json:"id" bson:"_id"`
	TeamID      primitive.ObjectID  `json:"team" bson:"team"`
	SiteID      string              `json:"site" bson:"site"`
	UserID      primitive.ObjectID  `json:"userid" bson:"userid"`
	Email       string              `json:"email" bson:"email"`
	Name        EmployeeName        `json:"name" bson:"name"`
	Data        *EmployeeData       `json:"data,omitempty" bson:"data,omitempty"`
	CompanyInfo CompanyInfo         `json:"companyinfo"`
	Assignments []Assignment        `json:"assignments,omitempty"`
	Variations  []Variation         `json:"variations,omitempty"`
	Balances    []AnnualLeave       `json:"balance,omitempty"`
	Leaves      []LeaveDay          `json:"leaves,omitempty"`
	Requests    []LeaveRequest      `json:"requests,omitempty"`
	LaborCodes  []EmployeeLaborCode `json:"laborCodes,omitempty"`
	User        *users.User         `json:"user,omitempty" bson:"-"`
	Work        []Work              `json:"work,omitempty" bson:"-"`
}

type ByEmployees []Employee

func (c ByEmployees) Len() int { return len(c) }
func (c ByEmployees) Less(i, j int) bool {
	if c[i].Name.LastName == c[j].Name.LastName {
		if c[i].Name.FirstName == c[j].Name.FirstName {
			return c[i].Name.MiddleName < c[j].Name.MiddleName
		}
		return c[i].Name.FirstName < c[j].Name.FirstName
	}
	return c[i].Name.LastName < c[j].Name.LastName
}
func (c ByEmployees) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (e *Employee) RemoveLeaves(start, end time.Time) {
	if e.Data != nil {
		e.ConvertFromData()
	}
	sort.Sort(ByLeaveDay(e.Leaves))
	startpos := -1
	endpos := -1
	for i, lv := range e.Leaves {
		if startpos < 0 && (lv.LeaveDate.Equal(start) || lv.LeaveDate.After(start)) &&
			(lv.LeaveDate.Equal(end) || lv.LeaveDate.Before(end)) {
			startpos = i
		} else if startpos >= 0 && (lv.LeaveDate.Equal(start) || lv.LeaveDate.After(start)) &&
			(lv.LeaveDate.Equal(end) || lv.LeaveDate.Before(end)) {
			endpos = i
		}
	}
	if startpos >= 0 {
		if endpos < 0 {
			endpos = startpos
		}
		e.Leaves = append(e.Leaves[:startpos], e.Leaves[endpos+1:]...)
	}
}

func (e *Employee) ConvertFromData() error {
	if e.Data != nil {
		e.CompanyInfo = e.Data.CompanyInfo
		e.Leaves = e.Data.Leaves
		e.Assignments = e.Data.Assignments
		e.Variations = e.Data.Variations
		e.Balances = e.Data.Balances
		e.Requests = e.Data.Requests
		for _, lc := range e.Data.LaborCodes {
			for a, asgmt := range e.Assignments {
				newLc := &EmployeeLaborCode{
					ChargeNumber: lc.ChargeNumber,
					Extension:    lc.Extension,
				}
				asgmt.LaborCodes = append(asgmt.LaborCodes, *newLc)
				e.Assignments[a] = asgmt
			}
		}
		e.Data = nil
	}
	return nil
}

type EmployeeName struct {
	FirstName  string `json:"first"`
	MiddleName string `json:"middle"`
	LastName   string `json:"last"`
	Suffix     string `json:"suffix"`
}

func (en *EmployeeName) GetLastFirst() string {
	if en.MiddleName != "" {
		return en.LastName + ", " + en.FirstName + " " + en.MiddleName[0:1]
	}
	return en.LastName + ", " + en.FirstName
}

type EmployeeData struct {
	CompanyInfo CompanyInfo         `json:"companyinfo"`
	Assignments []Assignment        `json:"assignments,omitempty"`
	Variations  []Variation         `json:"variations,omitempty"`
	Balances    []AnnualLeave       `json:"balance,omitempty"`
	Leaves      []LeaveDay          `json:"leaves,omitempty"`
	Requests    []LeaveRequest      `json:"requests,omitempty"`
	LaborCodes  []EmployeeLaborCode `json:"laborCodes,omitempty"`
}

func (e *Employee) IsActive(date time.Time) bool {
	if e.Data != nil {
		e.ConvertFromData()
	}
	answer := false
	for _, asgmt := range e.Assignments {
		if asgmt.UseAssignment(e.SiteID, date) {
			answer = true
		}
	}
	return answer
}

func (e *Employee) IsAssigned(site, workcenter string, start, end time.Time) bool {
	if e.Data != nil {
		e.ConvertFromData()
	}
	answer := false
	for _, asgmt := range e.Assignments {
		if strings.EqualFold(asgmt.Site, site) &&
			strings.EqualFold(asgmt.Workcenter, workcenter) &&
			asgmt.StartDate.After(end) && asgmt.EndDate.Before((start)) {
			answer = true
		}
	}
	return answer
}

func (e *Employee) AtSite(site string, start, end time.Time) bool {
	if e.Data != nil {
		e.ConvertFromData()
	}
	answer := false
	for _, asgmt := range e.Assignments {
		if strings.EqualFold(asgmt.Site, site) &&
			asgmt.StartDate.Before(end) && asgmt.EndDate.After((start)) {
			answer = true
		}
	}
	return answer
}

func (e *Employee) GetWorkday(date time.Time, offset float64) *Workday {
	if e.Data != nil {
		e.ConvertFromData()
	}
	var wkday *Workday = nil
	var siteid string = ""
	for _, asgmt := range e.Assignments {
		if (asgmt.StartDate.Before(date) || asgmt.StartDate.Equal(date)) &&
			(asgmt.EndDate.After(date) || asgmt.EndDate.Equal(date)) {
			siteid = asgmt.Site
			wkday = asgmt.GetWorkday(date, offset)
		}
	}
	for _, vari := range e.Variations {
		if (vari.StartDate.Before(date) || vari.StartDate.Equal(date)) &&
			(vari.EndDate.After(date) || vari.EndDate.Equal(date)) {
			wkday = vari.GetWorkday(siteid, date)
		}
	}
	bLeave := false
	for _, lv := range e.Leaves {
		if lv.LeaveDate.Hour() != 0 {
			delta := time.Hour * time.Duration(offset)
			lv.LeaveDate = lv.LeaveDate.Add(delta)
		}
		if lv.LeaveDate.Year() == date.Year() &&
			lv.LeaveDate.Month() == date.Month() &&
			lv.LeaveDate.Day() == date.Day() {
			if !bLeave {
				wkday = &Workday{
					ID:         uint(0),
					Workcenter: "",
					Code:       lv.Code,
					Hours:      lv.Hours,
				}
				bLeave = true
			} else {
				if lv.Hours <= wkday.Hours {
					wkday.Hours += lv.Hours
				} else {
					wkday.Hours += lv.Hours
					wkday.Code = lv.Code
				}
			}
		}
	}
	return wkday
}

func (e *Employee) GetWorkdayActual(date time.Time, offset float64) *Workday {
	if e.Data != nil {
		e.ConvertFromData()
	}
	var wkday *Workday = nil
	var siteid string = ""
	for _, asgmt := range e.Assignments {
		if (asgmt.StartDate.Before(date) || asgmt.StartDate.Equal(date)) &&
			(asgmt.EndDate.After(date) || asgmt.EndDate.Equal(date)) {
			siteid = asgmt.Site
			wkday = asgmt.GetWorkday(date, offset)
		}
	}
	for _, vari := range e.Variations {
		if (vari.StartDate.Before(date) || vari.StartDate.Equal(date)) &&
			(vari.EndDate.After(date) || vari.EndDate.Equal(date)) {
			wkday = vari.GetWorkday(siteid, date)
		}
	}
	bLeave := false
	for _, lv := range e.Leaves {
		if lv.LeaveDate.Equal(date) &&
			strings.EqualFold(lv.Status, "actual") {
			if !bLeave {
				wkday = &Workday{
					ID:         uint(0),
					Workcenter: "",
					Code:       lv.Code,
					Hours:      lv.Hours,
				}
				bLeave = true
			} else {
				if lv.Hours <= wkday.Hours {
					wkday.Hours += lv.Hours
				} else {
					wkday.Hours += lv.Hours
					wkday.Code = lv.Code
				}
			}
		}
	}
	return wkday
}

func (e *Employee) GetWorkdayWOLeave(date time.Time, offset float64) *Workday {
	if e.Data != nil {
		e.ConvertFromData()
	}
	var wkday *Workday = nil
	var siteid string = ""
	for _, asgmt := range e.Assignments {
		if (asgmt.StartDate.Before(date) || asgmt.StartDate.Equal(date)) &&
			(asgmt.EndDate.After(date) || asgmt.EndDate.Equal(date)) {
			siteid = asgmt.Site
			wkday = asgmt.GetWorkday(date, offset)
		}
	}
	for _, vari := range e.Variations {
		if (vari.StartDate.Before(date) || vari.StartDate.Equal(date)) &&
			(vari.EndDate.After(date) || vari.EndDate.Equal(date)) {
			wkday = vari.GetWorkday(siteid, date)
		}
	}
	return wkday
}

func (e *Employee) GetStandardWorkday(date time.Time) float64 {
	if e.Data != nil {
		e.ConvertFromData()
	}
	answer := 8.0
	count := 0
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0,
		time.UTC)
	end := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	for start.Weekday() != time.Sunday {
		start = start.AddDate(0, 0, -1)
	}
	for end.Weekday() != time.Saturday {
		end = end.AddDate(0, 0, 1)
	}
	for start.Before(end) || start.Equal(end) {
		wd := e.GetWorkday(start, 0.0)
		if wd != nil && wd.Code != "" {
			count++
		}
		start = start.AddDate(0, 0, 1)
	}
	if count < 5 {
		answer = 10.0
	}
	return answer
}

func (e *Employee) AddAssignment(site, wkctr string, start time.Time) {
	if e.Data != nil {
		e.ConvertFromData()
	}
	// get next assignment id as one plus the highest in employee data
	max := 0
	for _, asgmt := range e.Assignments {
		if int(asgmt.ID) > max {
			max = int(asgmt.ID)
		}
	}

	// set the current highest or last end date to one day before
	// this assignment start date
	sort.Sort(ByAssignment(e.Assignments))
	if len(e.Assignments) > 0 {
		lastAsgmt := e.Assignments[len(e.Assignments)-1]
		lastAsgmt.EndDate = start.AddDate(0, 0, -1)
		e.Assignments[len(e.Assignments)-1] = lastAsgmt
	}

	// create the new assignment
	newAsgmt := Assignment{
		ID:           uint(max + 1),
		Site:         site,
		Workcenter:   wkctr,
		StartDate:    start,
		EndDate:      time.Date(9999, 12, 30, 0, 0, 0, 0, time.UTC),
		RotationDate: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
		RotationDays: 0,
	}
	// add a single schedule, plus it's seven workdays, set schedule
	// automatically to M-F/workcenter/8 hours/day shift.
	newAsgmt.AddSchedule(7)
	for i, wd := range newAsgmt.Schedules[0].Workdays {
		if i != 0 && i != 6 {
			wd.Code = "D"
			wd.Workcenter = wkctr
			wd.Hours = 8.0
			newAsgmt.Schedules[0].Workdays[i] = wd
		}
	}

	// add it employees assignment list and sort them
	e.Assignments = append(e.Assignments, newAsgmt)
	sort.Sort(ByAssignment(e.Assignments))
}

func (e *Employee) RemoveAssignment(id uint) {
	if e.Data != nil {
		e.ConvertFromData()
	}
	pos := -1
	if id > 1 {
		sort.Sort(ByAssignment(e.Assignments))
		for i, asgmt := range e.Assignments {
			if asgmt.ID == id {
				pos = i
			}
		}
		if pos >= 0 {
			asgmt := e.Assignments[pos-1]
			asgmt.EndDate = time.Date(9999, 12, 30, 0, 0, 0, 0, time.UTC)
			e.Assignments[pos-1] = asgmt
			e.Assignments = append(e.Assignments[:pos],
				e.Assignments[pos+1:]...)
		}
	}
}

func (e *Employee) PurgeOldData(date time.Time) {
	if e.Data != nil {
		e.ConvertFromData()
	}
	// purge old assignments based on assignment end date
	sort.Sort(ByAssignment(e.Assignments))
	for i := len(e.Assignments) - 1; i >= 0; i-- {
		if e.Assignments[i].EndDate.Before(date) {
			e.Assignments = append(e.Assignments[:i],
				e.Assignments[i+1:]...)
		}
	}
	// purge old variations based on variation end date
	sort.Sort(ByVariation(e.Variations))
	for i := len(e.Variations) - 1; i >= 0; i-- {
		if e.Variations[i].EndDate.Before(date) {
			e.Variations = append(e.Variations[:i],
				e.Variations[i+1:]...)
		}
	}
}

func (e *Employee) CreateLeaveBalance(year int) {
	if e.Data != nil {
		e.ConvertFromData()
	}
	found := false
	lastAnnual := 0.0
	lastCarry := 0.0
	for _, al := range e.Balances {
		if al.Year == year {
			found = true
		}
		if al.Year == year-1 {
			lastAnnual = al.Annual
			lastCarry = al.Carryover
		}
	}
	if !found {
		al := AnnualLeave{
			Year:      year,
			Annual:    lastAnnual,
			Carryover: 0.0,
		}
		if lastAnnual == 0.0 {
			al.Annual = 120.0
		} else {
			carry := lastAnnual + lastCarry
			for _, lv := range e.Leaves {
				if lv.LeaveDate.Year() == year-1 && strings.ToLower(lv.Code) == "v" &&
					strings.ToLower(lv.Status) == "actual" {
					carry -= lv.Hours
				}
			}
			al.Carryover = carry
		}
		e.Balances = append(e.Balances, al)
	}
}

func (e *Employee) UpdateAnnualLeave(year int, annual, carry float64) {
	if e.Data != nil {
		e.ConvertFromData()
	}
	found := false
	for _, al := range e.Balances {
		if al.Year == year {
			found = true
			al.Annual = annual
			al.Carryover = carry
		}
	}
	if !found {
		al := AnnualLeave{
			Year:      year,
			Annual:    annual,
			Carryover: carry,
		}
		e.Balances = append(e.Balances, al)
		sort.Sort(ByBalance(e.Balances))
	}
}

func (e *Employee) AddLeave(id int, date time.Time, code, status string,
	hours float64, requestID *primitive.ObjectID) {
	if e.Data != nil {
		e.ConvertFromData()
	}
	found := false
	max := 0
	for _, lv := range e.Leaves {
		if (lv.LeaveDate.Equal(date) &&
			strings.EqualFold(lv.Code, code)) || lv.ID == id {
			found = true
			lv.Status = status
			lv.Hours = hours
			if requestID != nil {
				lv.RequestID = requestID.Hex()
			}
		} else if lv.ID > max {
			max = lv.ID
		}
	}
	if !found {
		lv := LeaveDay{
			ID:        max + 1,
			LeaveDate: date,
			Code:      code,
			Hours:     hours,
			Status:    status,
			RequestID: requestID.Hex(),
		}
		e.Leaves = append(e.Leaves, lv)
		sort.Sort(ByLeaveDay(e.Leaves))
	}
}

func (e *Employee) UpdateLeave(id int, field, value string) error {
	if e.Data != nil {
		e.ConvertFromData()
	}
	found := false
	for i := 0; i < len(e.Leaves) && !found; i++ {
		lv := e.Leaves[i]
		if lv.ID == id {
			switch strings.ToLower(field) {
			case "date":
				date, err := time.ParseInLocation("01/02/2006", value, time.UTC)
				if err != nil {
					return err
				}
				lv.LeaveDate = date
			case "code":
				lv.Code = value
			case "hours":
				hrs, err := strconv.ParseFloat(value, 64)
				if err != nil {
					return err
				}
				lv.Hours = hrs
			case "status":
				lv.Status = value
			case "requestid":
				lv.RequestID = value
			}
			e.Leaves[i] = lv
		}
	}
	return nil
}

func (e *Employee) DeleteLeave(id int) {
	if e.Data != nil {
		e.ConvertFromData()
	}
	pos := -1
	for i, lv := range e.Leaves {
		if lv.ID == id {
			pos = i
		}
	}
	if pos >= 0 {
		e.Leaves = append(e.Leaves[:pos], e.Leaves[pos+1:]...)
	}
}

func (e *Employee) GetLeaveHours(start, end time.Time) float64 {
	if e.Data != nil {
		e.ConvertFromData()
	}
	answer := 0.0

	sort.Sort(ByLeaveDay(e.Leaves))
	for _, lv := range e.Leaves {
		if (lv.LeaveDate.After(start) ||
			lv.LeaveDate.Equal(start)) &&
			lv.LeaveDate.Before(end) &&
			strings.EqualFold(lv.Status, "actual") {
			answer += lv.Hours
		}
	}
	return answer
}

func (e *Employee) GetPTOHours(start, end time.Time) float64 {
	if e.Data != nil {
		e.ConvertFromData()
	}
	answer := 0.0

	sort.Sort(ByLeaveDay(e.Leaves))
	for _, lv := range e.Leaves {
		if (lv.LeaveDate.After(start) ||
			lv.LeaveDate.Equal(start)) &&
			lv.LeaveDate.Before(end) &&
			strings.EqualFold(lv.Status, "actual") &&
			strings.EqualFold(lv.Code, "v") {
			answer += lv.Hours
		}
	}
	return answer
}

func (e *Employee) NewLeaveRequest(empID, code string, start, end time.Time,
	offset float64) {
	if e.Data != nil {
		e.ConvertFromData()
	}
	lr := LeaveRequest{
		ID:          primitive.NewObjectID().Hex(),
		EmployeeID:  empID,
		RequestDate: time.Now().UTC(),
		PrimaryCode: code,
		StartDate:   start,
		EndDate:     end,
		Status:      "DRAFT",
	}
	zoneID := "UTC"
	if offset > 0 {
		zoneID += "+" + fmt.Sprintf("%0.1f", offset)
	} else if offset < 0 {
		zoneID += fmt.Sprintf("%0.1f", offset)
	}
	sDate := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0,
		time.UTC)
	std := e.GetStandardWorkday(sDate)
	for sDate.Before(end) || sDate.Equal(end) {
		wd := e.GetWorkday(sDate, offset)
		if wd.Code != "" {
			hours := wd.Hours
			if hours == 0.0 {
				hours = std
			}
			if code == "H" {
				hours = 8.0
			}
			lv := LeaveDay{
				LeaveDate: sDate,
				Code:      code,
				Hours:     hours,
				Status:    "DRAFT",
				RequestID: lr.ID,
			}
			lr.RequestedDays = append(lr.RequestedDays, lv)
		}
		sDate = sDate.AddDate(0, 0, 1)
	}
	e.Requests = append(e.Requests, lr)
	sort.Sort(ByLeaveRequest(e.Requests))
}

func (e *Employee) UpdateLeaveRequest(request, field, value string,
	offset float64) (string, error) {
	if e.Data != nil {
		e.ConvertFromData()
	}
	message := ""
	for i, req := range e.Requests {
		if req.ID == request {
			switch strings.ToLower(field) {
			case "startdate", "start":
				lvDate, err := time.Parse("2006-01-02", value)
				if err != nil {
					return "", err
				}
				if lvDate.Before(req.StartDate) || lvDate.After(req.EndDate) {
					if strings.EqualFold(req.Status, "approved") {
						req.Status = "REQUESTED"
						req.ApprovalDate = time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
						req.ApprovedBy = ""
						message = fmt.Sprintf("Leave Request from %s: Starting date changed "+
							"needs reapproval", e.Name.GetLastFirst())
					}
					startPos := -1
					endPos := -1
					sort.Sort(ByLeaveDay(e.Leaves))
					for i, lv := range e.Leaves {
						if lv.RequestID == req.ID {
							if startPos < 0 {
								startPos = i
							} else {
								endPos = i
							}
						}
					}
					if startPos >= 0 {
						if endPos < 0 {
							endPos = startPos
						}
						endPos++
						if endPos > len(e.Leaves) {

						} else {
							e.Leaves = append(e.Leaves[:startPos],
								e.Leaves[endPos:]...)
						}
					}
				}
				req.StartDate = lvDate
				// reset the leave dates
				req.SetLeaveDays(e, offset)
				if req.Status == "APPROVED" {
					e.ChangeApprovedLeaveDates(req)
				}
			case "enddate", "end":
				lvDate, err := time.Parse("2006-01-02", value)
				if err != nil {
					return "", err
				}
				if lvDate.Before(req.StartDate) || lvDate.After(req.EndDate) {
					if strings.EqualFold(req.Status, "approved") {
						req.Status = "REQUESTED"
						req.ApprovalDate = time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
						req.ApprovedBy = ""
						message = fmt.Sprintf("Leave Request from %s: Ending Date changed "+
							"needs reapproval", e.Name.GetLastFirst())
					}
					startPos := -1
					endPos := -1
					sort.Sort(ByLeaveDay(e.Leaves))
					for i, lv := range e.Leaves {
						if lv.RequestID == req.ID {
							if startPos < 0 {
								startPos = i
							} else {
								endPos = i
							}
						}
					}
					if startPos >= 0 {
						if endPos < 0 {
							endPos = startPos
						}
						endPos++
						if endPos > len(e.Leaves) {

						} else {
							e.Leaves = append(e.Leaves[:startPos],
								e.Leaves[endPos:]...)
						}
					}
				}
				req.EndDate = lvDate
				// reset the leave dates
				req.SetLeaveDays(e, offset)
				if req.Status == "APPROVED" {
					e.ChangeApprovedLeaveDates(req)
				}
			case "code", "primarycode":
				req.PrimaryCode = value
			case "dates":
				parts := strings.Split(value, "|")
				start, err := time.ParseInLocation("2006-01-02", parts[0], time.UTC)
				if err != nil {
					return "", err
				}
				end, err := time.ParseInLocation("2006-01-02", parts[1], time.UTC)
				if err != nil {
					return "", nil
				}
				start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0,
					time.UTC)
				end = time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0,
					time.UTC)
				if start.Before(req.StartDate) || start.After(req.EndDate) ||
					end.Before(req.StartDate) || end.After(req.EndDate) {
					if strings.EqualFold(req.Status, "approved") {
						req.Status = "REQUESTED"
						req.ApprovalDate = time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
						req.ApprovedBy = ""
						message = fmt.Sprintf("Leave Request from %s: dates changed "+
							"needs reapproval", e.Name.GetLastFirst())
					}
					startPos := -1
					endPos := -1
					sort.Sort(ByLeaveDay(e.Leaves))
					for i, lv := range e.Leaves {
						if lv.RequestID == req.ID {
							if startPos < 0 {
								startPos = i
							} else {
								endPos = i
							}
						}
					}
					if startPos >= 0 {
						if endPos < 0 {
							endPos = startPos
						}
						endPos++
						if endPos > len(e.Leaves) {

						} else {
							e.Leaves = append(e.Leaves[:startPos],
								e.Leaves[endPos:]...)
						}
					}
				}
				req.StartDate = time.Date(start.Year(), start.Month(), start.Day(), 0,
					0, 0, 0, time.UTC)
				req.EndDate = time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0,
					time.UTC)
				req.SetLeaveDays(e, offset)
				if req.Status == "APPROVED" {
					e.ChangeApprovedLeaveDates(req)
				}
			case "requested":
				req.Status = "REQUESTED"
				for d, day := range req.RequestedDays {
					day.Status = "REQUESTED"
					req.RequestedDays[d] = day
				}
				message = fmt.Sprintf("Leave Request: Leave Request from %s ",
					e.Name.GetLastFirst()) + "submitted for approval."
			case "approve":
				req.ApprovedBy = value
				req.ApprovalDate = time.Now().UTC()
				req.Status = "APPROVED"
				for d, day := range req.RequestedDays {
					day.Status = "APPROVED"
					req.RequestedDays[d] = day
				}
				message = "Leave Request: Leave Request approved."
				e.ChangeApprovedLeaveDates(req)
			case "unapprove":
				req.ApprovedBy = ""
				req.ApprovalDate = time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
				req.Status = "REQUESTED"
				for d, day := range req.RequestedDays {
					day.Status = "REQUESTED"
					req.RequestedDays[d] = day
				}
				cmt := LeaveRequestComment{
					CommentDate: time.Now().UTC(),
					Comment:     value,
				}
				req.Comments = append(req.Comments, cmt)
				message = "Leave Request: Leave Request unapproved.\n" +
					"Comment: " + value
			case "day", "requestday":
				parts := strings.Split(value, "|")
				lvDate, _ := time.Parse("2006-01-02", parts[0])
				code := parts[1]
				hours, _ := strconv.ParseFloat(parts[2], 64)
				found := false
				status := ""
				for j, lv := range req.RequestedDays {
					if lv.LeaveDate.Equal(lvDate) {
						found = true
						lv.Code = code
						if status == "" {
							status = lv.Status
						}
						if code == "" {
							lv.Hours = 0.0
						} else {
							lv.Hours = hours
						}
						req.RequestedDays[j] = lv
					}
				}
				if !found {
					lv := LeaveDay{
						LeaveDate: lvDate,
						Code:      code,
						Hours:     hours,
						Status:    status,
						RequestID: req.ID,
					}
					req.RequestedDays = append(req.RequestedDays, lv)
				}
			}
			e.Requests[i] = req
		}
	}
	return message, nil
}

func (e *Employee) ChangeApprovedLeaveDates(lr LeaveRequest) {
	if e.Data != nil {
		e.ConvertFromData()
	}
	// approved leave affects the leave listing, so we will
	// remove old leaves for the period then add the new ones
	startPos := -1
	endPos := -1
	maxId := -1
	sort.Sort(ByLeaveDay(e.Leaves))
	for i, lv := range e.Leaves {
		if (lv.LeaveDate.After(lr.StartDate) || lv.LeaveDate.Equal(lr.StartDate)) &&
			(lv.LeaveDate.Before(lr.EndDate) || lv.LeaveDate.Equal(lr.EndDate)) {
			if startPos < 0 {
				startPos = i
			} else {
				endPos = i
			}
		}
		if maxId < lv.ID {
			maxId = lv.ID
		}
	}
	if startPos > 0 {
		if endPos < 0 {
			endPos = startPos
		}
		endPos++
		e.Leaves = append(e.Leaves[:startPos], e.Leaves[endPos:]...)
	}

	// now add the leave request's leave days to the leave list
	for _, lv := range lr.RequestedDays {
		maxId++
		lv.ID = maxId
		lv.Status = lr.Status
		e.Leaves = append(e.Leaves, lv)
	}
	sort.Sort(ByLeaveDay(e.Leaves))
}

func (e *Employee) DeleteLeaveRequest(request string) error {
	if e.Data != nil {
		e.ConvertFromData()
	}
	pos := -1
	for i, req := range e.Requests {
		if req.ID == request {
			pos = i
		}
	}
	if pos < 0 {
		return errors.New("request not found")
	}
	e.Requests = append(e.Requests[:pos], e.Requests[pos+1:]...)
	// delete all leaves associated with this leave request, except if the leave
	// has a status of actual
	sort.Sort(ByLeaveDay(e.Leaves))
	var deletes []int
	for i, lv := range e.Leaves {
		if lv.RequestID == request && strings.ToLower(lv.Status) != "actual" {
			deletes = append(deletes, i)
		}
	}
	if len(deletes) > 0 {
		for i := len(deletes) - 1; i >= 0; i-- {
			e.Leaves = append(e.Leaves[:deletes[i]],
				e.Leaves[deletes[i]+1:]...)
		}
	}
	return nil
}

func (e *Employee) HasLaborCode(chargeNumber, extension string) bool {
	if e.Data != nil {
		e.ConvertFromData()
	}
	found := false
	for _, asgmt := range e.Assignments {
		for _, lc := range asgmt.LaborCodes {
			if strings.EqualFold(lc.ChargeNumber, chargeNumber) &&
				strings.EqualFold(lc.Extension, extension) {
				found = true
			}
		}
	}
	return found
}

func (e *Employee) DeleteLaborCode(chargeNo, ext string) {
	if e.Data != nil {
		e.ConvertFromData()
	}
	if e.HasLaborCode(chargeNo, ext) {
		for a, asgmt := range e.Assignments {
			pos := -1
			for i, lc := range asgmt.LaborCodes {
				if lc.ChargeNumber == chargeNo && lc.Extension == ext {
					pos = i
				}
			}
			if pos >= 0 {
				asgmt.LaborCodes = append(asgmt.LaborCodes[:pos], asgmt.LaborCodes[pos+1:]...)
				e.Assignments[a] = asgmt
			}
		}
	}
}

func (e *Employee) DeleteLeavesBetweenDates(start, end time.Time) {
	if e.Data != nil {
		e.ConvertFromData()
	}
	for i := len(e.Leaves) - 1; i >= 0; i-- {
		if e.Leaves[i].LeaveDate.Equal(start) ||
			e.Leaves[i].LeaveDate.Equal(end) ||
			(e.Leaves[i].LeaveDate.After(start) &&
				e.Leaves[i].LeaveDate.Before(end)) {
			e.Leaves = append(e.Leaves[:i], e.Leaves[i+1:]...)
		}
	}
}

func (e *Employee) GetAssignment(start, end time.Time) (string, string) {
	assigned := make(map[string]int)
	current := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0,
		time.UTC)
	for current.Before(end) {
		wd := e.GetWorkdayWOLeave(current, 0.0)
		if wd != nil {
			label := wd.Workcenter + "-" + wd.Code
			if label != "-" {
				val, ok := assigned[label]
				if ok {
					assigned[label] = val + 1
				} else {
					assigned[label] = 1
				}
			}
		}
		current = current.AddDate(0, 0, 1)
	}
	max := 0
	answer := ""
	for k, v := range assigned {
		if v > max {
			answer = k
			max = v
		}
	}
	if answer != "" {
		parts := strings.Split(answer, "-")
		return parts[0], parts[1]
	}
	return "", ""
}

func (e *Employee) GetWorkedHours(start, end time.Time) float64 {
	answer := 0.0

	for _, wk := range e.Work {
		if (wk.DateWorked.Equal(start) ||
			wk.DateWorked.After(start)) &&
			wk.DateWorked.Before(end) {
			answer += wk.Hours
		}
	}

	return answer
}

func (e *Employee) GetWorkedHoursForLabor(chgno, ext string,
	start, end time.Time) float64 {
	answer := 0.0

	for _, wk := range e.Work {
		if (wk.DateWorked.Equal(start) ||
			wk.DateWorked.After(start)) &&
			wk.DateWorked.Before(end) &&
			strings.EqualFold(chgno, wk.ChargeNumber) &&
			strings.EqualFold(ext, wk.Extension) {
			answer += wk.Hours
		}
	}
	return answer
}

func (e *Employee) GetForecastHours(chgno, ext string,
	start, end time.Time, workcodes []EmployeeCompareCode,
	offset float64) float64 {
	if e.Data != nil {
		e.ConvertFromData()
	}
	answer := 0.0

	// first check to see if assigned this labor code, if not
	// return 0 hours
	found := false
	for _, asgmt := range e.Assignments {
		for _, lc := range asgmt.LaborCodes {
			if strings.EqualFold(chgno, lc.ChargeNumber) &&
				strings.EqualFold(ext, lc.Extension) {
				found = true
			}
		}
	}
	if !found {
		return 0.0
	}

	// determine last day of actual recorded work so than
	// forecast hours don't overlap.
	lastWork := time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
	if len(e.Work) > 0 {
		sort.Sort(ByEmployeeWork(e.Work))
		lastWork = e.Work[len(e.Work)-1].DateWorked
	}

	// now step through the days of the period to:
	// 1) see if they had worked any charge numbers during
	//		the period, if working add 0 hours
	// 2) see if they were supposed to be working on this
	//		date, compare workday code to workcodes to ensure
	//		they weren't on leave.  If not on leave, add
	// 		standard work day.
	current := time.Date(start.Year(), start.Month(),
		start.Day(), 0, 0, 0, 0, time.UTC)
	for current.Before(end) {
		if current.After(lastWork) {
			hours := e.GetWorkedHours(current, current.AddDate(0, 0, 1))
			if hours == 0.0 {
				wd := e.GetWorkday(current, offset)
				if wd != nil && wd.Code != "" {
					for _, wc := range workcodes {
						if strings.EqualFold(wc.Code, wd.Code) && !wc.IsLeave {
							std := e.GetStandardWorkday(current)
							for _, asgmt := range e.Assignments {
								if current.Equal(asgmt.StartDate) || current.Equal(asgmt.EndDate) ||
									(current.After(asgmt.StartDate) && current.Before(asgmt.EndDate)) {
									for _, lc := range asgmt.LaborCodes {
										if strings.EqualFold(chgno, lc.ChargeNumber) &&
											strings.EqualFold(ext, lc.Extension) {
											answer += std
										}
									}
								}
							}
						}
					}
				}
			}
		}
		current = current.AddDate(0, 0, 1)
	}

	return answer
}

func (e *Employee) GetLastWorkday() time.Time {
	if e.Data != nil {
		e.ConvertFromData()
	}
	sort.Sort(ByEmployeeWork(e.Work))
	answer := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	if len(e.Work) > 0 {
		work := e.Work[len(e.Work)-1]
		answer = time.Date(work.DateWorked.Year(), work.DateWorked.Month(),
			work.DateWorked.Day(), 0, 0, 0, 0, time.UTC)
	}
	return answer
}

type EmployeeCompareCode struct {
	Code    string
	IsLeave bool
}
