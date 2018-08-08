package jasper

func sliceContains(group []string, name string) bool {
	for _, g := range group {
		if name == g {
			return true
		}
	}

	return false
}
