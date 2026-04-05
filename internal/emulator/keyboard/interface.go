package keyboard

type Keyboard interface {
	IsKeyPressed(key byte) bool
	AnyKeyPressed() (byte, bool)
	SetKey(key byte, pressed bool)
}
