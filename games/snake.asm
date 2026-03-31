START:
	CALL INITIALIZE
LOOP:
	CALL DRAW_SNAKE
	JP LOOP

INITIALIZE:
	LD V0, 32         ; Head X
	LD V1, 16         ; Head Y
	LD V2, 1          ; Vel X
	LD V3, 0          ; Vel Y
	LD V4, 4          ; Length

	; Save these initial values to our RAM labels
	LD I, SNAKE_HEAD_X
	LD [I], V4        ; Stores V0, V1, V2, V3, and V4 into RAM
	RET

DRAW_SNAKE:
	; 1. Load Head position from RAM back into registers
	LD I, SNAKE_HEAD_X
	LD V4, [I]        ; Fills V0 through V4 with the saved data

	; 2. Draw the head using V0 (X) and V1 (Y)
	LD I, SPRITE_DOT
	DRW V0, V1, 1

	; 3. Draw the body (The Loop)
	; You would point I to SNAKE_BODY_DATA and loop V4 times
	RET

; --- Data Section ---
; Aligning labels so they don't overlap
SPRITE_DOT:
	DB 0x80

; We use enough space to store the snake's state
; LD [I], V4 needs 5 bytes of space (V0, V1, V2, V3, V4)
SNAKE_HEAD_X:	  ; Stores Head X (V0)
	DB 0x00
SNAKE_HEAD_Y:	; Stores Head Y (V1)
	DB 0x00
SNAKE_VEL_X:	 ; Stores Vel X (V2)
	DB 0x00
SNAKE_VEL_Y:	 ; Stores Vel Y (V3)
	DB 0x00
SNAKE_LEN:	   ; Stores Length (V4)
	DB 0x00

SNAKE_BODY_DATA:
; Each snake body has two bytes for X and Y coordinates
