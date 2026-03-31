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
		ADD I, V4
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
	; --- Check Key 2 (UP, key 2) ---
	LD V0, 0x02
	SKNP V0            ; If Key 2 is pressed, don't skip
	CALL SET_UP

	; --- Check Key 8 (DOWN, key s) ---
	LD V0, 0x08
	SKNP V0
	CALL SET_DOWN

	; --- Check Key 4 (LEFT, key q) ---
	LD V0, 0x04
	SKNP V0
	CALL SET_LEFT

	; --- Check Key 6 (RIGHT, key e) ---
	LD V0, 0x06
	SKNP V0
	CALL SET_RIGHT
	RET

; --- Direction Setters ---
; We update the VelX (V0) and VelY (V1) and save them to RAM
SET_UP:
	LD  I, SNAKE_VEL_X
	LD  V1, [I]           ; Load current VelY into V2
	SNE  V1, 1             ; If VelY == 1 (moving DOWN), skip the RET
	RET                   ; Already moving down — ignore
	LD  V0, 0
	LD  V1, 0xFF
	LD  I, SNAKE_VEL_X
	LD  [I], V1
	RET

SET_DOWN:
	LD  I, SNAKE_VEL_X
	LD  V1, [I]
	SNE  V1, 0xFF          ; If VelY == 0xFF (moving UP), skip the RET
	RET                   ; Already moving up — ignore
	LD  V0, 0
	LD  V1, 1
	LD  I, SNAKE_VEL_X
	LD  [I], V1
	RET

SET_LEFT:
	LD  I, SNAKE_VEL_X
	LD  V1, [I]
	SNE  V0, 1             ; If VelX == 1 (moving RIGHT), skip the RET
	RET                   ; Already moving right — ignore
	LD  V0, 0xFF
	LD  V1, 0
	LD  I, SNAKE_VEL_X
	LD  [I], V1
	RET

SET_RIGHT:
	LD  I, SNAKE_VEL_X
	LD  V1, [I]
	SNE  V0, 0xFF          ; If VelX == 0xFF (moving LEFT), skip the RET
	RET                   ; Already moving left — ignore
	LD  V0, 1
	LD  V1, 0
	LD  I, SNAKE_VEL_X
	LD  [I], V1
	RET

MOVE_SNAKE:
	LD I, SNAKE_LEN
	LD V0, [I]
	LD V3, V0 ; Get snake length
	ADD V3, V0 ; Get snake length (*2 for the fact each body is 2 bytes)
	LD V4, 2 ; for decrementing counter

	CHECK_COLLISION_WITH_BODY:
		; Load position of the snake head
		LD I, SNAKE_BODY_DATA
		LD V1, [I]

		LD V5, V0
		LD V6, V1

		LD V7, V3
		CHECK_COLLISION_WITH_BODY_LOOP:
			SUB V7, V4

			; Load coordinates of snake body piece
			LD I, SNAKE_BODY_DATA
			ADD I, V7

			LD V1, [I]

			; Check if the corrdinates are same
			SE V0, V5
			JP CHECK_COLLISION_WITH_BODY_LOOP_continue
			SE V1, V6
			JP CHECK_COLLISION_WITH_BODY_LOOP_continue

			JP STOP_GAME

			CHECK_COLLISION_WITH_BODY_LOOP_continue:
			SE V7, 2
			JP CHECK_COLLISION_WITH_BODY_LOOP

	CHECK_COLLISION_WITH_FOOD:
		; Load position of food
		LD I, FOOD_X
		LD V1, [I]

		LD V5, V0
		LD V6, V1

		; Load position of snake head
		LD I, SNAKE_BODY_DATA
		LD V1, [I]

		SE V5, V0
		JP REMOVE_TAIL
		SNE V6, V1 ; Collision of head with food
		JP PRESERVE_TAIL
		JP REMOVE_TAIL

		PRESERVE_TAIL: ; Make the snake grow
			; Prevent the tail from getting deleted
			ADD V3, V4

			; Increase length of snake
			LD I, SNAKE_LEN
			LD V0, [I]
			ADD V0, 1
			LD [I], V0

			; Play sound
			CALL PLAY_BEEP

			; Draw new food
			CALL REMOVE_OLD_FOOD
			CALL GENERATE_AND_DRAW_FOOD
			JP START_MAKE_SNAKE_LOOP

		REMOVE_TAIL:
			SUB V3, V4 ; Get the index of last body part

			; Draw over the last tail to remove it
			LD I, SNAKE_BODY_DATA
			ADD I, V3
			LD V1, [I]
			LD I, SPRITE_DOT
			DRW V0, V1, 1 ; Reset the tail of the sna

			; Reset the registors
			ADD V3, V4
			JP START_MAKE_SNAKE_LOOP ; make (to make it move forward)


	START_MAKE_SNAKE_LOOP:
		MOVE_SNAKE_LOOP:
			; Decrement the index counter for snake body to get the
			; body just ahead of this one
			; V3 gets subtracted by 2 (each snake body is 2 bytes)
			SUB V3, V4

			; Loading the position of the current snake body
			LD I, SNAKE_BODY_DATA
			ADD I, V3
			LD V1, [I]

			; Getting the position of the current snake body piece
			; Add 2 for reseting the changes used to get previous body
			ADD V3, V4
			LD I, SNAKE_BODY_DATA
			ADD I, V3
			LD [I], V1

			; Finally decrementing the counter
			SUB V3, V4

			SE V3, 0
			JP MOVE_SNAKE_LOOP

	LD I, SNAKE_VEL_X
	LD V1, [I]
	LD V2, V0
	LD V3, V1
	LD I, SNAKE_BODY_DATA
	LD V1, [I]
	ADD V0, V2
	ADD V1, V3

	LD V4, 0x3F   ; Mask for 63 (Width - 1)
	AND V0, V4    ; If V0 was 64, it now becomes 0

	LD V4, 0x1F   ; Mask for 31 (Height - 1)
	AND V1, V4    ; If V1 was 32, it now becomes 0

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
		ADD I, V3
		ADD I, V3         ; Add V3 twice because the equation is ith body = I + V3 * 2
		LD V1, [I]        ; Fills V0 through V4 with the saved data

		; 2. Draw the head using V0 (X) and V1 (Y)
		LD I, SPRITE_DOT
		DRW V0, V1, 1

		ADD V3, 1
		SE V3, V2
		JP DRAW_SNAKE_BODY_LOOP

	; 3. Draw the body (The Loop)
	RET

PLAY_BEEP:
	LD V0, 20
	LD ST, V0
	RET

STOP_GAME:
	CLS                  ; Clear the screen
	CALL PLAY_BEEP

	; ── Row 1: "GAME" at y=8 ──────────────────────────────────
	LD  V1, 8

	LD  V0, 20
	LD  I, SPR_GO_G
	DRW V0, V1, 8

	LD  V0, 26
	LD  I, SPR_GO_A
	DRW V0, V1, 8

	LD  V0, 32
	LD  I, SPR_GO_M
	DRW V0, V1, 8

	LD  V0, 38
	LD  I, SPR_GO_E
	DRW V0, V1, 8

	; ── Row 2: "OVER" at y=20 ─────────────────────────────────
	LD  V1, 20

	LD  V0, 20
	LD  I, SPR_GO_O
	DRW V0, V1, 8

	LD  V0, 26
	LD  I, SPR_GO_V
	DRW V0, V1, 8

	LD  V0, 32
	LD  I, SPR_GO_E
	DRW V0, V1, 8

	LD  V0, 38
	LD  I, SPR_GO_R
	DRW V0, V1, 8

	STOP_GAME_HALT:
		JP  STOP_GAME_HALT       ; Freeze here forever

	RET

; --- Data Section ---
; Aligning labels so they don't overlap
SPRITE_DOT:
	DB 0x80

; G
SPR_GO_G:
	DB 0x70
	DB 0x80
	DB 0x80
	DB 0xB8
	DB 0x88
	DB 0x78
	DB 0x00
	DB 0x00

; A
SPR_GO_A:
	DB 0x70
	DB 0x88
	DB 0x88
	DB 0xF8
	DB 0x88
	DB 0x88
	DB 0x00
	DB 0x00

; M
SPR_GO_M:
	DB 0x88
	DB 0xD8
	DB 0xA8
	DB 0x88
	DB 0x88
	DB 0x88
	DB 0x00
	DB 0x00

; E
SPR_GO_E:
	DB 0xF8
	DB 0x80
	DB 0x80
	DB 0xF0
	DB 0x80
	DB 0xF8
	DB 0x00
	DB 0x00

; O
SPR_GO_O:
	DB 0x70
	DB 0x88
	DB 0x88
	DB 0x88
	DB 0x88
	DB 0x70
	DB 0x00
	DB 0x00

; V
SPR_GO_V:
	DB 0x88
	DB 0x88
	DB 0x88
	DB 0x50
	DB 0x50
	DB 0x20
	DB 0x00
	DB 0x00

; R
SPR_GO_R:
	DB 0xF0
	DB 0x88
	DB 0x88
	DB 0xF0
	DB 0xA0
	DB 0x90
	DB 0x00
	DB 0x00

; We use enough space to store the snake's state
; LD [I], V4 needs 5 bytes of space (V0, V1, V2, V3, V4)
SNAKE_VEL_X:	 ; Stores Vel X (V2)
	DB 0x00
SNAKE_VEL_Y:	 ; Stores Vel Y (V3)
	DB 0x00
SNAKE_LEN:	     ; Stores Length (V4)
	DB 0x00

FOOD_X:
	DB 0x00
FOOD_Y:
	DB 0x00

SCORE_STORAGE:
	DB 0x00
	DB 0x00
	DB 0x00

SNAKE_BODY_DATA:
; Each snake body has 2 bytes for X and Y coordinates
; This doesn't include the snake head
