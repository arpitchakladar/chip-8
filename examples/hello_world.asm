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
