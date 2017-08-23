package nhc

// Message NHC Message interface
type Message interface {
	Save()
}

// Action holds one individual nhc action (equipment)
type Action struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Type     int    `json:"type"`
	Location int    `json:"location"`
	Value1   int    `json:"value1"`
	Value2   int    `json:"value2"`
	Value3   int    `json:"value3"`
}

// Event holds an individual event
type Event struct {
	ID    int `json:"id"`
	Value int `json:"value1"`
}

// Location holds one nhc location
type Location struct {
	ID   int
	Name string
}

// Save process to db
func (loc Location) Save() {
	SaveLocation(loc)
}

// Save process to db
func (act Action) Save() {
	SaveAction(act)
}

// Save process to db
func (evt Event) Save() {
	ProcessEvent(evt)
}

// Item NHC equipment definition
type Item struct {
	Provider string `json:"provider"`
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Location string `json:"location"`
	State    int    `json:"state"`
}
