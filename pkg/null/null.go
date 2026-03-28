package null

func FromString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func FromInt(i int) *int {
	if i == 0 {
		return nil
	}
	return &i
}

func FromFloat(f float64) *float64 {
	if f == 0 {
		return nil
	}
	return &f
}

func FromBool(b bool) *bool {
	if !b {
		return nil
	}
	return &b
}
