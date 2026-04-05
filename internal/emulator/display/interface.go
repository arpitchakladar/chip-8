package display

type Display interface {
	Init() error
	Clear()
	SetPixel(x, y uint8) (bool, error)
	Present() error
	Reset()
	Close() error
}
