START:
	CALL INITIALIZE
	CALL DRAW_SNAKE
LOOP:
	JP LOOP

INITIALIZE:
	LD V0, 1          ; Vel X
	LD V1, 0          ; Vel Y
	LD V2, 4          ; Length

	; Save these initial values to our RAM labels
	LD I, SNAKE_VEL_X
	LD [I], V2        ; Stores V0, V1, V2, V3, and V4 into RAM

	LD I, SNAKE_BODY_DATA
	LD V0, 32         ; Initial X axis of snake head
	LD V1, 16         ; Initial Y axis of snake head
	LD V3, 0          ; Loop counter
	LD V4, 2
	INITIALIZE_SNAKE_BODY_LOOP:
		LD [I], V1
		SUB V0, 1 ; Decrement the X position by 1 (for the next body part)
		ADD V3, 1
		SE V3, 2 ; Add as many bodies as length of snake
		LD I, V4
		JP INITIALIZE_SNAKE_BODY_LOOP
	RET

DRAW_SNAKE:
	; 1. Load Head position from RAM back into registers
	LD I, SNAKE_BODY_DATA
	LD V1, [I]        ; Fills V0 through V4 with the saved data

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
SNAKE_VEL_X:	 ; Stores Vel X (V2)
	DB 0x00
SNAKE_VEL_Y:	 ; Stores Vel Y (V3)
	DB 0x00
SNAKE_LEN:	   ; Stores Length (V4)
	DB 0x00

SNAKE_BODY_DATA:
	DB 0x32
	DB 0x32
	DB 0x32
	DB 0x32
	DB 0x32
	DB 0x32
	DB 0x32
	DB 0x32
	DB 0x32
	DB 0x32
; Each snake body has four bytes for X and Y coordinates
; This doesn't include the snake head
