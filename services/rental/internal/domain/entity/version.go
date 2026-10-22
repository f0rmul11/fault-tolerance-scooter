package entity

type version struct {
	Old int
	new *int
}

func (v *version) inc() {
	next := v.Old + 1
	if v.new != nil {
		next = *v.new + 1
	}
	v.new = &next
}
