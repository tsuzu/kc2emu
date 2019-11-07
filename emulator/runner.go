package emulator

type Runner struct {
	phase     Phase
	nextPhase int

	*Storage
}

func NewRunner(executable []uint8) *Runner {
	storage := NewStorage()

	copy(storage.Memory, executable)

	return &Runner{
		Storage: storage,
	}
}

func (r *Runner) Run() {
	for !r.RunSingleInstruction() {
	}
}

func (r *Runner) RunSingleInstruction() (halt bool) {
	for {
		halt, nextPhase := r.RunSinglePhase()

		if halt {
			return true
		}

		if nextPhase == 0 {
			break
		}
	}

	return false
}

func (r *Runner) RunSinglePhase() (halt bool, nextPhase int) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.halt {
		return true, 0
	}

	if r.phase == nil {
		r.phase = p0
	}

	r.phase = r.phase(r.Storage)

	r.nextPhase++
	if r.phase == nil {
		r.nextPhase = 0
	}

	if r.halt {
		return true, 0
	}

	return false, r.nextPhase
}
