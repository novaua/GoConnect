package utils

// DublicateChanInt saves received packets id and resends
func DublicateChanInt(input <-chan int) (out1 chan int, out2 chan int) {
	out1 = make(chan int, 2)
	out2 = make(chan int)

	go func() {
		defer close(out1)
		defer close(out2)

		for data := range input {
			d1 := data
			d2 := data
			out1 <- d1
			out2 <- d2
		}
	}()

	return
}
