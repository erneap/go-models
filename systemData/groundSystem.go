package systemdata

type GroundSystemExploitation struct {
	PlatformID      string `json:"platformID"`
	SensorType      string `json:"sensorType"`
	Exploitation    string `json:"exploitation"`
	CommunicationID string `json:"communicationID"`
}

type ByGSExploitation []GroundSystemExploitation

func (c ByGSExploitation) Len() int { return len(c) }
func (c ByGSExploitation) Less(i, j int) bool {
	if c[i].PlatformID == c[j].PlatformID {
		if c[i].SensorType == c[j].SensorType {
			if c[i].Exploitation == c[j].Exploitation {
				return c[i].CommunicationID < c[j].CommunicationID
			}
			return c[i].Exploitation < c[j].Exploitation
		}
		return c[i].SensorType < c[j].SensorType
	}
	return c[i].PlatformID < c[j].PlatformID
}
func (c ByGSExploitation) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

type GroundSystem struct {
	ID            string                     `json:"id"`
	Enclaves      []string                   `json:"enclaves"`
	ShowOnGEOINT  bool                       `json:"showOnGEOINT"`
	ShowOnGSEG    bool                       `json:"showOnGSEG"`
	ShowOnMIST    bool                       `json:"showOnMIST"`
	ShowOnXINT    bool                       `json:"showOnXINT"`
	Exploitations []GroundSystemExploitation `json:"exploitations"`
}

type ByGroundSystem []GroundSystem

func (c ByGroundSystem) Len() int { return len(c) }
func (c ByGroundSystem) Less(i, j int) bool {
	return c[i].ID < c[j].ID
}
func (c ByGroundSystem) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (gs *GroundSystem) UseMissionSensor(platform, sensor, exploit, comm string) bool {
	if len(gs.Exploitations) == 0 {
		return true
	}
	answer := false
	for _, exp := range gs.Exploitations {
		if exp.PlatformID == platform && exp.SensorType == sensor &&
			exp.Exploitation == exploit && exp.CommunicationID == comm {
			answer = true
		}
	}
	return answer
}
