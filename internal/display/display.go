package display

const (
	Width  = 64
	Height = 32
)

// Display handles the 64x32 monochrome pixel buffer.
type Display struct {
	// Pixels represents the screen state (0 = off, 1 = on).
	Pixels [Width * Height]byte
}

// New creates a new, cleared display instance.
func New() *Display {
	return new(Display)
}

// Reset clears the entire pixel buffer to black (0).
func (d *Display) Reset() {
	d.Pixels = [Width * Height]byte{}
}

// Clear is an alias for Reset, often called by the 00E0 opcode.
func (d *Display) Clear() {
	d.Reset()
}

// SetPixel toggles a pixel at (x, y) and returns true if a collision occurred.
// Chip-8 uses XOR drawing: if a pixel is already on and we draw it again, it turns off.
func (d *Display) SetPixel(x, y uint8) bool {
	// Wrap coordinates (standard Chip-8 behavior)
	x %= Width
	y %= Height

	index := uint16(x) + (uint16(y) * Width)

	// Check for collision (pixel was 1, now will be 0)
	collision := d.Pixels[index] == 1

	// XOR the pixel
	d.Pixels[index] ^= 1

	return collision
}
