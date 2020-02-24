package config

type Labels []string

func (labels Labels) HasLabels(otherLabels Labels) bool {
	for _, label := range labels {
		for _, otherLabel := range otherLabels {
			if label == otherLabel {
				return true
			}
		}
	}

	return false
}

func CompareLabels(source, destination Labels) bool {
	return source != nil && destination != nil && source.HasLabels(destination)
}
