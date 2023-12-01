package plans

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ReadingPlan struct {
	UserID    primitive.ObjectID `json:"userid,omitempty" bson:"userid,omitempty"`
	StartDate *time.Time         `json:"start,omitempty" bson:"start,omitempty"`
	Months    []ReadingMonth     `json:"months,omitempty" bson:"months,omitempty"`
}

type ByReadingPlan []ReadingPlan

func (c ByReadingPlan) Len() int { return len(c) }
func (c ByReadingPlan) Less(i, j int) bool {
	if c[i].UserID.Hex() == c[j].UserID.Hex() {
		return c[i].StartDate.Before(*c[j].StartDate)
	}
	return c[i].UserID.Hex() < c[j].UserID.Hex()
}
func (c ByReadingPlan) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (p *ReadingPlan) IsCompleted() bool {
	answer := true
	for _, month := range p.Months {
		if !month.IsCompleted() {
			answer = false
		}
	}
	return answer
}
