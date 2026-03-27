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
    DW 0x8888 ; 10001000
    DW 0xF888 ; 10001000
    DW 0x8800 ; 10001000

CHAR_E:
    DW 0xF880 ; 11111000
    DW 0xF080 ; 11110000
    DW 0xF800 ; 11111000

CHAR_L:
    DW 0x8080 ; 10000000
    DW 0x8080 ; 10000000
    DW 0xF800 ; 11111000

CHAR_O:
    DW 0x7088 ; 10001000
    DW 0x8888 ; 10001000
    DW 0x7000 ; 01110000
