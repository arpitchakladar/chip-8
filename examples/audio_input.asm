; --- CHIP-8 Audio & Input Test ---
; Press any key to hear a 0.5-second beep

START:
	; 1. Wait for a key press (Opcode FX0A)
	; This halts the CPU until a key is pressed, then stores key in V0
	LD V0, K

	; 2. Set the Sound Timer (ST) to 15
	; 15 / 60 = 0.25 seconds of beep
	LD ST, V0        ; First, put a value in a register (V0 is already loaded)
	LD V1, 15        ; 15 Frames of beep
	LD ST, V1        ; This triggers the "Beep" in your emulator

	; 3. Visual feedback
	; Draw a small block so we know the CPU is alive
	LD V2, 30        ; X
	LD V3, 15        ; Y
	LD I, BLOCK
	DRW V2, V3, 5

BEEP_WAIT:
	; 4. Check if sound timer is still running
	; We stay in this loop until the timer hits 0
	LD V4, ST
	SE V4, 0x00      ; Skip if ST == 0
	JP BEEP_WAIT     ; Otherwise, keep waiting

	; 5. Clear screen and repeat
	CLS
	JP START

; --- Sprite Data ---
BLOCK:
	DB 0xFF ; ########
	DB 0xFF ; ########
	DB 0xFF ; ########
	DB 0xFF ; ########
	DB 0xFF ; ########
