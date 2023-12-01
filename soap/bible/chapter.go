package bible

import (
	"sort"

	"github.com/erneap/go-models/soap/plans"
)

type BibleChapter struct {
	Id     int             `json:"id" bson:"id"`
	Verses []plans.Passage `json:"verses,omitempty" bson:"verses,omitempty"`
}

type ByBibleChapter []BibleChapter

func (c ByBibleChapter) Len() int { return len(c) }
func (c ByBibleChapter) Less(i, j int) bool {
	return c[i].Id < c[j].Id
}
func (c ByBibleChapter) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (bc *BibleChapter) IsComplete() bool {
	if len(bc.Verses) == 0 {
		return false
	}
	// sort the verses, then check the verses starting at 1 and if not present
	// then not complete
	sort.Sort(plans.ByPassage(bc.Verses))
	current := 0
	for _, verse := range bc.Verses {
		if current+1 != verse.StartVerse {
			return false
		}
		current = verse.EndVerse
	}
	return true
}
