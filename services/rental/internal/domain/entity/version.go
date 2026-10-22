package entity

type version struct {
	Old int
	new *int
}

func (v *version) inc() {
	next := v.Old + 1
	v.new = &next
}
func (e *Rental) OldVersion() int {
	return e.version.Old
}
func (e *Rental) Version() int {
	if e.version.new != nil {
		return *e.version.new
	}
	return e.version.Old
}
