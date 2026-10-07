; DQ3 Issue #5 matching probe; source is semantic assembly, with no db bytes.
; Input: assets_raw/DQ3.EXE, 115282 bytes.
; SHA256: 5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c.
; IDA Pro 9.4: sub_1E6B9, linear 1E6B9..1E6C9.
; Logical E6B9..E6C9; file FA29..FA39; DS-relative word 0B5A.
; Instruction identity: confirmed by IDA bytes and original file comparison.
; DS=DGROUP at the NPC caller is separately scoped by docs/188.
; No claims about an original source language, compiler, or campaign RNG order.

BITS 16
seed_word equ 0x0b5a

sub_1E6B9:
    mov ax, [seed_word]
    add ax, 0x9018
    rol ax, 1
    rol ax, 1
    rol ax, 1
    mov [seed_word], ax
    ret
