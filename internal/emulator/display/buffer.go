package display

const (
	Width  = 64
	Height = 32
)

type DisplayBuffer struct {
	Pixels [Width * Height]byte
}

func NewDisplayBuffer() *DisplayBuffer {
	return &DisplayBuffer{}
}

func (b *DisplayBuffer) Clear() {
	b.Pixels = [Width * Height]byte{}
}

func (b *DisplayBuffer) SetPixel(x, y uint8) (bool, error) {
	x %= Width
	y %= Height

	index := uint16(x) + (uint16(y) * Width)
	collision := b.Pixels[index] == 1
	b.Pixels[index] ^= 1

	return collision, nil
}

func (b *DisplayBuffer) GetPixels() []byte {
	return b.Pixels[:]
}
