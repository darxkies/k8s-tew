package config

type Features []string

func (features Features) HasFeatures(otherFeatures Features) bool {
	for _, feature := range features {
		for _, otherFeature := range otherFeatures {
			if feature == otherFeature {
				return true
			}
		}
	}

	return false
}

func CompareFeatures(source, destination Features) bool {
	return source != nil && destination != nil && source.HasFeatures(destination)
}
