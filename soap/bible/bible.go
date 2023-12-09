package bible

import (
	"errors"
	"sort"
	"strings"

	"github.com/erneap/go-models/soap/plans"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Bible struct {
	Id      primitive.ObjectID `json:"id" bson:"_id"`
	Version string             `json:"version,omitempty" bson:"version,omitempty"`
	Name    string             `json:"name,omitempty" bson:"name,omitempty"`
	Books   []BibleBook        `json:"books,omitempty" bson:"books,omitempty"`
}

type ByBible []Bible

func (c ByBible) Len() int { return len(c) }
func (c ByBible) Less(i, j int) bool {
	return c[i].Version < c[j].Version
}
func (c ByBible) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (b *Bible) AddPassage(book string, chptr, start, end int,
	text string) *plans.Passage {
	var passage *plans.Passage
	found := false
	for bid, bk := range b.Books {
		if strings.EqualFold(bk.Title[:2], book[:2]) {
			for c, ch := range bk.Chapters {
				if ch.Id == chptr {
					for p, psg := range ch.Passages {
						if psg.StartVerse == start && psg.EndVerse == end {
							found = true
							psg.Passage = text
							ch.Passages[p] = psg
						}
					}
					if !found {
						psg := &plans.Passage{
							ID:         len(ch.Passages) + 1,
							BookID:     bk.Id,
							Book:       book,
							Chapter:    chptr,
							StartVerse: start,
							EndVerse:   end,
							Passage:    text,
						}
						passage = psg
						found = true
						ch.Passages = append(ch.Passages, *psg)
						sort.Sort(plans.ByPassage(ch.Passages))
					}
					bk.Chapters[c] = ch
				}
			}
			if !found {
				// chapter and passage not found, so add chapter and passage at once
				ch := &BibleChapter{
					Id: chptr,
				}
				psg := &plans.Passage{
					ID:         len(ch.Passages) + 1,
					BookID:     bk.Id,
					Book:       book,
					Chapter:    chptr,
					StartVerse: start,
					EndVerse:   end,
					Passage:    text,
				}
				passage = psg
				ch.Passages = append(ch.Passages, *psg)
				bk.Chapters = append(bk.Chapters, *ch)
				sort.Sort(ByBibleChapter(bk.Chapters))
				found = true
			}
			b.Books[bid] = bk
		}
	}
	if !found {
		bk := &BibleBook{
			Id:    len(b.Books) + 1,
			Code:  strings.ToLower(book[:2]),
			Title: book,
		}
		ch := &BibleChapter{
			Id: chptr,
		}
		psg := &plans.Passage{
			ID:         len(ch.Passages) + 1,
			BookID:     bk.Id,
			Book:       book,
			Chapter:    chptr,
			StartVerse: start,
			EndVerse:   end,
			Passage:    text,
		}
		ch.Passages = append(ch.Passages, *psg)
		passage = psg
		bk.Chapters = append(bk.Chapters, *ch)
		b.Books = append(b.Books, *bk)
	}
	return passage
}

func (b *Bible) GetPassageText(book string, chptr, start,
	end int) (string, error) {
	answer := ""
	found := false
	for _, bk := range b.Books {
		if strings.EqualFold(bk.Title[:2], book[:2]) {
			for _, ch := range bk.Chapters {
				if ch.Id == chptr {
					if start == 0 && len(ch.Passages) > 0 {
						found = true
						answer = ch.Passages[0].Passage
					} else if start > 0 {
						for _, psg := range ch.Passages {
							if psg.StartVerse == start && psg.EndVerse == end {
								found = true
								answer = psg.Passage
							}
						}
					}
				}
			}
		}
	}
	if !found || answer == "" {
		return "", errors.New("not Found")
	}
	return answer, nil
}

func (b *Bible) RemovePassage(book string, chptr, start,
	end int) (*plans.Passage, error) {
	var passage *plans.Passage
	for i, bk := range b.Books {
		if strings.EqualFold(bk.Title[:2], book[:2]) {
			for c, ch := range bk.Chapters {
				if ch.Id == chptr {
					pos := -1
					for p, psg := range ch.Passages {
						if psg.StartVerse == start && psg.EndVerse == end {
							pos = p
							passage = &psg
						}
					}
					if pos >= 0 {
						ch.Passages = append(ch.Passages[:pos], ch.Passages[pos+1:]...)
					}
				}
				bk.Chapters[c] = ch
			}
			b.Books[i] = bk
		}
	}
	if passage == nil {
		return nil, errors.New("not found")
	}
	return passage, nil
}

type BibleStandards struct {
	Books    []StandardBibleBook `json:"books,omitempty" bson:"books,omitempty"`
	Versions []BibleVersion      `json:"versions" bson:"versions"`
}

type BibleVersion struct {
	Code          string `json:"code" bson:"code"`
	Title         string `json:"title" bson:"title"`
	Language      string `json:"language" bson:"language"`
	LanguageTitle string `json:"languagetitle" bson:"languagetitle"`
}
type ByBibleVersion []BibleVersion

func (c ByBibleVersion) Len() int { return len(c) }
func (c ByBibleVersion) Less(i, j int) bool {
	if strings.EqualFold(c[i].Language, c[j].Language) {
		return strings.ToLower(c[i].Title) < strings.ToLower(c[j].Title)
	}
	return strings.ToLower(c[i].Language) < strings.ToLower(c[j].Language)
}
func (c ByBibleVersion) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
