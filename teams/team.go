package teams

import (
	"github.com/erneap/go-models/sites"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Team struct {
	ID        primitive.ObjectID `json:"id" bson:"_id"`
	Name      string             `json:"name" bson:"name"`
	Workcodes []Workcode         `json:"workcodes" bson:"workcodes"`
	Sites     []sites.Site       `json:"sites" bson:"sites"`
	Companies []Company          `json:"companies,omitempty" bson:"companies,omitempty"`
}

type ByTeam []Team

func (c ByTeam) Len() int { return len(c) }
func (c ByTeam) Less(i, j int) bool {
	return c[i].Name < c[j].Name
}
func (c ByTeam) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
