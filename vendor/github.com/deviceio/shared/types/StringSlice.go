package types

// StringSlice ...
type StringSlice []string

// Contains ...
func (t StringSlice) Contains(strings ...string) bool {
	if len(strings) == 1 {
		for _, s := range t {
			if s == strings[0] {
				return true
			}
		}
	} else if len(strings) > 1 {
		for _, s := range strings {
			if !t.Contains(s) {
				return false
			}
		}
		return true
	}

	return false
}

// ToSlice ...
func (t StringSlice) ToSlice() []string {
	return t
}
