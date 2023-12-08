package bible

type BibleBook struct {
	Id       int            `json:"id" bson:"id"`
	Code     string         `json:"code,omitempty" bson:"code:omitempty"`
	Title    string         `json:"title" bson:"title"`
	Chapters []BibleChapter `json:"chapters,omitempty" bson:"chapter,omitempty"`
}

type ByBibleBook []BibleBook

func (c ByBibleBook) Len() int { return len(c) }
func (c ByBibleBook) Less(i, j int) bool {
	return c[i].Id < c[j].Id
}
func (c ByBibleBook) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (b *BibleBook) AddChapter() (*BibleChapter, error) {
	chptr := &BibleChapter{
		Id: len(b.Chapters) + 1,
	}
	b.Chapters = append(b.Chapters, *chptr)
	return chptr, nil
}
