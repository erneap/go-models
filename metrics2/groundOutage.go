package metrics2

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GroundOutage2 struct {
	ID             primitive.ObjectID `json:"id" bson:"_id"`
	OutageDate     time.Time          `json:"outageDate" bson:"outageDate"`
	GroundSystem   string             `json:"groundSystem" bson:"groundSystem"`
	Classification string             `json:"classification" bson:"classification"`
	OutageNumber   uint               `json:"outageNumber" bson:"outageNumber"`
	OutageMinutes  uint               `json:"outageMinutes" bson:"outageMinutes"`
	Subsystem      string             `json:"subSystem" bson:"subSystem"`
	ReferenceID    string             `json:"referenceId" bson:"referenceId"`
	MajorSystem    string             `json:"majorSystem" bson:"majorSystem"`
	Problem        string             `json:"-" bson:"problem"`
	FixAction      string             `json:"-" bson:"fixAction"`
	MissionOutage  bool               `json:"missionOutage" bson:"missionOutage"`
	Capability     string             `json:"capability,omitempty" bson:"capability,omitempty"`
}

type ByOutage2 []GroundOutage2

func (c ByOutage2) Len() int { return len(c) }
func (c ByOutage2) Less(i, j int) bool {
	if c[i].OutageDate.Equal(c[j].OutageDate) {
		return c[i].OutageNumber < c[j].OutageNumber
	}
	return c[i].OutageDate.Before(c[j].OutageDate)
}
func (c ByOutage2) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
