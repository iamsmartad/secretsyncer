package helpers

import (
	"strings"
)

// func hasAnnotation(annotations map[string]string, key string, value string) (bool, bool) {
// 	if val, ok := annotations[key]; ok {
// 		if val == value {
// 			return true, true
// 		}
// 		return true, false
// 	}
// 	return false, false
// }

// FindMatchingAnnotation pass two maps of annotations and check if there are any matches between them
// values on individual keys are split by `,` to a foo=bar,baz would match foo=baz,too
func FindMatchingAnnotation(pool, criteria map[string]string) bool {
	splitPool := make(map[string][]string)
	splitCriteria := make(map[string][]string)
	// allow for multiple values in annotation i.e. `environment=dev,staging,production`
	for k, v := range pool {
		splitPool[k] = strings.Split(v, ",")
	}
	for k, v := range criteria {
		splitCriteria[k] = strings.Split(v, ",")
	}

	// iterate all candidates
	for key, value := range splitPool {
		// check for the annoation key in the criteria map
		if criteriaValue, ok := splitCriteria[key]; ok {
			for _, vpool := range value {
				for _, vcrit := range criteriaValue {
					if vpool == vcrit {
						return true
					}
				}
			}
		}
	}
	return false
}

// func filterOutNamspaces(secrets []v1.Secret, namespaces []string) []v1.Secret {
// 	var newSecrets []v1.Secret
// 	for _, v := range secrets {
// 		found := false
// 		for _, ns := range namespaces {
// 			if v.Namespace == ns {
// 				found = true
// 			}
// 		}
// 		if !found {
// 			newSecrets = append(newSecrets, v)
// 		}
// 	}

// 	return newSecrets
// }
