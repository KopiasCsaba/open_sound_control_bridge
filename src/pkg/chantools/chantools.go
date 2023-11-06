package chantools

func ChanIsOpenReader[T comparable](c <-chan T) bool {
	select {
	case _, ok := <-c:
		if !ok {
			return false
		}
		return true

	default:
		return true
	}
}
