package customregexp

import (
	"fmt"
	"regexp"
)

func FilterBySubexpNames(pattern string, s string, content any) any {
	contentList, ok := content.([]any)
	if !ok {
		return nil
	}

	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(s)
	if match == nil {
		return nil
	}

	if re.NumSubexp() == 0 {
		return nil
	}

	regexpGroupNames := make(map[string]string)
	for i, name := range re.SubexpNames() {
		if i == 0 {
			continue
		}
		regexpGroupNames[name] = match[i]
	}

	for _, content := range contentList {
		c, ok := content.(map[string]any)
		if !ok {
			break
		}

		var matches int
		for k, v := range regexpGroupNames {
			contentPropertyValue, ok := c[k]
			if !ok {
				continue
			}

			strContentPropertyValue := fmt.Sprintf("%v", contentPropertyValue)
			strPossibleValue := fmt.Sprintf("%v", v)
			if strContentPropertyValue == strPossibleValue {
				matches++
			}
		}

		if matches == len(regexpGroupNames) {
			return c
		}
	}

	return nil
}
