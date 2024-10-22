package systemdata

type ImageType struct {
	ID           string      `json:"id"`
	Collected    uint        `json:"collected,omitempty"`
	NotCollected uint        `json:"notcollected,omitempty"`
	SortID       uint        `json:"sortID"`
	Subtypes     []ImageType `json:"subtypes,omitempty"`
}

type ByImageType []ImageType

func (c ByImageType) Len() int { return len(c) }
func (c ByImageType) Less(i, j int) bool {
	return c[i].SortID < c[j].SortID
}
func (c ByImageType) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
