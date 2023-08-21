package tcp

var closedChan chan struct{}

func init() {
	closedChan = make(chan struct{})
	close(closedChan)
}
