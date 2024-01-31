package entity

type Record struct {
	ID   int               `json:"id"`
	Data map[string]string `json:"data"`
	Accumulated map[string]string `json:"accumulated"`
	Version int `json:"version"`
}

func (d *Record) Copy() Record {
	values := d.Data

	newMap := map[string]string{}
	for key, value := range values {
		newMap[key] = value
	}

	return Record{
		ID:   d.ID,
		Data: newMap,
	}
}
