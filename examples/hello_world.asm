; --- CHIP-8 Hello World ---
; Displays "HELO" on the screen

START:
	LD V0, 0x02      ; X coordinate start
	LD V1, 0x0A      ; Y coordinate

	; Draw 'H'
	LD I, CHAR_H
	DRW V0, V1, 5
	ADD V0, 0x05     ; Move X for next character

	; Draw 'E'
	LD I, CHAR_E
	DRW V0, V1, 5
	ADD V0, 0x05

	; Draw 'L'
	LD I, CHAR_L
	DRW V0, V1, 5
	ADD V0, 0x05

	; Draw 'O'
	LD I, CHAR_O
	DRW V0, V1, 5

LOOP:
	JP LOOP          ; Halt here

; --- Character Sprites (5 bytes high) ---

CHAR_H:
	DB 0x88 ; 10001000
	DB 0x88 ; 10001000
	DB 0xF8 ; 11111000
	DB 0x88 ; 10001000
	DB 0x88 ; 10001000
	DB 0x00

CHAR_E:
	DB 0xF8 ; 11111000
	DB 0x80 ; 10000000
	DB 0xF0 ; 11110000
	DB 0x80 ; 10000000
	DB 0xF8 ; 11111000
	DB 0x00

CHAR_L:
	DB 0x80 ; 10000000
	DB 0x80 ; 10000000
	DB 0x80 ; 10000000
	DB 0x80 ; 10000000
	DB 0xF8 ; 11111000
	DB 0x00

CHAR_O:
	DB 0x70 ; 01110000
	DB 0x88 ; 10001000
	DB 0x88 ; 10001000
	DB 0x88 ; 10001000
	DB 0x70 ; 01110000
	DB 0x00
