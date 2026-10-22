package entity

type version struct {
	Old      int
	new      *int
	expected *int
}

func (v *version) inc() {
	next := v.Old + 1
	v.new = &next
}
