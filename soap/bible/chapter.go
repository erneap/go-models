package bible

import (
	"errors"
	"sort"
	"strconv"
	"strings"

	"github.com/erneap/go-models/soap/plans"
)

type BibleChapter struct {
	Id       int             `json:"id" bson:"id"`
	Passages []plans.Passage `json:"passages,omitempty" bson:"passages,omitempty"`
}

type ByBibleChapter []BibleChapter

func (c ByBibleChapter) Len() int { return len(c) }
func (c ByBibleChapter) Less(i, j int) bool {
	return c[i].Id < c[j].Id
}
func (c ByBibleChapter) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (bc *BibleChapter) IsComplete() bool {
	if len(bc.Passages) == 0 {
		return false
	}
	// sort the verses, then check the verses starting at 1 and if not present
	// then not complete
	sort.Sort(plans.ByPassage(bc.Passages))
	current := 0
	for _, verse := range bc.Passages {
		if current+1 != verse.StartVerse {
			return false
		}
		current = verse.EndVerse
	}
	return true
}

func (d *BibleChapter) AddPassage(bookid int, book string, chapter, start, end int) error {
	found := false
	if bookid == 0 || chapter == 0 {
		return errors.New("not enough information to add passage")
	}
	for p, psg := range d.Passages {
		if psg.BookID == bookid && psg.Chapter == chapter {
			found = true
			if start != 0 || end != 0 {
				psg.StartVerse = start
				psg.EndVerse = end
				d.Passages[p] = psg
			}
		}
	}
	if !found {
		psg := &plans.Passage{
			ID:         len(d.Passages) + 1,
			BookID:     bookid,
			Book:       book,
			Chapter:    chapter,
			StartVerse: start,
			EndVerse:   end,
		}
		d.Passages = append(d.Passages, *psg)
		sort.Sort(plans.ByPassage(d.Passages))
	}
	return nil
}

func (d *BibleChapter) UpdatePassage(id int, field, value string) error {
	found := false
	if id == 0 {
		return errors.New("not enough information to add passage text")
	}
	for p, psg := range d.Passages {
		if psg.ID == id {
			found = true
			switch strings.ToLower(field) {
			case "bookid":
				iValue, err := strconv.Atoi(value)
				if err != nil {
					return err
				}
				psg.BookID = iValue
			case "book":
				psg.Book = value
			case "chapter":
				iValue, err := strconv.Atoi(value)
				if err != nil {
					return err
				}
				psg.Chapter = iValue
			case "startverse", "start":
				iValue, err := strconv.Atoi(value)
				if err != nil {
					return err
				}
				psg.StartVerse = iValue
			case "endverse", "end":
				iValue, err := strconv.Atoi(value)
				if err != nil {
					return err
				}
				psg.EndVerse = iValue
			case "text":
				psg.Passage = value
			case "completed":
				psg.Completed = strings.EqualFold(value, "true")
			}
			d.Passages[p] = psg
		}
	}
	if !found {
		return errors.New("passage not found")
	}
	return nil
}

func (d *BibleChapter) UpdatePassageText(id int, text string) error {
	found := false
	if id == 0 {
		return errors.New("not enough information to add passage text")
	}
	for p, psg := range d.Passages {
		if psg.ID == id {
			found = true
			psg.Passage = text
			d.Passages[p] = psg
		}
	}
	if !found {
		return errors.New("passage not found")
	}
	return nil
}

func (d *BibleChapter) DeletePassage(id int) error {
	pos := -1
	if id == 0 {
		return errors.New("not enough information to add passage text")
	}
	for p, psg := range d.Passages {
		if psg.ID == id {
			pos = p
		}
	}
	if pos >= 0 {
		d.Passages = append(d.Passages[:pos], d.Passages[pos+1:]...)
	} else {
		return errors.New("passage not found")
	}
	sort.Sort(plans.ByPassage(d.Passages))
	for p, psg := range d.Passages {
		pos++
		psg.ID = p + 1
		d.Passages[p] = psg
	}
	return nil
}
