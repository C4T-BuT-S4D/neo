package neoproc

// SetSubreaper is not available on darwin.
// Live with it.
func SetSubreaper() error {
	return nil
}
