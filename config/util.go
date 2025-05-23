package config

func UseOrDefault(a any, b any) any {
	if a != nil {
		return a
	}
	return b
}
