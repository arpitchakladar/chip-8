; --- Hello World (Draw a Pixel) ---
; This program sets Index register I to a sprite
; and draws it at coordinates (V0, V1)

START:
    LD V0, 0x10      ; X coordinate = 16
    LD V1, 0x10      ; Y coordinate = 16
    LD I, SPRITE     ; Point I to the label 'SPRITE'
    DRW V0, V1, 5    ; Draw 5 bytes from memory at (V0, V1)

LOOP:
    JP LOOP          ; Infinite loop to keep the screen visible

; --- Data Section ---
; A simple 'H' shape sprite (5 bytes high)
SPRITE:
    DB 0x82
    DB 0x82
    DB 0xFE
    DB 0x82
    DB 0x82
