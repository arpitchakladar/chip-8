START:
	CALL INITIALIZE
	CALL GENERATE_AND_DRAW_FOOD
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
	LD V5, 1
	INITIALIZE_SNAKE_BODY_LOOP:
		SUB V0, V5 ; Decrement the X position by 1 (for the next body part)
		LD [I], V1
		ADD V3, 1
		LD I, V4
		SE V3, V2 ; Add as many bodies as length of snake
		JP INITIALIZE_SNAKE_BODY_LOOP
	RET

GENERATE_AND_DRAW_FOOD:
	RND V0, 0xFF
	RND V1, 0xFF
	LD I, FOOD_X
	LD [I], V1
	LD I, SPRITE_DOT
	DRW V0, V1, 1
	RET

DRAW_SNAKE:
	; 1. Load Head position from RAM back into registers
	LD I, SNAKE_LEN
	LD V0, [I]
	LD V2, V0
	LD V3, 0
	LD V4, 2
	DRAW_SNAKE_BODY_LOOP:
		LD I, SNAKE_BODY_DATA
		LD I, V3
		LD I, V3          ; Add V3 twice because the equation is ith body = I + V3 * 2
		LD V1, [I]        ; Fills V0 through V4 with the saved data

		; 2. Draw the head using V0 (X) and V1 (Y)
		LD I, SPRITE_DOT
		DRW V0, V1, 1

		ADD V3, 1
		SE V3, V2
		JP DRAW_SNAKE_BODY_LOOP

	; 3. Draw the body (The Loop)
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

FOOD_X:
	DB 0x00
FOOD_Y:
	DB 0x00

SNAKE_BODY_DATA:
; Each snake body has four bytes for X and Y coordinates
; This doesn't include the snake head
