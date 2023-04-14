package utils

// Returns -1 if the searchTerm is not found
func BinarySearch[T int | string](a []T, searchTerm T) (location int) {
	mid := len(a) / 2

	switch {
	case len(a) == 0:
		location = -1 // not found
	case a[mid] > searchTerm:
		location = BinarySearch(a[:mid], searchTerm)
	case a[mid] < searchTerm:
		location = BinarySearch(a[mid+1:], searchTerm)
		if location >= 0 { // if anything but the -1 "not found" result
			location += mid + 1
		}
	default: // a[mid] == search
		location = mid // found
	}

	return
}
