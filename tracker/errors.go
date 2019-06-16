package tracker

type clientError struct {
	message string
}

func (c *clientError) Error() string {
	return c.message
}
