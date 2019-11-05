package main

type SyncableDstStdout struct {
	fn func()
}

func (s *SyncableDstStdout) SetIBUFKicker(fn func()) {
	s.fn = fn
}
func (s *SyncableDstStdout) GetInput() (ibuf, ibufFlag uint8) {
	return 0, 0
}
func (s *SyncableDstStdout) SetInput(iv uint8) {
	//fmt.Printf("%c", byte(iv))
	s.fn()
}
