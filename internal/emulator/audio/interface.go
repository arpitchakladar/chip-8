package audio

type Audio interface {
	Init() error
	GenerateBeep() error
	Play()
	Pause()
	Close()
}
