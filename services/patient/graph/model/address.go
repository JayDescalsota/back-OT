package model

type Address struct {
	ID string `json:"id"`
}

func (Address) IsEntity() {}

type PatientAddress struct {
	ID        string `json:"id"`
	PatientID string `json:"patientId"`
	Address   string `json:"address"`
	Baranggay string `json:"baranggay"`
	City      string `json:"city"`
	State     string `json:"state"`
	ZipCode   string `json:"zipCode"`
	Country   string `json:"country"`
}
