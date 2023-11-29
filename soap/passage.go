package soap

import "strings"

type Passage struct {
	Version    string `json:"version,omitempty" bson:"version,omitempty"`
	BookID     int    `json:"bookid" bson:"bookid"`
	Book       string `json:"book" bson:"book"`
	Chapter    int    `json:"chapter" bson:"chapter"`
	StartVerse int    `json:"startverse,omitempty" bson:"startverse,omitempty"`
	EndVerse   int    `json:"endverse,omitempty" bson:"endverse,omitempty"`
	Passage    string `json:"passage,omitempty" bson:"passage,omitempty"`
}

type ByPassage []Passage

func (c ByPassage) Len() int { return len(c) }
func (c ByPassage) Less(i, j int) bool {
	if strings.EqualFold(c[i].Version, c[j].Version) {
		if c[i].BookID == c[j].BookID {
			if c[i].Chapter == c[j].Chapter {
				if c[i].StartVerse == c[j].StartVerse {
					return c[i].EndVerse < c[j].EndVerse
				}
				return c[i].StartVerse < c[j].StartVerse
			}
			return c[i].Chapter < c[j].Chapter
		}
		return c[i].BookID < c[j].BookID
	}
	return c[i].Version < c[j].Version
}
func (c ByPassage) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
