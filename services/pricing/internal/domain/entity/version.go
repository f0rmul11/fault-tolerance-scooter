package entity

type version struct {
	Old int
	new *int
}

func (v *version) inc() {
	next := v.Old + 1
	v.new = &next
}
func (e *Trip) OldVersion() int {
	return e.version.Old
}
func (e *Trip) Version() int {
	if e.version.new != nil {
		return *e.version.new
	}
	return e.version.Old
}
