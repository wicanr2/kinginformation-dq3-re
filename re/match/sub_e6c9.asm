; DQ3 Issue #5: bounded RNG, instruction-source matching prototype.
; Input: assets_raw/DQ3.EXE, 115282 bytes.
; SHA256: 5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c.
; IDA Pro 9.4: sub_1E6C9, linear 1E6C9..1E6E7.
; Logical E6C9..E6E7; file FA39..FA57; DS-relative word 0B5A.
; Instruction identity: verified from IDA and original file; ABI probe in docs/25.
; BX supplies the divisor. DX receives the remainder, AX the quotient.
; BX=0 returns DX=0 without advancing the DS-relative state word.
; Names below are navigation aliases; no original source-language claim.

BITS 16
state_word equ 0x0b5a

sub_1E6C9:
    cmp bx, byte 0
    jnz short .nonzero
    mov dx, 0
    ret
.nonzero:
    mov ax, [state_word]
    add ax, 0x9018
    rol ax, 1
    rol ax, 1
    rol ax, 1
    mov [state_word], ax
    mov dx, 0
    div bx
    ret
