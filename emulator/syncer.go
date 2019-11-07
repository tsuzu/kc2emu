package emulator

type SyncableSrc interface {
	SetOBUFKicker(fn func())
	GetOutput() (obuf, obufFlag uint8)
	ClearOBUFFlag()
}

type SyncableDst interface {
	SetIBUFKicker(fn func())
	GetInput() (ibuf, ibufFlag uint8)
	SetInput(iv uint8)
}

type Syncer struct {
	from SyncableSrc
	to   SyncableDst
}

func NewSyncer(from SyncableSrc, to SyncableDst) {
	s := Syncer{
		from: from,
		to:   to,
	}

	from.SetOBUFKicker(s.obufKicker)
	to.SetIBUFKicker(s.ibufKicker)
}

func (s *Syncer) obufKicker() {
	obuf, obufFlag := s.from.GetOutput()

	if obufFlag != 0 {
		s.to.SetInput(obuf)
	}
}

func (s *Syncer) ibufKicker() {
	_, ibufFlag := s.to.GetInput()

	if ibufFlag == 0 {
		s.from.ClearOBUFFlag()
	}
}
