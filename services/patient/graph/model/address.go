package model

type Address struct {
	ID string `json:"id"`
}

func (Address) IsEntity() {}
