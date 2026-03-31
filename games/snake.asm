START:
	CALL INITIALIZE
	CALL GENERATE_AND_DRAW_FOOD
	CALL DRAW_SNAKE
	LOOP:
		LD V0, 10        ; Delay for ~166ms (10/60 seconds)
		LD DT, V0

	WAIT_LOOP:
		CALL CHECK_INPUT
		LD V0, DT
		SNE V0, 0        ; Wait for timer to hit zero
		JP TRIGGER_MOVE
		JP WAIT_LOOP

	TRIGGER_MOVE:
		CALL MOVE_SNAKE
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

REMOVE_OLD_FOOD:
	LD I, FOOD_X
	LD V1, [I]    ; Read coordinates from memory
	LD I, SPRITE_DOT
	DRW V0, V1, 1 ; Remove food
	RET

GENERATE_AND_DRAW_FOOD:
	RND V0, 0x3F  ; Generate random X coordinate
	RND V1, 0x1F  ; Generate random Y coordinate
	LD I, FOOD_X
	LD [I], V1    ; Save coordinates to memory
	LD I, SPRITE_DOT
	DRW V0, V1, 1 ; Draw food
	RET

CHECK_INPUT:
	; --- Check Key 2 (UP) ---
	LD V0, 0x02
	SKNP V0            ; If Key 2 is pressed, don't skip
	CALL SET_UP

	; --- Check Key 8 (DOWN) ---
	LD V0, 0x08
	SKNP V0
	CALL SET_DOWN

	; --- Check Key 4 (LEFT) ---
	LD V0, 0x04
	SKNP V0
	CALL SET_LEFT

	; --- Check Key 6 (RIGHT) ---
	LD V0, 0x06
	SKNP V0
	CALL SET_RIGHT
	RET

; --- Direction Setters ---
; We update the VelX (V0) and VelY (V1) and save them to RAM
SET_UP:
	LD V0, 0          ; VelX = 0
	LD V1, 0xFF       ; VelY = -1 (255 in 8-bit unsigned)
	LD I, SNAKE_VEL_X
	LD [I], V1        ; Save V0 and V1
	RET

SET_DOWN:
	LD V0, 0          ; VelX = 0
	LD V1, 1          ; VelY = 1
	LD I, SNAKE_VEL_X
	LD [I], V1
	RET

SET_LEFT:
	LD V0, 0xFF       ; VelX = -1
	LD V1, 0          ; VelY = 0
	LD I, SNAKE_VEL_X
	LD [I], V1
	RET

SET_RIGHT:
	LD V0, 1          ; VelX = 1
	LD V1, 0          ; VelY = 0
	LD I, SNAKE_VEL_X
	LD [I], V1
	RET

MOVE_SNAKE:
	LD I, SNAKE_LEN
	LD V0, [I]
	LD V3, V0 ; Get snake length
	ADD V3, V0 ; Get snake length (*2 for the fact each body is 2 bytes)
	LD V4, 2 ; for decrementing counter

	LD I, FOOD_X
	LD V1, [I]

	LD V5, V0
	LD V6, V1

	LD I, SNAKE_BODY_DATA
	LD V1, [I]

	SE V5, V0
	JP REMOVE_TAIL
	SNE V6, V1 ; Collision of head with food
	JP PRESERVE_TAIL
	JP REMOVE_TAIL

	PRESERVE_TAIL: ; Make the snake grow
		ADD V3, V4
		LD I, SNAKE_LEN
		LD V0, [I]
		ADD V0, 1
		LD [I], V0
		LD V0, 20
		LD ST, V0
		CALL REMOVE_OLD_FOOD
		CALL GENERATE_AND_DRAW_FOOD
		JP START_MAKE_SNAKE_LOOP

	REMOVE_TAIL:
		SUB V3, V4 ; Get the index of last body part
		LD I, SNAKE_BODY_DATA
		LD I, V3
		LD V1, [I]
		LD I, SPRITE_DOT
		DRW V0, V1, 1 ; Reset the tail of the sna
		ADD V3, V4
		JP START_MAKE_SNAKE_LOOP ; make (to make it move forward)

	START_MAKE_SNAKE_LOOP:
		MOVE_SNAKE_LOOP:
			SUB V3, V4
			LD I, SNAKE_BODY_DATA
			LD I, V3
			LD V1, [I]
			ADD V3, V4
			LD I, SNAKE_BODY_DATA
			LD I, V3
			LD [I], V1
			SUB V3, V4

			SE V3, 0
			JP MOVE_SNAKE_LOOP

	LD I, SNAKE_VEL_X
	LD V1, [I]
	LD V2, V0
	LD v3, V1
	LD I, SNAKE_BODY_DATA
	LD V1, [I]
	ADD V0, V2
	ADD V1, V3
	LD [I], V1
	LD I, SPRITE_DOT
	DRW V0, V1, 1 ; Create the new head
	RET

DRAW_SNAKE:
	; 1. Load snake information from RAM back into registers
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
