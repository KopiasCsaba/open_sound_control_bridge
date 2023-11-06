package slicetools

func Map[T any, U any](slice []T, cb func(T) U) []U {
	r := make([]U, len(slice))
	for i, v := range slice {
		r[i] = cb(v)
	}
	return r
}

func FindOne[T any](slice []T, cb func(T) bool) *T {
	for _, v := range slice {
		if cb(v) {
			return &v
		}
	}
	return nil
}

func Filter[T any](slice []T, cb func(T) bool) []T {
	r := []T{}
	for _, v := range slice {
		if cb(v) {
			r = append(r, v)
		}
	}
	return r
}

func FilterOnType[T any, Q any](slice []T, cb func(T) (Q, bool)) []Q {
	r := []Q{}
	var now Q
	var ok bool
	for _, v := range slice {
		if now, ok = cb(v); ok {
			r = append(r, now)
		}
	}
	return r
}

func Reduce[T, U any](slice []T, init U, cb func(U, T) U) U {
	r := init
	for _, v := range slice {
		r = cb(r, v)
	}
	return r
}

func IndexOf[T comparable](slice []T, needle T) int {
	for index, v := range slice {
		if v == needle {
			return index
		}
	}
	return -1
}

func Contains[T comparable](slice []T, element T) bool {
	return IndexOf(slice, element) != -1
}

func ContainsTheSameElements[T comparable](a []T, b []T) bool {
	if len(a) != len(b) {
		return false
	}

outer:
	for ai := range a {
		for bi := range b {
			if a[ai] == b[bi] {
				continue outer
			}
		}
		return false
	}

	return true
}

func Diff[T comparable](a []T, b []T) []T {
	diff := []T{}
	var found bool
	for _, aElement := range a {
		found = false
		for _, bElement := range b {
			if aElement == bElement {
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, aElement)
		}
	}

	return diff
}

func SmallestItem[T int | int64 | int32 | float32 | float64](slice []T) (int, *T) {
	if len(slice) == 0 {
		return -1, nil
	}

	smallest := slice[0]
	index := 0
	for i, element := range slice {
		if element < smallest {
			smallest = element
			index = i
		}
	}

	return index, &smallest
}

func LargestItem[T int | int64 | int32 | float32 | float64](slice []T) (int, *T) {
	if len(slice) == 0 {
		return -1, nil
	}

	smallest := slice[0]
	index := 0
	for i, element := range slice {
		if element > smallest {
			smallest = element
			index = i
		}
	}

	return index, &smallest
}
