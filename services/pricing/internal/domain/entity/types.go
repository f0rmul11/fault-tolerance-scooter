package entity

type Status string

const (
	Reserved  Status = "RESERVED"
	Active    Status = "ACTIVE"
	Completed Status = "COMPLETED"
	Canceled  Status = "CANCELED"
)
