package main

func Map(data []string, mutator func(string) string) (result []string) {
	result = make([]string, len(data))
	for index, value := range data {
		result[index] = mutator(value)
	}
	return
}

func Filter(data []string, predicate func(string) bool) (result []string) {
	tmp := make([]string, 0, len(data))
	for _, value := range data {
		if predicate(value) {
			tmp = append(tmp, value)
		}
	}
	return append([]string(nil), tmp[:]...)
}

func Reduce(data []string, combinator func(string, string) string) (result string) {
	if len(data) == 0 {
		return ""
	}

	result = data[0]
	for _, value := range data[1:] {
		result = combinator(result, value)
	}
	return
}

func Any(data []string, predicate func(string) bool) bool {
	for _, value := range data {
		if predicate(value) {
			return true
		}
	}
	return false
}

func All(data []string, predicate func(string) bool) bool {
	for _, value := range data {
		if !predicate(value) {
			return false
		}
	}
	return true
}
