package tmp

type poolChannel[T any] struct {
	channel    chan *T
	bufferSize int
}

func NewPoolChannel[T any](chanSize int, bufferSize int) poolChannel[T] {
	return poolChannel[T]{
		channel:    make(chan *T, chanSize),
		bufferSize: bufferSize,
	}
}

func (pc *poolChannel[T]) buffer() {
	for i := 0; i < pc.bufferSize; i++ {
		pc.channel <- new(T)
	}
}

func (pc *poolChannel[T]) Get() *T {
	// if channel is empty buffer channel
	if len(pc.channel) == 0 {
		pc.buffer()
	}

	return <-pc.channel
}

func (pc *poolChannel[T]) Put(data *T) {
	// if channel is full drop data
	if len(pc.channel) == cap(pc.channel) {
		return
	}

	pc.channel <- data
}
