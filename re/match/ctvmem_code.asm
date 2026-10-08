; Verified instruction source only. Original SDK data is not reconstructed.
; Input DQ3.EXE 115282 bytes SHA256 5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c
; CTVMEM.ASM original IDA linear 0x22f60..0x2391d
; BYTE/WORD qualifiers and absolute EQU symbols preserve assembler encoding.
.8086
EXTRN IMM_000F:ABS
EXTRN IMM_0001:ABS
EXTRN IMM_0004:ABS
EXTRN IMM_0006:ABS
EXTRN IMM_0008:ABS
EXTRN IMM_0021:ABS
EXTRN IMM_0026:ABS
EXTRN DISP_0000:ABS
EXTRN DISP_0001:ABS
EXTRN DISP_0003:ABS
_TEXT SEGMENT PARA PUBLIC USE16 'CODE'
ASSUME CS:_TEXT, DS:_TEXT
PUBLIC module_entry
module_entry LABEL NEAR
; Original operand: jmp     near ptr sub_236B4
L_0000: ; original 0x22f60; IDA sub_22F60; logical 0x12f60; file 0x142d0
    jmp NEAR PTR L_0754
; Unknown data skipped: 0x22f63..0x23043
ORG 0E3h
; Original operand: push    cx
L_00E3: ; original 0x23043; IDA sub_23043; logical 0x13043; file 0x143b3
    push cx
; Original operand: mov     cx, 200h
L_00E4: ; original 0x23044; IDA unnamed; logical 0x13044; file 0x143b4
    mov cx, 0200h
; Original operand: mov     ah, al
L_00E7: ; original 0x23047; IDA unnamed; logical 0x13047; file 0x143b7
    mov ah, al
; Original operand: in      al, dx
L_00E9: ; original 0x23049; IDA loc_23049; logical 0x13049; file 0x143b9
    in al, dx
; Original operand: or      al, al
L_00EA: ; original 0x2304a; IDA unnamed; logical 0x1304a; file 0x143ba
    or al, al
; Original operand: jns     short loc_23053
L_00EC: ; original 0x2304c; IDA unnamed; logical 0x1304c; file 0x143bc
    jns SHORT L_00F3
; Original operand: loop    loc_23049
L_00EE: ; original 0x2304e; IDA unnamed; logical 0x1304e; file 0x143be
    loop L_00E9
; Original operand: stc
L_00F0: ; original 0x23050; IDA unnamed; logical 0x13050; file 0x143c0
    stc 
; Original operand: jmp     short loc_23057
L_00F1: ; original 0x23051; IDA unnamed; logical 0x13051; file 0x143c1
    jmp SHORT L_00F7
; Original operand: mov     al, ah
L_00F3: ; original 0x23053; IDA loc_23053; logical 0x13053; file 0x143c3
    mov al, ah
; Original operand: out     dx, al
L_00F5: ; original 0x23055; IDA unnamed; logical 0x13055; file 0x143c5
    out dx, al
; Original operand: clc
L_00F6: ; original 0x23056; IDA unnamed; logical 0x13056; file 0x143c6
    clc 
; Original operand: pop     cx
L_00F7: ; original 0x23057; IDA loc_23057; logical 0x13057; file 0x143c7
    pop cx
; Original operand: retn
L_00F8: ; original 0x23058; IDA unnamed; logical 0x13058; file 0x143c8
    ret 
; Original operand: push    cx
L_00F9: ; original 0x23059; IDA sub_23059; logical 0x13059; file 0x143c9
    push cx
; Original operand: push    dx
L_00FA: ; original 0x2305a; IDA unnamed; logical 0x1305a; file 0x143ca
    push dx
; Original operand: mov     dx, ds:30h
L_00FB: ; original 0x2305b; IDA unnamed; logical 0x1305b; file 0x143cb
    mov dx, WORD PTR ds:[030h]
; Original operand: add     dl, 0Eh
L_00FF: ; original 0x2305f; IDA unnamed; logical 0x1305f; file 0x143cf
    add dl, 0Eh
; Original operand: mov     cx, 200h
L_0102: ; original 0x23062; IDA unnamed; logical 0x13062; file 0x143d2
    mov cx, 0200h
; Original operand: in      al, dx
L_0105: ; original 0x23065; IDA loc_23065; logical 0x13065; file 0x143d5
    in al, dx
; Original operand: or      al, al
L_0106: ; original 0x23066; IDA unnamed; logical 0x13066; file 0x143d6
    or al, al
; Original operand: js      short loc_2306F
L_0108: ; original 0x23068; IDA unnamed; logical 0x13068; file 0x143d8
    js SHORT L_010F
; Original operand: loop    loc_23065
L_010A: ; original 0x2306a; IDA unnamed; logical 0x1306a; file 0x143da
    loop L_0105
; Original operand: stc
L_010C: ; original 0x2306c; IDA unnamed; logical 0x1306c; file 0x143dc
    stc 
; Original operand: jmp     short loc_23074
L_010D: ; original 0x2306d; IDA unnamed; logical 0x1306d; file 0x143dd
    jmp SHORT L_0114
; Original operand: sub     dl, 4
L_010F: ; original 0x2306f; IDA loc_2306F; logical 0x1306f; file 0x143df
    sub dl, 04h
; Original operand: in      al, dx
L_0112: ; original 0x23072; IDA unnamed; logical 0x13072; file 0x143e2
    in al, dx
; Original operand: clc
L_0113: ; original 0x23073; IDA unnamed; logical 0x13073; file 0x143e3
    clc 
; Original operand: pop     dx
L_0114: ; original 0x23074; IDA loc_23074; logical 0x13074; file 0x143e4
    pop dx
; Original operand: pop     cx
L_0115: ; original 0x23075; IDA unnamed; logical 0x13075; file 0x143e5
    pop cx
; Original operand: retn
L_0116: ; original 0x23076; IDA unnamed; logical 0x13076; file 0x143e6
    ret 
; Original operand: mov     ah, al
L_0117: ; original 0x23077; IDA sub_23077; logical 0x13077; file 0x143e7
    mov ah, al
; Original operand: mov     al, 0F0h
L_0119: ; original 0x23079; IDA unnamed; logical 0x13079; file 0x143e9
    mov al, 0F0h
; Original operand: in      al, dx
L_011B: ; original 0x2307b; IDA loc_2307B; logical 0x1307b; file 0x143eb
    in al, dx
; Original operand: or      al, al
L_011C: ; original 0x2307c; IDA unnamed; logical 0x1307c; file 0x143ec
    or al, al
; Original operand: js      short loc_2307B
L_011E: ; original 0x2307e; IDA unnamed; logical 0x1307e; file 0x143ee
    js SHORT L_011B
; Original operand: mov     al, ah
L_0120: ; original 0x23080; IDA unnamed; logical 0x13080; file 0x143f0
    mov al, ah
; Original operand: out     dx, al
L_0122: ; original 0x23082; IDA unnamed; logical 0x13082; file 0x143f2
    out dx, al
; Original operand: retn
L_0123: ; original 0x23083; IDA unnamed; logical 0x13083; file 0x143f3
    ret 
; Original operand: push    dx
L_0124: ; original 0x23084; IDA sub_23084; logical 0x13084; file 0x143f4
    push dx
; Original operand: mov     dx, ds:30h
L_0125: ; original 0x23085; IDA unnamed; logical 0x13085; file 0x143f5
    mov dx, WORD PTR ds:[030h]
; Original operand: add     dl, 0Eh
L_0129: ; original 0x23089; IDA unnamed; logical 0x13089; file 0x143f9
    add dl, 0Eh
; Original operand: sub     al, al
L_012C: ; original 0x2308c; IDA unnamed; logical 0x1308c; file 0x143fc
    sub al, al
; Original operand: in      al, dx
L_012E: ; original 0x2308e; IDA loc_2308E; logical 0x1308e; file 0x143fe
    in al, dx
; Original operand: or      al, al
L_012F: ; original 0x2308f; IDA unnamed; logical 0x1308f; file 0x143ff
    or al, al
; Original operand: jns     short loc_2308E
L_0131: ; original 0x23091; IDA unnamed; logical 0x13091; file 0x14401
    jns SHORT L_012E
; Original operand: sub     dl, 4
L_0133: ; original 0x23093; IDA unnamed; logical 0x13093; file 0x14403
    sub dl, 04h
; Original operand: in      al, dx
L_0136: ; original 0x23096; IDA unnamed; logical 0x13096; file 0x14406
    in al, dx
; Original operand: pop     dx
L_0137: ; original 0x23097; IDA unnamed; logical 0x13097; file 0x14407
    pop dx
; Original operand: retn
L_0138: ; original 0x23098; IDA unnamed; logical 0x13098; file 0x14408
    ret 
; Original operand: mov     dx, ds:30h
L_0139: ; original 0x23099; IDA sub_23099; logical 0x13099; file 0x14409
    mov dx, WORD PTR ds:[030h]
; Original operand: add     dl, 6
L_013D: ; original 0x2309d; IDA unnamed; logical 0x1309d; file 0x1440d
    add dl, 06h
; Original operand: mov     al, 1
L_0140: ; original 0x230a0; IDA unnamed; logical 0x130a0; file 0x14410
    mov al, 01h
; Original operand: out     dx, al
L_0142: ; original 0x230a2; IDA unnamed; logical 0x130a2; file 0x14412
    out dx, al
; Original operand: in      al, dx
L_0143: ; original 0x230a3; IDA unnamed; logical 0x130a3; file 0x14413
    in al, dx
; Original operand: in      al, dx
L_0144: ; original 0x230a4; IDA unnamed; logical 0x130a4; file 0x14414
    in al, dx
; Original operand: in      al, dx
L_0145: ; original 0x230a5; IDA unnamed; logical 0x130a5; file 0x14415
    in al, dx
; Original operand: in      al, dx
L_0146: ; original 0x230a6; IDA unnamed; logical 0x130a6; file 0x14416
    in al, dx
; Original operand: sub     al, al
L_0147: ; original 0x230a7; IDA unnamed; logical 0x130a7; file 0x14417
    sub al, al
; Original operand: out     dx, al
L_0149: ; original 0x230a9; IDA unnamed; logical 0x130a9; file 0x14419
    out dx, al
; Original operand: mov     bl, 10h
L_014A: ; original 0x230aa; IDA unnamed; logical 0x130aa; file 0x1441a
    mov bl, 010h
; Original operand: call    sub_23059
L_014C: ; original 0x230ac; IDA loc_230AC; logical 0x130ac; file 0x1441c
    call L_00F9
; Original operand: cmp     al, 0AAh
L_014F: ; original 0x230af; IDA unnamed; logical 0x130af; file 0x1441f
    cmp al, 0AAh
; Original operand: jz      short loc_230BD
L_0151: ; original 0x230b1; IDA unnamed; logical 0x130b1; file 0x14421
    jz SHORT L_015D
; Original operand: dec     bl
L_0153: ; original 0x230b3; IDA unnamed; logical 0x130b3; file 0x14423
    dec bl
; Original operand: jnz     short loc_230AC
L_0155: ; original 0x230b5; IDA unnamed; logical 0x130b5; file 0x14425
    jnz SHORT L_014C
; Original operand: mov     ax, 2
L_0157: ; original 0x230b7; IDA unnamed; logical 0x130b7; file 0x14427
    mov ax, 02h
; Original operand: stc
L_015A: ; original 0x230ba; IDA unnamed; logical 0x130ba; file 0x1442a
    stc 
; Original operand: jmp     short loc_230BF
L_015B: ; original 0x230bb; IDA unnamed; logical 0x130bb; file 0x1442b
    jmp SHORT L_015F
; Original operand: sub     ax, ax
L_015D: ; original 0x230bd; IDA loc_230BD; logical 0x130bd; file 0x1442d
    sub ax, ax
; Original operand: or      ax, ax
L_015F: ; original 0x230bf; IDA loc_230BF; logical 0x130bf; file 0x1442f
    or ax, ax
; Original operand: retn
L_0161: ; original 0x230c1; IDA unnamed; logical 0x130c1; file 0x14431
    ret 
; Original operand: mov     bx, 2
L_0162: ; original 0x230c2; IDA sub_230C2; logical 0x130c2; file 0x14432
    mov bx, 02h
; Original operand: mov     al, 0E0h
L_0165: ; original 0x230c5; IDA unnamed; logical 0x130c5; file 0x14435
    mov al, 0E0h
; Original operand: mov     dx, ds:30h
L_0167: ; original 0x230c7; IDA unnamed; logical 0x130c7; file 0x14437
    mov dx, WORD PTR ds:[030h]
; Original operand: add     dx, 0Ch
L_016B: ; original 0x230cb; IDA unnamed; logical 0x130cb; file 0x1443b
    add dx, 0Ch
; Original operand: call    sub_23043
L_016E: ; original 0x230ce; IDA unnamed; logical 0x130ce; file 0x1443e
    call L_00E3
; Original operand: jb      short loc_230E5
L_0171: ; original 0x230d1; IDA unnamed; logical 0x130d1; file 0x14441
    jb SHORT L_0185
; Original operand: mov     al, 0AAh
L_0173: ; original 0x230d3; IDA unnamed; logical 0x130d3; file 0x14443
    mov al, 0AAh
; Original operand: call    sub_23043
L_0175: ; original 0x230d5; IDA unnamed; logical 0x130d5; file 0x14445
    call L_00E3
; Original operand: jb      short loc_230E5
L_0178: ; original 0x230d8; IDA unnamed; logical 0x130d8; file 0x14448
    jb SHORT L_0185
; Original operand: call    sub_23059
L_017A: ; original 0x230da; IDA unnamed; logical 0x130da; file 0x1444a
    call L_00F9
; Original operand: jb      short loc_230E5
L_017D: ; original 0x230dd; IDA unnamed; logical 0x130dd; file 0x1444d
    jb SHORT L_0185
; Original operand: cmp     al, 55h ; 'U'
L_017F: ; original 0x230df; IDA unnamed; logical 0x130df; file 0x1444f
    cmp al, 055h
; Original operand: jnz     short loc_230E5
L_0181: ; original 0x230e1; IDA unnamed; logical 0x130e1; file 0x14451
    jnz SHORT L_0185
; Original operand: sub     bx, bx
L_0183: ; original 0x230e3; IDA unnamed; logical 0x130e3; file 0x14453
    sub bx, bx
; Original operand: mov     ax, bx
L_0185: ; original 0x230e5; IDA loc_230E5; logical 0x130e5; file 0x14455
    mov ax, bx
; Original operand: or      ax, ax
L_0187: ; original 0x230e7; IDA unnamed; logical 0x130e7; file 0x14457
    or ax, ax
; Original operand: retn
L_0189: ; original 0x230e9; IDA unnamed; logical 0x130e9; file 0x14459
    ret 
; Original operand: mov     byte ptr ds:0B1h, 0
L_018A: ; original 0x230ea; IDA sub_230EA; logical 0x130ea; file 0x1445a
    mov BYTE PTR ds:[0B1h], 00h
; Original operand: mov     ax, ds:0B8h
L_018F: ; original 0x230ef; IDA unnamed; logical 0x130ef; file 0x1445f
    mov ax, WORD PTR ds:[0B8h]
; Original operand: or      ax, ax
L_0192: ; original 0x230f2; IDA unnamed; logical 0x130f2; file 0x14462
    or ax, ax
; Original operand: jz      short loc_230F9
L_0194: ; original 0x230f4; IDA unnamed; logical 0x130f4; file 0x14464
    jz SHORT L_0199
; Original operand: mov     ds:91h, ax
L_0196: ; original 0x230f6; IDA unnamed; logical 0x130f6; file 0x14466
    mov WORD PTR ds:[091h], ax
; Original operand: mov     ax, ds:91h
L_0199: ; original 0x230f9; IDA loc_230F9; logical 0x130f9; file 0x14469
    mov ax, WORD PTR ds:[091h]
; Original operand: mov     ds:0B8h, ax
L_019C: ; original 0x230fc; IDA unnamed; logical 0x130fc; file 0x1446c
    mov WORD PTR ds:[0B8h], ax
; Original operand: mov     ax, 34Eh
L_019F: ; original 0x230ff; IDA unnamed; logical 0x130ff; file 0x1446f
    mov ax, 034Eh
; Original operand: call    sub_2323B
L_01A2: ; original 0x23102; IDA unnamed; logical 0x13102; file 0x14472
    call L_02DB
; Original operand: mov     dx, cs
L_01A5: ; original 0x23105; IDA unnamed; logical 0x13105; file 0x14475
    mov dx, cs
; Original operand: mov     ax, 91h
L_01A7: ; original 0x23107; IDA unnamed; logical 0x13107; file 0x14477
    mov ax, 091h
; Original operand: call    sub_23204
L_01AA: ; original 0x2310a; IDA unnamed; logical 0x1310a; file 0x1447a
    call L_02A4
; Original operand: mov     cx, 1
L_01AD: ; original 0x2310d; IDA unnamed; logical 0x1310d; file 0x1447d
    mov cx, 01h
; Original operand: mov     dh, 45h ; 'E'
L_01B0: ; original 0x23110; IDA unnamed; logical 0x13110; file 0x14480
    mov dh, 045h
; Original operand: call    sub_231DB
L_01B2: ; original 0x23112; IDA unnamed; logical 0x13112; file 0x14482
    call L_027B
; Original operand: mov     dx, ds:30h
L_01B5: ; original 0x23115; IDA unnamed; logical 0x13115; file 0x14485
    mov dx, WORD PTR ds:[030h]
; Original operand: add     dx, 0Ch
L_01B9: ; original 0x23119; IDA unnamed; logical 0x13119; file 0x14489
    add dx, 0Ch
; Original operand: mov     al, 0E2h
L_01BC: ; original 0x2311c; IDA unnamed; logical 0x1311c; file 0x1448c
    mov al, 0E2h
; Original operand: call    sub_23077
L_01BE: ; original 0x2311e; IDA unnamed; logical 0x1311e; file 0x1448e
    call L_0117
; Original operand: mov     al, ds:91h
L_01C1: ; original 0x23121; IDA unnamed; logical 0x13121; file 0x14491
    mov al, BYTE PTR ds:[091h]
; Original operand: call    sub_23077
L_01C4: ; original 0x23124; IDA unnamed; logical 0x13124; file 0x14494
    call L_0117
; Original operand: mov     al, 0E2h
L_01C7: ; original 0x23127; IDA unnamed; logical 0x13127; file 0x14497
    mov al, 0E2h
; Original operand: call    sub_23077
L_01C9: ; original 0x23129; IDA unnamed; logical 0x13129; file 0x14499
    call L_0117
; Original operand: mov     al, ds:92h
L_01CC: ; original 0x2312c; IDA unnamed; logical 0x1312c; file 0x1449c
    mov al, BYTE PTR ds:[092h]
; Original operand: call    sub_23077
L_01CF: ; original 0x2312f; IDA unnamed; logical 0x1312f; file 0x1449f
    call L_0117
; Original operand: mov     al, 0E4h
L_01D2: ; original 0x23132; IDA unnamed; logical 0x13132; file 0x144a2
    mov al, 0E4h
; Original operand: call    sub_23077
L_01D4: ; original 0x23134; IDA unnamed; logical 0x13134; file 0x144a4
    call L_0117
; Original operand: mov     al, 0AAh
L_01D7: ; original 0x23137; IDA unnamed; logical 0x13137; file 0x144a7
    mov al, 0AAh
; Original operand: call    sub_23077
L_01D9: ; original 0x23139; IDA unnamed; logical 0x13139; file 0x144a9
    call L_0117
; Original operand: mov     al, 0E8h
L_01DC: ; original 0x2313c; IDA unnamed; logical 0x1313c; file 0x144ac
    mov al, 0E8h
; Original operand: call    sub_23077
L_01DE: ; original 0x2313e; IDA unnamed; logical 0x1313e; file 0x144ae
    call L_0117
; Original operand: call    sub_23084
L_01E1: ; original 0x23141; IDA loc_23141; logical 0x13141; file 0x144b1
    call L_0124
; Original operand: cmp     al, 0AAh
L_01E4: ; original 0x23144; IDA unnamed; logical 0x13144; file 0x144b4
    cmp al, 0AAh
; Original operand: jnz     short loc_23141
L_01E6: ; original 0x23146; IDA unnamed; logical 0x13146; file 0x144b6
    jnz SHORT L_01E1
; Original operand: mov     dx, cs
L_01E8: ; original 0x23148; IDA unnamed; logical 0x13148; file 0x144b8
    mov dx, cs
; Original operand: mov     ax, 0B6h
L_01EA: ; original 0x2314a; IDA unnamed; logical 0x1314a; file 0x144ba
    mov ax, 0B6h
; Original operand: call    sub_23204
L_01ED: ; original 0x2314d; IDA unnamed; logical 0x1314d; file 0x144bd
    call L_02A4
; Original operand: sub     cx, cx
L_01F0: ; original 0x23150; IDA unnamed; logical 0x13150; file 0x144c0
    sub cx, cx
; Original operand: mov     dh, 49h ; 'I'
L_01F2: ; original 0x23152; IDA unnamed; logical 0x13152; file 0x144c2
    mov dh, 049h
; Original operand: call    sub_231DB
L_01F4: ; original 0x23154; IDA unnamed; logical 0x13154; file 0x144c4
    call L_027B
; Original operand: mov     dx, ds:30h
L_01F7: ; original 0x23157; IDA unnamed; logical 0x13157; file 0x144c7
    mov dx, WORD PTR ds:[030h]
; Original operand: add     dx, 0Ch
L_01FB: ; original 0x2315b; IDA unnamed; logical 0x1315b; file 0x144cb
    add dx, 0Ch
; Original operand: mov     al, 40h ; '@'
L_01FE: ; original 0x2315e; IDA unnamed; logical 0x1315e; file 0x144ce
    mov al, 040h
; Original operand: call    sub_23077
L_0200: ; original 0x23160; IDA unnamed; logical 0x13160; file 0x144d0
    call L_0117
; Original operand: mov     al, 64h ; 'd'
L_0203: ; original 0x23163; IDA unnamed; logical 0x13163; file 0x144d3
    mov al, 064h
; Original operand: call    sub_23077
L_0205: ; original 0x23165; IDA unnamed; logical 0x13165; file 0x144d5
    call L_0117
; Original operand: mov     al, 14h
L_0208: ; original 0x23168; IDA unnamed; logical 0x13168; file 0x144d8
    mov al, 014h
; Original operand: call    sub_23077
L_020A: ; original 0x2316a; IDA unnamed; logical 0x1316a; file 0x144da
    call L_0117
; Original operand: sub     al, al
L_020D: ; original 0x2316d; IDA unnamed; logical 0x1316d; file 0x144dd
    sub al, al
; Original operand: call    sub_23077
L_020F: ; original 0x2316f; IDA unnamed; logical 0x1316f; file 0x144df
    call L_0117
; Original operand: sub     al, al
L_0212: ; original 0x23172; IDA unnamed; logical 0x13172; file 0x144e2
    sub al, al
; Original operand: call    sub_23077
L_0214: ; original 0x23174; IDA unnamed; logical 0x13174; file 0x144e4
    call L_0117
; Original operand: sub     ax, ax
L_0217: ; original 0x23177; IDA unnamed; logical 0x13177; file 0x144e7
    sub ax, ax
; Original operand: mov     cx, 2000h
L_0219: ; original 0x23179; IDA unnamed; logical 0x13179; file 0x144e9
    mov cx, 02000h
; Original operand: cmp     byte ptr ds:0B1h, 0
L_021C: ; original 0x2317c; IDA loc_2317C; logical 0x1317c; file 0x144ec
    cmp BYTE PTR ds:[0B1h], 00h
; Original operand: jnz     short loc_23188
L_0221: ; original 0x23181; IDA unnamed; logical 0x13181; file 0x144f1
    jnz SHORT L_0228
; Original operand: loop    loc_2317C
L_0223: ; original 0x23183; IDA unnamed; logical 0x13183; file 0x144f3
    loop L_021C
; Original operand: mov     ax, 3
L_0225: ; original 0x23185; IDA unnamed; logical 0x13185; file 0x144f5
    mov ax, 03h
; Original operand: push    ax
L_0228: ; original 0x23188; IDA loc_23188; logical 0x13188; file 0x144f8
    push ax
; Original operand: call    sub_2327D
L_0229: ; original 0x23189; IDA unnamed; logical 0x13189; file 0x144f9
    call L_031D
; Original operand: pop     ax
L_022C: ; original 0x2318c; IDA unnamed; logical 0x1318c; file 0x144fc
    pop ax
; Original operand: or      ax, ax
L_022D: ; original 0x2318d; IDA unnamed; logical 0x1318d; file 0x144fd
    or ax, ax
; Original operand: retn
L_022F: ; original 0x2318f; IDA unnamed; logical 0x1318f; file 0x144ff
    ret 
; Original operand: mov     al, 0E1h
L_0230: ; original 0x23190; IDA sub_23190; logical 0x13190; file 0x14500
    mov al, 0E1h
; Original operand: call    sub_23077
L_0232: ; original 0x23192; IDA unnamed; logical 0x13192; file 0x14502
    call L_0117
; Original operand: call    sub_23084
L_0235: ; original 0x23195; IDA unnamed; logical 0x13195; file 0x14505
    call L_0124
; Original operand: mov     ah, al
L_0238: ; original 0x23198; IDA unnamed; logical 0x13198; file 0x14508
    mov ah, al
; Original operand: call    sub_23084
L_023A: ; original 0x2319a; IDA unnamed; logical 0x1319a; file 0x1450a
    call L_0124
; Original operand: mov     bx, 1
L_023D: ; original 0x2319d; IDA unnamed; logical 0x1319d; file 0x1450d
    mov bx, 01h
; Original operand: cmp     ax, ds:35h
L_0240: ; original 0x231a0; IDA unnamed; logical 0x131a0; file 0x14510
    cmp ax, WORD PTR ds:[035h]
; Original operand: jb      short loc_231A8
L_0244: ; original 0x231a4; IDA unnamed; logical 0x131a4; file 0x14514
    jb SHORT L_0248
; Original operand: sub     bx, bx
L_0246: ; original 0x231a6; IDA unnamed; logical 0x131a6; file 0x14516
    sub bx, bx
; Original operand: mov     ax, bx
L_0248: ; original 0x231a8; IDA loc_231A8; logical 0x131a8; file 0x14518
    mov ax, bx
; Original operand: or      ax, ax
L_024A: ; original 0x231aa; IDA unnamed; logical 0x131aa; file 0x1451a
    or ax, ax
; Original operand: retn
L_024C: ; original 0x231ac; IDA unnamed; logical 0x131ac; file 0x1451c
    ret 
; Original operand: pushf
L_024D: ; original 0x231ad; IDA sub_231AD; logical 0x131ad; file 0x1451d
    pushf 
; Original operand: push    si
L_024E: ; original 0x231ae; IDA unnamed; logical 0x131ae; file 0x1451e
    push si
; Original operand: mov     ah, 0D0h
L_024F: ; original 0x231af; IDA unnamed; logical 0x131af; file 0x1451f
    mov ah, 0D0h
; Original operand: mov     bx, 0B2h
L_0251: ; original 0x231b1; IDA unnamed; logical 0x131b1; file 0x14521
    mov bx, 0B2h
; Original operand: sub     cx, cx
L_0254: ; original 0x231b4; IDA unnamed; logical 0x131b4; file 0x14524
    sub cx, cx
; Original operand: mov     dx, ds:30h
L_0256: ; original 0x231b6; IDA unnamed; logical 0x131b6; file 0x14526
    mov dx, WORD PTR ds:[030h]
; Original operand: add     dl, 0Ch
L_025A: ; original 0x231ba; IDA unnamed; logical 0x131ba; file 0x1452a
    add dl, 0Ch
; Original operand: mov     si, 0FFFFh
L_025D: ; original 0x231bd; IDA unnamed; logical 0x131bd; file 0x1452d
    mov si, 0FFFFh
; Original operand: sti
L_0260: ; original 0x231c0; IDA loc_231C0; logical 0x131c0; file 0x14530
    sti 
; Original operand: cmp     cl, [bx]
L_0261: ; original 0x231c1; IDA unnamed; logical 0x131c1; file 0x14531
    cmp cl, BYTE PTR [bx]
; Original operand: jz      short loc_231D8
L_0263: ; original 0x231c3; IDA unnamed; logical 0x131c3; file 0x14533
    jz SHORT L_0278
; Original operand: cli
L_0265: ; original 0x231c5; IDA unnamed; logical 0x131c5; file 0x14535
    cli 
; Original operand: in      al, dx
L_0266: ; original 0x231c6; IDA unnamed; logical 0x131c6; file 0x14536
    in al, dx
; Original operand: or      al, al
L_0267: ; original 0x231c7; IDA unnamed; logical 0x131c7; file 0x14537
    or al, al
; Original operand: js      short loc_231D0
L_0269: ; original 0x231c9; IDA unnamed; logical 0x131c9; file 0x14539
    js SHORT L_0270
; Original operand: dec     si
L_026B: ; original 0x231cb; IDA unnamed; logical 0x131cb; file 0x1453b
    dec si
; Original operand: jnz     short loc_231C0
L_026C: ; original 0x231cc; IDA unnamed; logical 0x131cc; file 0x1453c
    jnz SHORT L_0260
; Original operand: jmp     short loc_231D8
L_026E: ; original 0x231ce; IDA unnamed; logical 0x131ce; file 0x1453e
    jmp SHORT L_0278
; Original operand: in      al, dx
L_0270: ; original 0x231d0; IDA loc_231D0; logical 0x131d0; file 0x14540
    in al, dx
; Original operand: or      al, al
L_0271: ; original 0x231d1; IDA unnamed; logical 0x131d1; file 0x14541
    or al, al
; Original operand: js      short loc_231D0
L_0273: ; original 0x231d3; IDA unnamed; logical 0x131d3; file 0x14543
    js SHORT L_0270
; Original operand: mov     al, ah
L_0275: ; original 0x231d5; IDA unnamed; logical 0x131d5; file 0x14545
    mov al, ah
; Original operand: out     dx, al
L_0277: ; original 0x231d7; IDA unnamed; logical 0x131d7; file 0x14547
    out dx, al
; Original operand: pop     si
L_0278: ; original 0x231d8; IDA loc_231D8; logical 0x131d8; file 0x14548
    pop si
; Original operand: popf
L_0279: ; original 0x231d9; IDA unnamed; logical 0x131d9; file 0x14549
    popf 
; Original operand: retn
L_027A: ; original 0x231da; IDA unnamed; logical 0x131da; file 0x1454a
    ret 
; Original operand: push    bx
L_027B: ; original 0x231db; IDA sub_231DB; logical 0x131db; file 0x1454b
    push bx
; Original operand: mov     bx, ax
L_027C: ; original 0x231dc; IDA unnamed; logical 0x131dc; file 0x1454c
    mov bx, ax
; Original operand: mov     al, 5
L_027E: ; original 0x231de; IDA unnamed; logical 0x131de; file 0x1454e
    mov al, 05h
; Original operand: out     0Ah, al; DMA controller, 8237A-5.
L_0280: ; original 0x231e0; IDA unnamed; logical 0x131e0; file 0x14550
    out 0Ah, al
; Original operand: sub     al, al
L_0282: ; original 0x231e2; IDA unnamed; logical 0x131e2; file 0x14552
    sub al, al
; Original operand: out     0Ch, al; DMA controller, 8237A-5.
L_0284: ; original 0x231e4; IDA unnamed; logical 0x131e4; file 0x14554
    out 0Ch, al
; Original operand: mov     al, dh
L_0286: ; original 0x231e6; IDA unnamed; logical 0x131e6; file 0x14556
    mov al, dh
; Original operand: out     0Bh, al; DMA 8237A-5. mode register bits:
L_0288: ; original 0x231e8; IDA unnamed; logical 0x131e8; file 0x14558
    out 0Bh, al
; Original operand: mov     al, bl
L_028A: ; original 0x231ea; IDA unnamed; logical 0x131ea; file 0x1455a
    mov al, bl
; Original operand: out     2, al; DMA controller, 8237A-5.
L_028C: ; original 0x231ec; IDA unnamed; logical 0x131ec; file 0x1455c
    out 02h, al
; Original operand: mov     al, bh
L_028E: ; original 0x231ee; IDA unnamed; logical 0x131ee; file 0x1455e
    mov al, bh
; Original operand: out     2, al; DMA controller, 8237A-5.
L_0290: ; original 0x231f0; IDA unnamed; logical 0x131f0; file 0x14560
    out 02h, al
; Original operand: mov     al, cl
L_0292: ; original 0x231f2; IDA unnamed; logical 0x131f2; file 0x14562
    mov al, cl
; Original operand: out     3, al; DMA controller, 8237A-5.
L_0294: ; original 0x231f4; IDA unnamed; logical 0x131f4; file 0x14564
    out 03h, al
; Original operand: mov     al, ch
L_0296: ; original 0x231f6; IDA unnamed; logical 0x131f6; file 0x14566
    mov al, ch
; Original operand: out     3, al; DMA controller, 8237A-5.
L_0298: ; original 0x231f8; IDA unnamed; logical 0x131f8; file 0x14568
    out 03h, al
; Original operand: mov     al, dl
L_029A: ; original 0x231fa; IDA unnamed; logical 0x131fa; file 0x1456a
    mov al, dl
; Original operand: out     83h, al; DMA page register 74LS612:
L_029C: ; original 0x231fc; IDA unnamed; logical 0x131fc; file 0x1456c
    out 083h, al
; Original operand: mov     al, 1
L_029E: ; original 0x231fe; IDA unnamed; logical 0x131fe; file 0x1456e
    mov al, 01h
; Original operand: out     0Ah, al; DMA controller, 8237A-5.
L_02A0: ; original 0x23200; IDA unnamed; logical 0x13200; file 0x14570
    out 0Ah, al
; Original operand: pop     bx
L_02A2: ; original 0x23202; IDA unnamed; logical 0x13202; file 0x14572
    pop bx
; Original operand: retn
L_02A3: ; original 0x23203; IDA unnamed; logical 0x13203; file 0x14573
    ret 
; Original operand: push    cx
L_02A4: ; original 0x23204; IDA sub_23204; logical 0x13204; file 0x14574
    push cx
; Original operand: mov     cl, 4
L_02A5: ; original 0x23205; IDA unnamed; logical 0x13205; file 0x14575
    mov cl, 04h
; Original operand: rol     dx, cl
L_02A7: ; original 0x23207; IDA unnamed; logical 0x13207; file 0x14577
    rol dx, cl
; Original operand: mov     cx, dx
L_02A9: ; original 0x23209; IDA unnamed; logical 0x13209; file 0x14579
    mov cx, dx
; Original operand: and     dx, 0Fh
L_02AB: ; original 0x2320b; IDA unnamed; logical 0x1320b; file 0x1457b
    and dx, 0Fh
; Original operand: and     cx, 0FFF0h
L_02AE: ; original 0x2320e; IDA unnamed; logical 0x1320e; file 0x1457e
    and cx, 0FFF0h
; Original operand: add     ax, cx
L_02B1: ; original 0x23211; IDA unnamed; logical 0x13211; file 0x14581
    add ax, cx
; Original operand: adc     dx, 0
L_02B3: ; original 0x23213; IDA unnamed; logical 0x13213; file 0x14583
    adc dx, 00h
; Original operand: pop     cx
L_02B6: ; original 0x23216; IDA unnamed; logical 0x13216; file 0x14586
    pop cx
; Original operand: retn
L_02B7: ; original 0x23217; IDA unnamed; logical 0x13217; file 0x14587
    ret 
; Original operand: mov     dx, ds:0CCh
L_02B8: ; original 0x23218; IDA sub_23218; logical 0x13218; file 0x14588
    mov dx, WORD PTR ds:[0CCh]
; Original operand: add     ax, ds:0CAh
L_02BC: ; original 0x2321c; IDA unnamed; logical 0x1321c; file 0x1458c
    add ax, WORD PTR ds:[0CAh]
; Original operand: jnb     short locret_23225
L_02C0: ; original 0x23220; IDA unnamed; logical 0x13220; file 0x14590
    jnb SHORT L_02C5
; Original operand: add     dh, 10h
L_02C2: ; original 0x23222; IDA unnamed; logical 0x13222; file 0x14592
    add dh, 010h
; Original operand: retn
L_02C5: ; original 0x23225; IDA locret_23225; logical 0x13225; file 0x14595
    ret 
; Original operand: push    es
L_02C6: ; original 0x23226; IDA sub_23226; logical 0x13226; file 0x14596
    push es
; Original operand: push    di
L_02C7: ; original 0x23227; IDA unnamed; logical 0x13227; file 0x14597
    push di
; Original operand: les     di, ds:0CAh
L_02C8: ; original 0x23228; IDA unnamed; logical 0x13228; file 0x14598
    les di, DWORD PTR ds:[0CAh]
; Original operand: mov     ax, es:[di+1]
L_02CC: ; original 0x2322c; IDA unnamed; logical 0x1322c; file 0x1459c
    mov ax, WORD PTR es:[di+DISP_0001]
; Original operand: mov     dl, es:[di+3]
L_02D1: ; original 0x23231; IDA unnamed; logical 0x13231; file 0x145a1
    mov dl, BYTE PTR es:[di+DISP_0003]
; Original operand: sub     dh, dh
L_02D6: ; original 0x23236; IDA unnamed; logical 0x13236; file 0x145a6
    sub dh, dh
; Original operand: pop     di
L_02D8: ; original 0x23238; IDA unnamed; logical 0x13238; file 0x145a8
    pop di
; Original operand: pop     es
L_02D9: ; original 0x23239; IDA unnamed; logical 0x13239; file 0x145a9
    pop es
; Original operand: retn
L_02DA: ; original 0x2323a; IDA unnamed; logical 0x1323a; file 0x145aa
    ret 
; Original operand: pushf
L_02DB: ; original 0x2323b; IDA sub_2323B; logical 0x1323b; file 0x145ab
    pushf 
; Original operand: push    bx
L_02DC: ; original 0x2323c; IDA unnamed; logical 0x1323c; file 0x145ac
    push bx
; Original operand: push    cx
L_02DD: ; original 0x2323d; IDA unnamed; logical 0x1323d; file 0x145ad
    push cx
; Original operand: push    dx
L_02DE: ; original 0x2323e; IDA unnamed; logical 0x1323e; file 0x145ae
    push dx
; Original operand: cli
L_02DF: ; original 0x2323f; IDA unnamed; logical 0x1323f; file 0x145af
    cli 
; Original operand: mov     dx, ax
L_02E0: ; original 0x23240; IDA unnamed; logical 0x13240; file 0x145b0
    mov dx, ax
; Original operand: mov     al, ds:32h
L_02E2: ; original 0x23242; IDA unnamed; logical 0x13242; file 0x145b2
    mov al, BYTE PTR ds:[032h]
; Original operand: add     al, 8
L_02E5: ; original 0x23245; IDA unnamed; logical 0x13245; file 0x145b5
    add al, 08h
; Original operand: cbw
L_02E7: ; original 0x23247; IDA unnamed; logical 0x13247; file 0x145b7
    cbw 
; Original operand: shl     al, 1
L_02E8: ; original 0x23248; IDA unnamed; logical 0x13248; file 0x145b8
    shl al, 01h
; Original operand: shl     al, 1
L_02EA: ; original 0x2324a; IDA unnamed; logical 0x1324a; file 0x145ba
    shl al, 01h
; Original operand: mov     bx, ax
L_02EC: ; original 0x2324c; IDA unnamed; logical 0x1324c; file 0x145bc
    mov bx, ax
; Original operand: push    es
L_02EE: ; original 0x2324e; IDA unnamed; logical 0x1324e; file 0x145be
    push es
; Original operand: sub     ax, ax
L_02EF: ; original 0x2324f; IDA unnamed; logical 0x1324f; file 0x145bf
    sub ax, ax
; Original operand: mov     es, ax
L_02F1: ; original 0x23251; IDA unnamed; logical 0x13251; file 0x145c1
    mov es, ax
; Original operand: mov     ax, es:[bx]
L_02F3: ; original 0x23253; IDA unnamed; logical 0x13253; file 0x145c3
    mov ax, WORD PTR es:[bx]
; Original operand: mov     ds:0BEh, ax
L_02F6: ; original 0x23256; IDA unnamed; logical 0x13256; file 0x145c6
    mov WORD PTR ds:[0BEh], ax
; Original operand: mov     es:[bx], dx
L_02F9: ; original 0x23259; IDA unnamed; logical 0x13259; file 0x145c9
    mov WORD PTR es:[bx], dx
; Original operand: mov     ax, es:[bx+2]
L_02FC: ; original 0x2325c; IDA unnamed; logical 0x1325c; file 0x145cc
    mov ax, WORD PTR es:[bx+02h]
; Original operand: mov     ds:0C0h, ax
L_0300: ; original 0x23260; IDA unnamed; logical 0x13260; file 0x145d0
    mov WORD PTR ds:[0C0h], ax
; Original operand: mov     word ptr es:[bx+2], cs
L_0303: ; original 0x23263; IDA unnamed; logical 0x13263; file 0x145d3
    mov WORD PTR es:[bx+02h], cs
; Original operand: pop     es
L_0307: ; original 0x23267; IDA unnamed; logical 0x13267; file 0x145d7
    pop es
; Original operand: mov     cl, ds:32h
L_0308: ; original 0x23268; IDA unnamed; logical 0x13268; file 0x145d8
    mov cl, BYTE PTR ds:[032h]
; Original operand: mov     ah, 1
L_030C: ; original 0x2326c; IDA unnamed; logical 0x1326c; file 0x145dc
    mov ah, 01h
; Original operand: shl     ah, cl
L_030E: ; original 0x2326e; IDA unnamed; logical 0x1326e; file 0x145de
    shl ah, cl
; Original operand: not     ah
L_0310: ; original 0x23270; IDA unnamed; logical 0x13270; file 0x145e0
    not ah
; Original operand: in      al, 21h; Interrupt controller, 8259A.
L_0312: ; original 0x23272; IDA unnamed; logical 0x13272; file 0x145e2
    in al, 021h
; Original operand: and     al, ah
L_0314: ; original 0x23274; IDA unnamed; logical 0x13274; file 0x145e4
    and al, ah
; Original operand: out     21h, al; Interrupt controller, 8259A.
L_0316: ; original 0x23276; IDA unnamed; logical 0x13276; file 0x145e6
    out 021h, al
; Original operand: pop     dx
L_0318: ; original 0x23278; IDA unnamed; logical 0x13278; file 0x145e8
    pop dx
; Original operand: pop     cx
L_0319: ; original 0x23279; IDA unnamed; logical 0x13279; file 0x145e9
    pop cx
; Original operand: pop     bx
L_031A: ; original 0x2327a; IDA unnamed; logical 0x1327a; file 0x145ea
    pop bx
; Original operand: popf
L_031B: ; original 0x2327b; IDA unnamed; logical 0x1327b; file 0x145eb
    popf 
; Original operand: retn
L_031C: ; original 0x2327c; IDA unnamed; logical 0x1327c; file 0x145ec
    ret 
; Original operand: pushf
L_031D: ; original 0x2327d; IDA sub_2327D; logical 0x1327d; file 0x145ed
    pushf 
; Original operand: cli
L_031E: ; original 0x2327e; IDA unnamed; logical 0x1327e; file 0x145ee
    cli 
; Original operand: mov     al, ds:32h
L_031F: ; original 0x2327f; IDA unnamed; logical 0x1327f; file 0x145ef
    mov al, BYTE PTR ds:[032h]
; Original operand: add     al, 8
L_0322: ; original 0x23282; IDA unnamed; logical 0x13282; file 0x145f2
    add al, 08h
; Original operand: cbw
L_0324: ; original 0x23284; IDA unnamed; logical 0x13284; file 0x145f4
    cbw 
; Original operand: shl     al, 1
L_0325: ; original 0x23285; IDA unnamed; logical 0x13285; file 0x145f5
    shl al, 01h
; Original operand: shl     al, 1
L_0327: ; original 0x23287; IDA unnamed; logical 0x13287; file 0x145f7
    shl al, 01h
; Original operand: mov     di, ax
L_0329: ; original 0x23289; IDA unnamed; logical 0x13289; file 0x145f9
    mov di, ax
; Original operand: push    es
L_032B: ; original 0x2328b; IDA unnamed; logical 0x1328b; file 0x145fb
    push es
; Original operand: sub     ax, ax
L_032C: ; original 0x2328c; IDA unnamed; logical 0x1328c; file 0x145fc
    sub ax, ax
; Original operand: mov     es, ax
L_032E: ; original 0x2328e; IDA unnamed; logical 0x1328e; file 0x145fe
    mov es, ax
; Original operand: mov     ax, ds:0BEh
L_0330: ; original 0x23290; IDA unnamed; logical 0x13290; file 0x14600
    mov ax, WORD PTR ds:[0BEh]
; Original operand: mov     es:[di], ax
L_0333: ; original 0x23293; IDA unnamed; logical 0x13293; file 0x14603
    mov WORD PTR es:[di], ax
; Original operand: mov     ax, ds:0C0h
L_0336: ; original 0x23296; IDA unnamed; logical 0x13296; file 0x14606
    mov ax, WORD PTR ds:[0C0h]
; Original operand: mov     es:[di+2], ax
L_0339: ; original 0x23299; IDA unnamed; logical 0x13299; file 0x14609
    mov WORD PTR es:[di+02h], ax
; Original operand: pop     es
L_033D: ; original 0x2329d; IDA unnamed; logical 0x1329d; file 0x1460d
    pop es
; Original operand: mov     cl, ds:32h
L_033E: ; original 0x2329e; IDA unnamed; logical 0x1329e; file 0x1460e
    mov cl, BYTE PTR ds:[032h]
; Original operand: mov     ah, 1
L_0342: ; original 0x232a2; IDA unnamed; logical 0x132a2; file 0x14612
    mov ah, 01h
; Original operand: shl     ah, cl
L_0344: ; original 0x232a4; IDA unnamed; logical 0x132a4; file 0x14614
    shl ah, cl
; Original operand: in      al, 21h; Interrupt controller, 8259A.
L_0346: ; original 0x232a6; IDA unnamed; logical 0x132a6; file 0x14616
    in al, 021h
; Original operand: or      al, ah
L_0348: ; original 0x232a8; IDA unnamed; logical 0x132a8; file 0x14618
    or al, ah
; Original operand: out     21h, al; Interrupt controller, 8259A.
L_034A: ; original 0x232aa; IDA unnamed; logical 0x132aa; file 0x1461a
    out 021h, al
; Original operand: popf
L_034C: ; original 0x232ac; IDA unnamed; logical 0x132ac; file 0x1461c
    popf 
; Original operand: retn
L_034D: ; original 0x232ad; IDA unnamed; logical 0x132ad; file 0x1461d
    ret 
; Original operand: push    ds
L_034E: ; original 0x232ae; IDA unnamed; logical 0x132ae; file 0x1461e
    push ds
; Original operand: push    ax
L_034F: ; original 0x232af; IDA unnamed; logical 0x132af; file 0x1461f
    push ax
; Original operand: push    dx
L_0350: ; original 0x232b0; IDA unnamed; logical 0x132b0; file 0x14620
    push dx
; Original operand: mov     ax, cs
L_0351: ; original 0x232b1; IDA unnamed; logical 0x132b1; file 0x14621
    mov ax, cs
; Original operand: mov     ds, ax
L_0353: ; original 0x232b3; IDA unnamed; logical 0x132b3; file 0x14623
    mov ds, ax
; Original operand: mov     byte_23011, 1
L_0355: ; original 0x232b5; IDA unnamed; logical 0x132b5; file 0x14625
    mov BYTE PTR ds:[0B1h], 01h
; Original operand: mov     al, 20h ; ' '
L_035A: ; original 0x232ba; IDA unnamed; logical 0x132ba; file 0x1462a
    mov al, 020h
; Original operand: out     20h, al; Interrupt controller, 8259A.
L_035C: ; original 0x232bc; IDA unnamed; logical 0x132bc; file 0x1462c
    out 020h, al
; Original operand: mov     dx, word ptr byte_22F90
L_035E: ; original 0x232be; IDA unnamed; logical 0x132be; file 0x1462e
    mov dx, WORD PTR ds:[030h]
; Original operand: add     dx, 0Eh
L_0362: ; original 0x232c2; IDA unnamed; logical 0x132c2; file 0x14632
    add dx, 0Eh
; Original operand: in      al, dx
L_0365: ; original 0x232c5; IDA unnamed; logical 0x132c5; file 0x14635
    in al, dx
; Original operand: pop     dx
L_0366: ; original 0x232c6; IDA unnamed; logical 0x132c6; file 0x14636
    pop dx
; Original operand: pop     ax
L_0367: ; original 0x232c7; IDA unnamed; logical 0x132c7; file 0x14637
    pop ax
; Original operand: pop     ds
L_0368: ; original 0x232c8; IDA unnamed; logical 0x132c8; file 0x14638
    pop ds
; Original operand: iret
L_0369: ; original 0x232c9; IDA unnamed; logical 0x132c9; file 0x14639
    iret 
; Original operand: push    ds
L_036A: ; original 0x232ca; IDA sub_232CA; logical 0x132ca; file 0x1463a
    push ds
; Original operand: push    bx
L_036B: ; original 0x232cb; IDA unnamed; logical 0x132cb; file 0x1463b
    push bx
; Original operand: lds     bx, ds:0C2h
L_036C: ; original 0x232cc; IDA unnamed; logical 0x132cc; file 0x1463c
    lds bx, DWORD PTR ds:[0C2h]
; Original operand: mov     [bx], ax
L_0370: ; original 0x232d0; IDA unnamed; logical 0x132d0; file 0x14640
    mov WORD PTR [bx], ax
; Original operand: pop     bx
L_0372: ; original 0x232d2; IDA unnamed; logical 0x132d2; file 0x14642
    pop bx
; Original operand: pop     ds
L_0373: ; original 0x232d3; IDA unnamed; logical 0x132d3; file 0x14643
    pop ds
; Original operand: retn
L_0374: ; original 0x232d4; IDA unnamed; logical 0x132d4; file 0x14644
    ret 
; Original operand: push    ds
L_0375: ; original 0x232d5; IDA sub_232D5; logical 0x132d5; file 0x14645
    push ds
; Original operand: push    bx
L_0376: ; original 0x232d6; IDA unnamed; logical 0x132d6; file 0x14646
    push bx
; Original operand: lds     bx, ds:0CAh
L_0377: ; original 0x232d7; IDA unnamed; logical 0x132d7; file 0x14647
    lds bx, DWORD PTR ds:[0CAh]
; Original operand: mov     al, [bx+0]
L_037B: ; original 0x232db; IDA unnamed; logical 0x132db; file 0x1464b
    mov al, BYTE PTR [bx+DISP_0000]
; Original operand: pop     bx
L_037F: ; original 0x232df; IDA unnamed; logical 0x132df; file 0x1464f
    pop bx
; Original operand: pop     ds
L_0380: ; original 0x232e0; IDA unnamed; logical 0x132e0; file 0x14650
    pop ds
; Original operand: retn
L_0381: ; original 0x232e1; IDA unnamed; logical 0x132e1; file 0x14651
    ret 
; Original operand: mov     cx, ax
L_0382: ; original 0x232e2; IDA sub_232E2; logical 0x132e2; file 0x14652
    mov cx, ax
; Original operand: call    sub_23218
L_0384: ; original 0x232e4; IDA unnamed; logical 0x132e4; file 0x14654
    call L_02B8
; Original operand: call    sub_23204
L_0387: ; original 0x232e7; IDA unnamed; logical 0x132e7; file 0x14657
    call L_02A4
; Original operand: mov     ds:0D6h, dl
L_038A: ; original 0x232ea; IDA unnamed; logical 0x132ea; file 0x1465a
    mov BYTE PTR ds:[0D6h], dl
; Original operand: mov     ds:0D7h, ax
L_038E: ; original 0x232ee; IDA unnamed; logical 0x132ee; file 0x1465e
    mov WORD PTR ds:[0D7h], ax
; Original operand: call    sub_23226
L_0391: ; original 0x232f1; IDA unnamed; logical 0x132f1; file 0x14661
    call L_02C6
; Original operand: sub     cx, 4
L_0394: ; original 0x232f4; IDA unnamed; logical 0x132f4; file 0x14664
    sub cx, 04h
; Original operand: sub     ax, cx
L_0397: ; original 0x232f7; IDA unnamed; logical 0x132f7; file 0x14667
    sub ax, cx
; Original operand: sbb     dx, 0
L_0399: ; original 0x232f9; IDA unnamed; logical 0x132f9; file 0x14669
    sbb dx, 00h
; Original operand: mov     ds:0DCh, ax
L_039C: ; original 0x232fc; IDA unnamed; logical 0x132fc; file 0x1466c
    mov WORD PTR ds:[0DCh], ax
; Original operand: mov     ds:0DEh, dx
L_039F: ; original 0x232ff; IDA unnamed; logical 0x132ff; file 0x1466f
    mov WORD PTR ds:[0DEh], dx
; Original operand: sub     ax, 1
L_03A3: ; original 0x23303; IDA unnamed; logical 0x13303; file 0x14673
    sub ax, IMM_0001
; Original operand: sbb     dx, 0
L_03A6: ; original 0x23306; IDA unnamed; logical 0x13306; file 0x14676
    sbb dx, 00h
; Original operand: add     ax, ds:0D7h
L_03A9: ; original 0x23309; IDA unnamed; logical 0x13309; file 0x14679
    add ax, WORD PTR ds:[0D7h]
; Original operand: adc     dl, ds:0D6h
L_03AD: ; original 0x2330d; IDA unnamed; logical 0x1330d; file 0x1467d
    adc dl, BYTE PTR ds:[0D6h]
; Original operand: mov     ds:0E0h, ax
L_03B1: ; original 0x23311; IDA unnamed; logical 0x13311; file 0x14681
    mov WORD PTR ds:[0E0h], ax
; Original operand: sub     dl, ds:0D6h
L_03B4: ; original 0x23314; IDA unnamed; logical 0x13314; file 0x14684
    sub dl, BYTE PTR ds:[0D6h]
; Original operand: mov     ds:0DBh, dl
L_03B8: ; original 0x23318; IDA unnamed; logical 0x13318; file 0x14688
    mov BYTE PTR ds:[0DBh], dl
; Original operand: retn
L_03BC: ; original 0x2331c; IDA unnamed; logical 0x1331c; file 0x1468c
    ret 
; Original operand: call    sub_23204
L_03BD: ; original 0x2331d; IDA sub_2331D; logical 0x1331d; file 0x1468d
    call L_02A4
; Original operand: mov     ds:0D6h, dl
L_03C0: ; original 0x23320; IDA unnamed; logical 0x13320; file 0x14690
    mov BYTE PTR ds:[0D6h], dl
; Original operand: mov     ds:0D7h, ax
L_03C4: ; original 0x23324; IDA unnamed; logical 0x13324; file 0x14694
    mov WORD PTR ds:[0D7h], ax
; Original operand: add     ax, ds:0DCh
L_03C7: ; original 0x23327; IDA unnamed; logical 0x13327; file 0x14697
    add ax, WORD PTR ds:[0DCh]
; Original operand: adc     dx, ds:0DEh
L_03CB: ; original 0x2332b; IDA unnamed; logical 0x1332b; file 0x1469b
    adc dx, WORD PTR ds:[0DEh]
; Original operand: sub     ax, 1
L_03CF: ; original 0x2332f; IDA unnamed; logical 0x1332f; file 0x1469f
    sub ax, IMM_0001
; Original operand: sbb     dx, 0
L_03D2: ; original 0x23332; IDA unnamed; logical 0x13332; file 0x146a2
    sbb dx, 00h
; Original operand: mov     ds:0E0h, ax
L_03D5: ; original 0x23335; IDA unnamed; logical 0x13335; file 0x146a5
    mov WORD PTR ds:[0E0h], ax
; Original operand: sub     dl, ds:0D6h
L_03D8: ; original 0x23338; IDA unnamed; logical 0x13338; file 0x146a8
    sub dl, BYTE PTR ds:[0D6h]
; Original operand: mov     ds:0DBh, dl
L_03DC: ; original 0x2333c; IDA unnamed; logical 0x1333c; file 0x146ac
    mov BYTE PTR ds:[0DBh], dl
; Original operand: retn
L_03E0: ; original 0x23340; IDA unnamed; logical 0x13340; file 0x146b0
    ret 
; Original operand: push    ds
L_03E1: ; original 0x23341; IDA unnamed; logical 0x13341; file 0x146b1
    push ds
; Original operand: push    es
L_03E2: ; original 0x23342; IDA unnamed; logical 0x13342; file 0x146b2
    push es
; Original operand: push    ax
L_03E3: ; original 0x23343; IDA unnamed; logical 0x13343; file 0x146b3
    push ax
; Original operand: push    bx
L_03E4: ; original 0x23344; IDA unnamed; logical 0x13344; file 0x146b4
    push bx
; Original operand: push    cx
L_03E5: ; original 0x23345; IDA unnamed; logical 0x13345; file 0x146b5
    push cx
; Original operand: push    dx
L_03E6: ; original 0x23346; IDA unnamed; logical 0x13346; file 0x146b6
    push dx
; Original operand: push    di
L_03E7: ; original 0x23347; IDA unnamed; logical 0x13347; file 0x146b7
    push di
; Original operand: push    si
L_03E8: ; original 0x23348; IDA unnamed; logical 0x13348; file 0x146b8
    push si
; Original operand: push    bp
L_03E9: ; original 0x23349; IDA unnamed; logical 0x13349; file 0x146b9
    push bp
; Original operand: cld
L_03EA: ; original 0x2334a; IDA unnamed; logical 0x1334a; file 0x146ba
    cld 
; Original operand: mov     ax, cs
L_03EB: ; original 0x2334b; IDA unnamed; logical 0x1334b; file 0x146bb
    mov ax, cs
; Original operand: mov     ds, ax
L_03ED: ; original 0x2334d; IDA unnamed; logical 0x1334d; file 0x146bd
    mov ds, ax
; Original operand: mov     es, ax
L_03EF: ; original 0x2334f; IDA unnamed; logical 0x1334f; file 0x146bf
    mov es, ax
; Original operand: mov     dx, word ptr byte_22F90
L_03F1: ; original 0x23351; IDA unnamed; logical 0x13351; file 0x146c1
    mov dx, WORD PTR ds:[030h]
; Original operand: add     dl, 0Eh
L_03F5: ; original 0x23355; IDA unnamed; logical 0x13355; file 0x146c5
    add dl, 0Eh
; Original operand: in      al, dx
L_03F8: ; original 0x23358; IDA unnamed; logical 0x13358; file 0x146c8
    in al, dx
; Original operand: mov     al, 20h ; ' '
L_03F9: ; original 0x23359; IDA unnamed; logical 0x13359; file 0x146c9
    mov al, 020h
; Original operand: out     20h, al; Interrupt controller, 8259A.
L_03FB: ; original 0x2335b; IDA unnamed; logical 0x1335b; file 0x146cb
    out 020h, al
; Original operand: sti
L_03FD: ; original 0x2335d; IDA unnamed; logical 0x1335d; file 0x146cd
    sti 
; Original operand: mov     ax, word ptr byte_23012+2Ah
L_03FE: ; original 0x2335e; IDA unnamed; logical 0x1335e; file 0x146ce
    mov ax, WORD PTR ds:[0DCh]
; Original operand: or      ax, word ptr byte_23012+2Ch
L_0401: ; original 0x23361; IDA unnamed; logical 0x13361; file 0x146d1
    or ax, WORD PTR ds:[0DEh]
; Original operand: jnz     short loc_23379
L_0405: ; original 0x23365; IDA unnamed; logical 0x13365; file 0x146d5
    jnz SHORT L_0419
; Original operand: call    sub_23546
L_0407: ; original 0x23367; IDA unnamed; logical 0x13367; file 0x146d7
    call L_05E6
; Original operand: call    sub_23511
L_040A: ; original 0x2336a; IDA unnamed; logical 0x1336a; file 0x146da
    call L_05B1
; Original operand: cmp     byte_23012+30h, 0
L_040D: ; original 0x2336d; IDA unnamed; logical 0x1336d; file 0x146dd
    cmp BYTE PTR ds:[0E2h], 00h
; Original operand: jz      short loc_2337C
L_0412: ; original 0x23372; IDA unnamed; logical 0x13372; file 0x146e2
    jz SHORT L_041C
; Original operand: call    sub_233EC
L_0414: ; original 0x23374; IDA unnamed; logical 0x13374; file 0x146e4
    call L_048C
; Original operand: jmp     short loc_2337C
L_0417: ; original 0x23377; IDA unnamed; logical 0x13377; file 0x146e7
    jmp SHORT L_041C
; Original operand: call    sub_23386
L_0419: ; original 0x23379; IDA loc_23379; logical 0x13379; file 0x146e9
    call L_0426
; Original operand: pop     bp
L_041C: ; original 0x2337c; IDA loc_2337C; logical 0x1337c; file 0x146ec
    pop bp
; Original operand: pop     si
L_041D: ; original 0x2337d; IDA unnamed; logical 0x1337d; file 0x146ed
    pop si
; Original operand: pop     di
L_041E: ; original 0x2337e; IDA unnamed; logical 0x1337e; file 0x146ee
    pop di
; Original operand: pop     dx
L_041F: ; original 0x2337f; IDA unnamed; logical 0x1337f; file 0x146ef
    pop dx
; Original operand: pop     cx
L_0420: ; original 0x23380; IDA unnamed; logical 0x13380; file 0x146f0
    pop cx
; Original operand: pop     bx
L_0421: ; original 0x23381; IDA unnamed; logical 0x13381; file 0x146f1
    pop bx
; Original operand: pop     ax
L_0422: ; original 0x23382; IDA unnamed; logical 0x13382; file 0x146f2
    pop ax
; Original operand: pop     es
L_0423: ; original 0x23383; IDA unnamed; logical 0x13383; file 0x146f3
    pop es
; Original operand: pop     ds
L_0424: ; original 0x23384; IDA unnamed; logical 0x13384; file 0x146f4
    pop ds
; Original operand: iret
L_0425: ; original 0x23385; IDA unnamed; logical 0x13385; file 0x146f5
    iret 
; Original operand: mov     cx, 0FFFFh
L_0426: ; original 0x23386; IDA sub_23386; logical 0x13386; file 0x146f6
    mov cx, 0FFFFh
; Original operand: cmp     byte ptr ds:0DBh, 0
L_0429: ; original 0x23389; IDA unnamed; logical 0x13389; file 0x146f9
    cmp BYTE PTR ds:[0DBh], 00h
; Original operand: jnz     short loc_23398
L_042E: ; original 0x2338e; IDA unnamed; logical 0x1338e; file 0x146fe
    jnz SHORT L_0438
; Original operand: inc     byte ptr ds:0DBh
L_0430: ; original 0x23390; IDA unnamed; logical 0x13390; file 0x14700
    inc BYTE PTR ds:[0DBh]
; Original operand: mov     cx, ds:0E0h
L_0434: ; original 0x23394; IDA unnamed; logical 0x13394; file 0x14704
    mov cx, WORD PTR ds:[0E0h]
; Original operand: sub     cx, ds:0D7h
L_0438: ; original 0x23398; IDA loc_23398; logical 0x13398; file 0x14708
    sub cx, WORD PTR ds:[0D7h]
; Original operand: mov     ds:0D9h, cx
L_043C: ; original 0x2339c; IDA unnamed; logical 0x1339c; file 0x1470c
    mov WORD PTR ds:[0D9h], cx
; Original operand: inc     cx
L_0440: ; original 0x233a0; IDA unnamed; logical 0x133a0; file 0x14710
    inc cx
; Original operand: jz      short loc_233AE
L_0441: ; original 0x233a1; IDA unnamed; logical 0x133a1; file 0x14711
    jz SHORT L_044E
; Original operand: sub     ds:0DCh, cx
L_0443: ; original 0x233a3; IDA unnamed; logical 0x133a3; file 0x14713
    sub WORD PTR ds:[0DCh], cx
; Original operand: sbb     word ptr ds:0DEh, 0
L_0447: ; original 0x233a7; IDA unnamed; logical 0x133a7; file 0x14717
    sbb WORD PTR ds:[0DEh], 00h
; Original operand: jmp     short loc_233B2
L_044C: ; original 0x233ac; IDA unnamed; logical 0x133ac; file 0x1471c
    jmp SHORT L_0452
; Original operand: dec     word ptr ds:0DEh
L_044E: ; original 0x233ae; IDA loc_233AE; logical 0x133ae; file 0x1471e
    dec WORD PTR ds:[0DEh]
; Original operand: mov     dh, 49h ; 'I'
L_0452: ; original 0x233b2; IDA loc_233B2; logical 0x133b2; file 0x14722
    mov dh, 049h
; Original operand: mov     dl, ds:0D6h
L_0454: ; original 0x233b4; IDA unnamed; logical 0x133b4; file 0x14724
    mov dl, BYTE PTR ds:[0D6h]
; Original operand: mov     ax, ds:0D7h
L_0458: ; original 0x233b8; IDA unnamed; logical 0x133b8; file 0x14728
    mov ax, WORD PTR ds:[0D7h]
; Original operand: mov     cx, ds:0D9h
L_045B: ; original 0x233bb; IDA unnamed; logical 0x133bb; file 0x1472b
    mov cx, WORD PTR ds:[0D9h]
; Original operand: call    sub_231DB
L_045F: ; original 0x233bf; IDA unnamed; logical 0x133bf; file 0x1472f
    call L_027B
; Original operand: dec     byte ptr ds:0DBh
L_0462: ; original 0x233c2; IDA unnamed; logical 0x133c2; file 0x14732
    dec BYTE PTR ds:[0DBh]
; Original operand: inc     byte ptr ds:0D6h
L_0466: ; original 0x233c6; IDA unnamed; logical 0x133c6; file 0x14736
    inc BYTE PTR ds:[0D6h]
; Original operand: mov     word ptr ds:0D7h, 0
L_046A: ; original 0x233ca; IDA unnamed; logical 0x133ca; file 0x1473a
    mov WORD PTR ds:[0D7h], 00h
; Original operand: mov     cx, ds:0D9h
L_0470: ; original 0x233d0; IDA unnamed; logical 0x133d0; file 0x14740
    mov cx, WORD PTR ds:[0D9h]
; Original operand: mov     dx, ds:30h
L_0474: ; original 0x233d4; IDA unnamed; logical 0x133d4; file 0x14744
    mov dx, WORD PTR ds:[030h]
; Original operand: add     dl, 0Ch
L_0478: ; original 0x233d8; IDA unnamed; logical 0x133d8; file 0x14748
    add dl, 0Ch
; Original operand: mov     al, ds:0B3h
L_047B: ; original 0x233db; IDA unnamed; logical 0x133db; file 0x1474b
    mov al, BYTE PTR ds:[0B3h]
; Original operand: call    sub_23077
L_047E: ; original 0x233de; IDA unnamed; logical 0x133de; file 0x1474e
    call L_0117
; Original operand: mov     al, cl
L_0481: ; original 0x233e1; IDA unnamed; logical 0x133e1; file 0x14751
    mov al, cl
; Original operand: call    sub_23077
L_0483: ; original 0x233e3; IDA unnamed; logical 0x133e3; file 0x14753
    call L_0117
; Original operand: mov     al, ch
L_0486: ; original 0x233e6; IDA unnamed; logical 0x133e6; file 0x14756
    mov al, ch
; Original operand: call    sub_23077
L_0488: ; original 0x233e8; IDA unnamed; logical 0x133e8; file 0x14758
    call L_0117
; Original operand: retn
L_048B: ; original 0x233eb; IDA unnamed; logical 0x133eb; file 0x1475b
    ret 
; Original operand: mov     al, 5
L_048C: ; original 0x233ec; IDA sub_233EC; logical 0x133ec; file 0x1475c
    mov al, 05h
; Original operand: out     0Ah, al; DMA controller, 8237A-5.
L_048E: ; original 0x233ee; IDA unnamed; logical 0x133ee; file 0x1475e
    out 0Ah, al
; Original operand: call    sub_2327D
L_0490: ; original 0x233f0; IDA unnamed; logical 0x133f0; file 0x14760
    call L_031D
; Original operand: cmp     byte ptr ds:0B2h, 2
L_0493: ; original 0x233f3; IDA unnamed; logical 0x133f3; file 0x14763
    cmp BYTE PTR ds:[0B2h], 02h
; Original operand: jnz     short loc_23440
L_0498: ; original 0x233f8; IDA unnamed; logical 0x133f8; file 0x14768
    jnz SHORT L_04E0
; Original operand: push    ds
L_049A: ; original 0x233fa; IDA unnamed; logical 0x133fa; file 0x1476a
    push ds
; Original operand: lds     bx, ds:0C6h
L_049B: ; original 0x233fb; IDA unnamed; logical 0x133fb; file 0x1476b
    lds bx, DWORD PTR ds:[0C6h]
; Original operand: in      al, 3; DMA controller, 8237A-5.
L_049F: ; original 0x233ff; IDA unnamed; logical 0x133ff; file 0x1476f
    in al, 03h
; Original operand: mov     ah, al
L_04A1: ; original 0x23401; IDA unnamed; logical 0x13401; file 0x14771
    mov ah, al
; Original operand: in      al, 3; DMA controller, 8237A-5.
L_04A3: ; original 0x23403; IDA unnamed; logical 0x13403; file 0x14773
    in al, 03h
; Original operand: xchg    ah, al
L_04A5: ; original 0x23405; IDA unnamed; logical 0x13405; file 0x14775
    xchg al, ah
; Original operand: inc     ax
L_04A7: ; original 0x23407; IDA unnamed; logical 0x13407; file 0x14777
    inc ax
; Original operand: sub     [bx+1], ax
L_04A8: ; original 0x23408; IDA unnamed; logical 0x13408; file 0x14778
    sub WORD PTR [bx+DISP_0001], ax
; Original operand: sbb     byte ptr [bx+3], 0
L_04AC: ; original 0x2340c; IDA unnamed; logical 0x1340c; file 0x1477c
    sbb BYTE PTR [bx+DISP_0003], 00h
; Original operand: add     word ptr [bx+1], 2
L_04B1: ; original 0x23411; IDA unnamed; logical 0x13411; file 0x14781
    add WORD PTR [bx+DISP_0001], 02h
; Original operand: adc     byte ptr [bx+3], 0
L_04B6: ; original 0x23416; IDA unnamed; logical 0x13416; file 0x14786
    adc BYTE PTR [bx+DISP_0003], 00h
; Original operand: mov     dx, ds
L_04BB: ; original 0x2341b; IDA unnamed; logical 0x1341b; file 0x1478b
    mov dx, ds
; Original operand: mov     ax, bx
L_04BD: ; original 0x2341d; IDA unnamed; logical 0x1341d; file 0x1478d
    mov ax, bx
; Original operand: call    sub_23204
L_04BF: ; original 0x2341f; IDA unnamed; logical 0x1341f; file 0x1478f
    call L_02A4
; Original operand: add     ax, 4
L_04C2: ; original 0x23422; IDA unnamed; logical 0x13422; file 0x14792
    add ax, IMM_0004
; Original operand: adc     dx, 0
L_04C5: ; original 0x23425; IDA unnamed; logical 0x13425; file 0x14795
    adc dx, 00h
; Original operand: add     ax, [bx+1]
L_04C8: ; original 0x23428; IDA unnamed; logical 0x13428; file 0x14798
    add ax, WORD PTR [bx+DISP_0001]
; Original operand: adc     dl, [bx+3]
L_04CC: ; original 0x2342c; IDA unnamed; logical 0x1342c; file 0x1479c
    adc dl, BYTE PTR [bx+DISP_0003]
; Original operand: ror     dx, 1
L_04D0: ; original 0x23430; IDA unnamed; logical 0x13430; file 0x147a0
    ror dx, 01h
; Original operand: ror     dx, 1
L_04D2: ; original 0x23432; IDA unnamed; logical 0x13432; file 0x147a2
    ror dx, 01h
; Original operand: ror     dx, 1
L_04D4: ; original 0x23434; IDA unnamed; logical 0x13434; file 0x147a4
    ror dx, 01h
; Original operand: ror     dx, 1
L_04D6: ; original 0x23436; IDA unnamed; logical 0x13436; file 0x147a6
    ror dx, 01h
; Original operand: mov     ds, dx
L_04D8: ; original 0x23438; IDA unnamed; logical 0x13438; file 0x147a8
    mov ds, dx
; Original operand: mov     bx, ax
L_04DA: ; original 0x2343a; IDA unnamed; logical 0x1343a; file 0x147aa
    mov bx, ax
; Original operand: mov     byte ptr [bx], 0
L_04DC: ; original 0x2343c; IDA unnamed; logical 0x1343c; file 0x147ac
    mov BYTE PTR [bx], 00h
; Original operand: pop     ds
L_04DF: ; original 0x2343f; IDA unnamed; logical 0x1343f; file 0x147af
    pop ds
; Original operand: sub     ax, ax
L_04E0: ; original 0x23440; IDA loc_23440; logical 0x13440; file 0x147b0
    sub ax, ax
; Original operand: mov     ds:0B2h, al
L_04E2: ; original 0x23442; IDA unnamed; logical 0x13442; file 0x147b2
    mov BYTE PTR ds:[0B2h], al
; Original operand: mov     ds:0D4h, ax
L_04E5: ; original 0x23445; IDA unnamed; logical 0x13445; file 0x147b5
    mov WORD PTR ds:[0D4h], ax
; Original operand: mov     ds:0D2h, ax
L_04E8: ; original 0x23448; IDA unnamed; logical 0x13448; file 0x147b8
    mov WORD PTR ds:[0D2h], ax
; Original operand: call    sub_232CA
L_04EB: ; original 0x2344b; IDA unnamed; logical 0x1344b; file 0x147bb
    call L_036A
; Original operand: mov     dx, ds:30h
L_04EE: ; original 0x2344e; IDA unnamed; logical 0x1344e; file 0x147be
    mov dx, WORD PTR ds:[030h]
; Original operand: add     dl, 0Eh
L_04F2: ; original 0x23452; IDA unnamed; logical 0x13452; file 0x147c2
    add dl, 0Eh
; Original operand: in      al, dx
L_04F5: ; original 0x23455; IDA unnamed; logical 0x13455; file 0x147c5
    in al, dx
; Original operand: retn
L_04F6: ; original 0x23456; IDA unnamed; logical 0x13456; file 0x147c6
    ret 
; Original operand: push    ds
L_04F7: ; original 0x23457; IDA unnamed; logical 0x13457; file 0x147c7
    push ds
; Original operand: push    es
L_04F8: ; original 0x23458; IDA unnamed; logical 0x13458; file 0x147c8
    push es
; Original operand: push    ax
L_04F9: ; original 0x23459; IDA unnamed; logical 0x13459; file 0x147c9
    push ax
; Original operand: push    bx
L_04FA: ; original 0x2345a; IDA unnamed; logical 0x1345a; file 0x147ca
    push bx
; Original operand: push    cx
L_04FB: ; original 0x2345b; IDA unnamed; logical 0x1345b; file 0x147cb
    push cx
; Original operand: push    dx
L_04FC: ; original 0x2345c; IDA unnamed; logical 0x1345c; file 0x147cc
    push dx
; Original operand: push    di
L_04FD: ; original 0x2345d; IDA unnamed; logical 0x1345d; file 0x147cd
    push di
; Original operand: push    si
L_04FE: ; original 0x2345e; IDA unnamed; logical 0x1345e; file 0x147ce
    push si
; Original operand: push    bp
L_04FF: ; original 0x2345f; IDA unnamed; logical 0x1345f; file 0x147cf
    push bp
; Original operand: cld
L_0500: ; original 0x23460; IDA unnamed; logical 0x13460; file 0x147d0
    cld 
; Original operand: sti
L_0501: ; original 0x23461; IDA unnamed; logical 0x13461; file 0x147d1
    sti 
; Original operand: mov     ax, cs
L_0502: ; original 0x23462; IDA unnamed; logical 0x13462; file 0x147d2
    mov ax, cs
; Original operand: mov     ds, ax
L_0504: ; original 0x23464; IDA unnamed; logical 0x13464; file 0x147d4
    mov ds, ax
; Original operand: mov     es, ax
L_0506: ; original 0x23466; IDA unnamed; logical 0x13466; file 0x147d6
    mov es, ax
; Original operand: mov     ax, word ptr byte_23012+2Ah
L_0508: ; original 0x23468; IDA unnamed; logical 0x13468; file 0x147d8
    mov ax, WORD PTR ds:[0DCh]
; Original operand: or      ax, word ptr byte_23012+2Ch
L_050B: ; original 0x2346b; IDA unnamed; logical 0x1346b; file 0x147db
    or ax, WORD PTR ds:[0DEh]
; Original operand: jnz     short loc_23476
L_050F: ; original 0x2346f; IDA unnamed; logical 0x1346f; file 0x147df
    jnz SHORT L_0516
; Original operand: call    sub_233EC
L_0511: ; original 0x23471; IDA unnamed; logical 0x13471; file 0x147e1
    call L_048C
; Original operand: jmp     short loc_23479
L_0514: ; original 0x23474; IDA unnamed; logical 0x13474; file 0x147e4
    jmp SHORT L_0519
; Original operand: call    sub_2348F
L_0516: ; original 0x23476; IDA loc_23476; logical 0x13476; file 0x147e6
    call L_052F
; Original operand: mov     dx, word ptr byte_22F90
L_0519: ; original 0x23479; IDA loc_23479; logical 0x13479; file 0x147e9
    mov dx, WORD PTR ds:[030h]
; Original operand: add     dl, 0Eh
L_051D: ; original 0x2347d; IDA unnamed; logical 0x1347d; file 0x147ed
    add dl, 0Eh
; Original operand: in      al, dx
L_0520: ; original 0x23480; IDA unnamed; logical 0x13480; file 0x147f0
    in al, dx
; Original operand: mov     al, 20h ; ' '
L_0521: ; original 0x23481; IDA unnamed; logical 0x13481; file 0x147f1
    mov al, 020h
; Original operand: out     20h, al; Interrupt controller, 8259A.
L_0523: ; original 0x23483; IDA unnamed; logical 0x13483; file 0x147f3
    out 020h, al
; Original operand: pop     bp
L_0525: ; original 0x23485; IDA unnamed; logical 0x13485; file 0x147f5
    pop bp
; Original operand: pop     si
L_0526: ; original 0x23486; IDA unnamed; logical 0x13486; file 0x147f6
    pop si
; Original operand: pop     di
L_0527: ; original 0x23487; IDA unnamed; logical 0x13487; file 0x147f7
    pop di
; Original operand: pop     dx
L_0528: ; original 0x23488; IDA unnamed; logical 0x13488; file 0x147f8
    pop dx
; Original operand: pop     cx
L_0529: ; original 0x23489; IDA unnamed; logical 0x13489; file 0x147f9
    pop cx
; Original operand: pop     bx
L_052A: ; original 0x2348a; IDA unnamed; logical 0x1348a; file 0x147fa
    pop bx
; Original operand: pop     ax
L_052B: ; original 0x2348b; IDA unnamed; logical 0x1348b; file 0x147fb
    pop ax
; Original operand: pop     es
L_052C: ; original 0x2348c; IDA unnamed; logical 0x1348c; file 0x147fc
    pop es
; Original operand: pop     ds
L_052D: ; original 0x2348d; IDA unnamed; logical 0x1348d; file 0x147fd
    pop ds
; Original operand: iret
L_052E: ; original 0x2348e; IDA unnamed; logical 0x1348e; file 0x147fe
    iret 
; Original operand: mov     cx, 0FFFFh
L_052F: ; original 0x2348f; IDA sub_2348F; logical 0x1348f; file 0x147ff
    mov cx, 0FFFFh
; Original operand: cmp     byte ptr ds:0DBh, 0
L_0532: ; original 0x23492; IDA unnamed; logical 0x13492; file 0x14802
    cmp BYTE PTR ds:[0DBh], 00h
; Original operand: jnz     short loc_234A1
L_0537: ; original 0x23497; IDA unnamed; logical 0x13497; file 0x14807
    jnz SHORT L_0541
; Original operand: inc     byte ptr ds:0DBh
L_0539: ; original 0x23499; IDA unnamed; logical 0x13499; file 0x14809
    inc BYTE PTR ds:[0DBh]
; Original operand: mov     cx, ds:0E0h
L_053D: ; original 0x2349d; IDA unnamed; logical 0x1349d; file 0x1480d
    mov cx, WORD PTR ds:[0E0h]
; Original operand: sub     cx, ds:0D7h
L_0541: ; original 0x234a1; IDA loc_234A1; logical 0x134a1; file 0x14811
    sub cx, WORD PTR ds:[0D7h]
; Original operand: mov     ds:0D9h, cx
L_0545: ; original 0x234a5; IDA unnamed; logical 0x134a5; file 0x14815
    mov WORD PTR ds:[0D9h], cx
; Original operand: inc     cx
L_0549: ; original 0x234a9; IDA unnamed; logical 0x134a9; file 0x14819
    inc cx
; Original operand: jz      short loc_234C8
L_054A: ; original 0x234aa; IDA unnamed; logical 0x134aa; file 0x1481a
    jz SHORT L_0568
; Original operand: sub     ds:0DCh, cx
L_054C: ; original 0x234ac; IDA unnamed; logical 0x134ac; file 0x1481c
    sub WORD PTR ds:[0DCh], cx
; Original operand: sbb     word ptr ds:0DEh, 0
L_0550: ; original 0x234b0; IDA unnamed; logical 0x134b0; file 0x14820
    sbb WORD PTR ds:[0DEh], 00h
; Original operand: push    ds
L_0555: ; original 0x234b5; IDA unnamed; logical 0x134b5; file 0x14825
    push ds
; Original operand: push    bx
L_0556: ; original 0x234b6; IDA unnamed; logical 0x134b6; file 0x14826
    push bx
; Original operand: lds     bx, ds:0C6h
L_0557: ; original 0x234b7; IDA unnamed; logical 0x134b7; file 0x14827
    lds bx, DWORD PTR ds:[0C6h]
; Original operand: add     [bx+1], cx
L_055B: ; original 0x234bb; IDA unnamed; logical 0x134bb; file 0x1482b
    add WORD PTR [bx+DISP_0001], cx
; Original operand: adc     byte ptr [bx+3], 0
L_055F: ; original 0x234bf; IDA unnamed; logical 0x134bf; file 0x1482f
    adc BYTE PTR [bx+DISP_0003], 00h
; Original operand: pop     bx
L_0564: ; original 0x234c4; IDA unnamed; logical 0x134c4; file 0x14834
    pop bx
; Original operand: pop     ds
L_0565: ; original 0x234c5; IDA unnamed; logical 0x134c5; file 0x14835
    pop ds
; Original operand: jmp     short loc_234D8
L_0566: ; original 0x234c6; IDA unnamed; logical 0x134c6; file 0x14836
    jmp SHORT L_0578
; Original operand: dec     word ptr ds:0DEh
L_0568: ; original 0x234c8; IDA loc_234C8; logical 0x134c8; file 0x14838
    dec WORD PTR ds:[0DEh]
; Original operand: push    ds
L_056C: ; original 0x234cc; IDA unnamed; logical 0x134cc; file 0x1483c
    push ds
; Original operand: push    bx
L_056D: ; original 0x234cd; IDA unnamed; logical 0x134cd; file 0x1483d
    push bx
; Original operand: lds     bx, ds:0C6h
L_056E: ; original 0x234ce; IDA unnamed; logical 0x134ce; file 0x1483e
    lds bx, DWORD PTR ds:[0C6h]
; Original operand: inc     byte ptr [bx+3]
L_0572: ; original 0x234d2; IDA unnamed; logical 0x134d2; file 0x14842
    inc BYTE PTR [bx+DISP_0003]
; Original operand: pop     bx
L_0576: ; original 0x234d6; IDA unnamed; logical 0x134d6; file 0x14846
    pop bx
; Original operand: pop     ds
L_0577: ; original 0x234d7; IDA unnamed; logical 0x134d7; file 0x14847
    pop ds
; Original operand: mov     dh, 45h ; 'E'
L_0578: ; original 0x234d8; IDA loc_234D8; logical 0x134d8; file 0x14848
    mov dh, 045h
; Original operand: mov     dl, ds:0D6h
L_057A: ; original 0x234da; IDA unnamed; logical 0x134da; file 0x1484a
    mov dl, BYTE PTR ds:[0D6h]
; Original operand: mov     ax, ds:0D7h
L_057E: ; original 0x234de; IDA unnamed; logical 0x134de; file 0x1484e
    mov ax, WORD PTR ds:[0D7h]
; Original operand: mov     cx, ds:0D9h
L_0581: ; original 0x234e1; IDA unnamed; logical 0x134e1; file 0x14851
    mov cx, WORD PTR ds:[0D9h]
; Original operand: call    sub_231DB
L_0585: ; original 0x234e5; IDA unnamed; logical 0x134e5; file 0x14855
    call L_027B
; Original operand: dec     byte ptr ds:0DBh
L_0588: ; original 0x234e8; IDA unnamed; logical 0x134e8; file 0x14858
    dec BYTE PTR ds:[0DBh]
; Original operand: inc     byte ptr ds:0D6h
L_058C: ; original 0x234ec; IDA unnamed; logical 0x134ec; file 0x1485c
    inc BYTE PTR ds:[0D6h]
; Original operand: mov     word ptr ds:0D7h, 0
L_0590: ; original 0x234f0; IDA unnamed; logical 0x134f0; file 0x14860
    mov WORD PTR ds:[0D7h], 00h
; Original operand: mov     cx, ds:0D9h
L_0596: ; original 0x234f6; IDA unnamed; logical 0x134f6; file 0x14866
    mov cx, WORD PTR ds:[0D9h]
; Original operand: mov     dx, ds:30h
L_059A: ; original 0x234fa; IDA unnamed; logical 0x134fa; file 0x1486a
    mov dx, WORD PTR ds:[030h]
; Original operand: add     dl, 0Ch
L_059E: ; original 0x234fe; IDA unnamed; logical 0x134fe; file 0x1486e
    add dl, 0Ch
; Original operand: mov     al, 24h ; '$'
L_05A1: ; original 0x23501; IDA unnamed; logical 0x13501; file 0x14871
    mov al, 024h
; Original operand: call    sub_23077
L_05A3: ; original 0x23503; IDA unnamed; logical 0x13503; file 0x14873
    call L_0117
; Original operand: mov     al, cl
L_05A6: ; original 0x23506; IDA unnamed; logical 0x13506; file 0x14876
    mov al, cl
; Original operand: call    sub_23077
L_05A8: ; original 0x23508; IDA unnamed; logical 0x13508; file 0x14878
    call L_0117
; Original operand: mov     al, ch
L_05AB: ; original 0x2350b; IDA unnamed; logical 0x1350b; file 0x1487b
    mov al, ch
; Original operand: call    sub_23077
L_05AD: ; original 0x2350d; IDA unnamed; logical 0x1350d; file 0x1487d
    call L_0117
; Original operand: retn
L_05B0: ; original 0x23510; IDA unnamed; logical 0x13510; file 0x14880
    ret 
; Original operand: cmp     word ptr ds:0BCh, 0
L_05B1: ; original 0x23511; IDA sub_23511; logical 0x13511; file 0x14881
    cmp WORD PTR ds:[0BCh], 00h
; Original operand: jnz     short loc_2351F
L_05B6: ; original 0x23516; IDA unnamed; logical 0x13516; file 0x14886
    jnz SHORT L_05BF
; Original operand: cmp     word ptr ds:0BAh, 0
L_05B8: ; original 0x23518; IDA unnamed; logical 0x13518; file 0x14888
    cmp WORD PTR ds:[0BAh], 00h
; Original operand: jz      short loc_2352B
L_05BD: ; original 0x2351d; IDA unnamed; logical 0x1351d; file 0x1488d
    jz SHORT L_05CB
; Original operand: push    es
L_05BF: ; original 0x2351f; IDA loc_2351F; logical 0x1351f; file 0x1488f
    push es
; Original operand: les     bx, ds:0CAh
L_05C0: ; original 0x23520; IDA unnamed; logical 0x13520; file 0x14890
    les bx, DWORD PTR ds:[0CAh]
; Original operand: call    dword ptr ds:0BAh
L_05C4: ; original 0x23524; IDA unnamed; logical 0x13524; file 0x14894
    call DWORD PTR ds:[0BAh]
; Original operand: pop     es
L_05C8: ; original 0x23528; IDA unnamed; logical 0x13528; file 0x14898
    pop es
; Original operand: jb      short loc_23540
L_05C9: ; original 0x23529; IDA unnamed; logical 0x13529; file 0x14899
    jb SHORT L_05E0
; Original operand: call    sub_232D5
L_05CB: ; original 0x2352b; IDA loc_2352B; logical 0x1352b; file 0x1489b
    call L_0375
; Original operand: cbw
L_05CE: ; original 0x2352e; IDA unnamed; logical 0x1352e; file 0x1489e
    cbw 
; Original operand: cmp     ax, 8
L_05CF: ; original 0x2352f; IDA unnamed; logical 0x1352f; file 0x1489f
    cmp ax, IMM_0008
; Original operand: jnb     short loc_23540
L_05D2: ; original 0x23532; IDA unnamed; logical 0x13532; file 0x148a2
    jnb SHORT L_05E0
; Original operand: mov     bx, ax
L_05D4: ; original 0x23534; IDA unnamed; logical 0x13534; file 0x148a4
    mov bx, ax
; Original operand: shl     bx, 1
L_05D6: ; original 0x23536; IDA unnamed; logical 0x13536; file 0x148a6
    shl bx, 01h
; Original operand: call    word ptr [bx+0A1h]
L_05D8: ; original 0x23538; IDA unnamed; logical 0x13538; file 0x148a8
    call WORD PTR [bx+0A1h]
; Original operand: jb      short sub_23511
L_05DC: ; original 0x2353c; IDA unnamed; logical 0x1353c; file 0x148ac
    jb SHORT L_05B1
; Original operand: jmp     short locret_23545
L_05DE: ; original 0x2353e; IDA unnamed; logical 0x1353e; file 0x148ae
    jmp SHORT L_05E5
; Original operand: call    sub_23546
L_05E0: ; original 0x23540; IDA loc_23540; logical 0x13540; file 0x148b0
    call L_05E6
; Original operand: jmp     short sub_23511
L_05E3: ; original 0x23543; IDA unnamed; logical 0x13543; file 0x148b3
    jmp SHORT L_05B1
; Original operand: retn
L_05E5: ; original 0x23545; IDA locret_23545; logical 0x13545; file 0x148b5
    ret 
; Original operand: push    es
L_05E6: ; original 0x23546; IDA sub_23546; logical 0x13546; file 0x148b6
    push es
; Original operand: push    ax
L_05E7: ; original 0x23547; IDA unnamed; logical 0x13547; file 0x148b7
    push ax
; Original operand: push    bx
L_05E8: ; original 0x23548; IDA unnamed; logical 0x13548; file 0x148b8
    push bx
; Original operand: push    dx
L_05E9: ; original 0x23549; IDA unnamed; logical 0x13549; file 0x148b9
    push dx
; Original operand: les     bx, ds:0CAh
L_05EA: ; original 0x2354a; IDA unnamed; logical 0x1354a; file 0x148ba
    les bx, DWORD PTR ds:[0CAh]
; Original operand: mov     ax, es:[bx+1]
L_05EE: ; original 0x2354e; IDA unnamed; logical 0x1354e; file 0x148be
    mov ax, WORD PTR es:[bx+01h]
; Original operand: mov     dl, es:[bx+3]
L_05F2: ; original 0x23552; IDA unnamed; logical 0x13552; file 0x148c2
    mov dl, BYTE PTR es:[bx+03h]
; Original operand: sub     dh, dh
L_05F6: ; original 0x23556; IDA unnamed; logical 0x13556; file 0x148c6
    sub dh, dh
; Original operand: add     ax, 4
L_05F8: ; original 0x23558; IDA unnamed; logical 0x13558; file 0x148c8
    add ax, IMM_0004
; Original operand: adc     dx, 0
L_05FB: ; original 0x2355b; IDA unnamed; logical 0x1355b; file 0x148cb
    adc dx, 00h
; Original operand: add     ax, ds:0CAh
L_05FE: ; original 0x2355e; IDA unnamed; logical 0x1355e; file 0x148ce
    add ax, WORD PTR ds:[0CAh]
; Original operand: adc     dx, 0
L_0602: ; original 0x23562; IDA unnamed; logical 0x13562; file 0x148d2
    adc dx, 00h
; Original operand: ror     dx, 1
L_0605: ; original 0x23565; IDA unnamed; logical 0x13565; file 0x148d5
    ror dx, 01h
; Original operand: ror     dx, 1
L_0607: ; original 0x23567; IDA unnamed; logical 0x13567; file 0x148d7
    ror dx, 01h
; Original operand: ror     dx, 1
L_0609: ; original 0x23569; IDA unnamed; logical 0x13569; file 0x148d9
    ror dx, 01h
; Original operand: ror     dx, 1
L_060B: ; original 0x2356b; IDA unnamed; logical 0x1356b; file 0x148db
    ror dx, 01h
; Original operand: add     dx, ds:0CCh
L_060D: ; original 0x2356d; IDA unnamed; logical 0x1356d; file 0x148dd
    add dx, WORD PTR ds:[0CCh]
; Original operand: mov     bx, ax
L_0611: ; original 0x23571; IDA unnamed; logical 0x13571; file 0x148e1
    mov bx, ax
; Original operand: shr     bx, 1
L_0613: ; original 0x23573; IDA unnamed; logical 0x13573; file 0x148e3
    shr bx, 01h
; Original operand: shr     bx, 1
L_0615: ; original 0x23575; IDA unnamed; logical 0x13575; file 0x148e5
    shr bx, 01h
; Original operand: shr     bx, 1
L_0617: ; original 0x23577; IDA unnamed; logical 0x13577; file 0x148e7
    shr bx, 01h
; Original operand: shr     bx, 1
L_0619: ; original 0x23579; IDA unnamed; logical 0x13579; file 0x148e9
    shr bx, 01h
; Original operand: add     dx, bx
L_061B: ; original 0x2357b; IDA unnamed; logical 0x1357b; file 0x148eb
    add dx, bx
; Original operand: and     ax, 0Fh
L_061D: ; original 0x2357d; IDA unnamed; logical 0x1357d; file 0x148ed
    and ax, IMM_000F
; Original operand: mov     ds:0CCh, dx
L_0620: ; original 0x23580; IDA unnamed; logical 0x13580; file 0x148f0
    mov WORD PTR ds:[0CCh], dx
; Original operand: mov     ds:0CAh, ax
L_0624: ; original 0x23584; IDA unnamed; logical 0x13584; file 0x148f4
    mov WORD PTR ds:[0CAh], ax
; Original operand: pop     dx
L_0627: ; original 0x23587; IDA unnamed; logical 0x13587; file 0x148f7
    pop dx
; Original operand: pop     bx
L_0628: ; original 0x23588; IDA unnamed; logical 0x13588; file 0x148f8
    pop bx
; Original operand: pop     ax
L_0629: ; original 0x23589; IDA unnamed; logical 0x13589; file 0x148f9
    pop ax
; Original operand: pop     es
L_062A: ; original 0x2358a; IDA unnamed; logical 0x1358a; file 0x148fa
    pop es
; Original operand: retn
L_062B: ; original 0x2358b; IDA unnamed; logical 0x1358b; file 0x148fb
    ret 
; Original operand: push    ax
L_062C: ; original 0x2358c; IDA sub_2358C; logical 0x1358c; file 0x148fc
    push ax
; Original operand: shr     ax, 1
L_062D: ; original 0x2358d; IDA unnamed; logical 0x1358d; file 0x148fd
    shr ax, 01h
; Original operand: shr     ax, 1
L_062F: ; original 0x2358f; IDA unnamed; logical 0x1358f; file 0x148ff
    shr ax, 01h
; Original operand: shr     ax, 1
L_0631: ; original 0x23591; IDA unnamed; logical 0x13591; file 0x14901
    shr ax, 01h
; Original operand: shr     ax, 1
L_0633: ; original 0x23593; IDA unnamed; logical 0x13593; file 0x14903
    shr ax, 01h
; Original operand: add     dx, ax
L_0635: ; original 0x23595; IDA unnamed; logical 0x13595; file 0x14905
    add dx, ax
; Original operand: pop     ax
L_0637: ; original 0x23597; IDA unnamed; logical 0x13597; file 0x14907
    pop ax
; Original operand: and     ax, 0Fh
L_0638: ; original 0x23598; IDA unnamed; logical 0x13598; file 0x14908
    and ax, IMM_000F
; Original operand: retn
L_063B: ; original 0x2359b; IDA unnamed; logical 0x1359b; file 0x1490b
    ret 
; Original operand: mov     byte ptr ds:0E2h, 1
L_063C: ; original 0x2359c; IDA unnamed; logical 0x1359c; file 0x1490c
    mov BYTE PTR ds:[0E2h], 01h
; Original operand: clc
L_0641: ; original 0x235a1; IDA unnamed; logical 0x135a1; file 0x14911
    clc 
; Original operand: retn
L_0642: ; original 0x235a2; IDA unnamed; logical 0x135a2; file 0x14912
    ret 
; Original operand: push    es
L_0643: ; original 0x235a3; IDA unnamed; logical 0x135a3; file 0x14913
    push es
; Original operand: push    di
L_0644: ; original 0x235a4; IDA unnamed; logical 0x135a4; file 0x14914
    push di
; Original operand: les     di, ds:0CAh
L_0645: ; original 0x235a5; IDA unnamed; logical 0x135a5; file 0x14915
    les di, DWORD PTR ds:[0CAh]
; Original operand: mov     dx, ds:30h
L_0649: ; original 0x235a9; IDA unnamed; logical 0x135a9; file 0x14919
    mov dx, WORD PTR ds:[030h]
; Original operand: add     dl, 0Ch
L_064D: ; original 0x235ad; IDA unnamed; logical 0x135ad; file 0x1491d
    add dl, 0Ch
; Original operand: mov     al, 40h ; '@'
L_0650: ; original 0x235b0; IDA unnamed; logical 0x135b0; file 0x14920
    mov al, 040h
; Original operand: call    sub_23077
L_0652: ; original 0x235b2; IDA unnamed; logical 0x135b2; file 0x14922
    call L_0117
; Original operand: mov     al, es:[di+4]
L_0655: ; original 0x235b5; IDA unnamed; logical 0x135b5; file 0x14925
    mov al, BYTE PTR es:[di+04h]
; Original operand: call    sub_23077
L_0659: ; original 0x235b9; IDA unnamed; logical 0x135b9; file 0x14929
    call L_0117
; Original operand: mov     bl, es:[di+5]
L_065C: ; original 0x235bc; IDA unnamed; logical 0x135bc; file 0x1492c
    mov bl, BYTE PTR es:[di+05h]
; Original operand: sub     bh, bh
L_0660: ; original 0x235c0; IDA unnamed; logical 0x135c0; file 0x14930
    sub bh, bh
; Original operand: mov     al, [bx+749h]
L_0662: ; original 0x235c2; IDA unnamed; logical 0x135c2; file 0x14932
    mov al, BYTE PTR [bx+0749h]
; Original operand: mov     ds:0B3h, al
L_0666: ; original 0x235c6; IDA unnamed; logical 0x135c6; file 0x14936
    mov BYTE PTR ds:[0B3h], al
; Original operand: pop     di
L_0669: ; original 0x235c9; IDA unnamed; logical 0x135c9; file 0x14939
    pop di
; Original operand: pop     es
L_066A: ; original 0x235ca; IDA unnamed; logical 0x135ca; file 0x1493a
    pop es
; Original operand: mov     ax, 6
L_066B: ; original 0x235cb; IDA unnamed; logical 0x135cb; file 0x1493b
    mov ax, 06h
; Original operand: call    sub_232E2
L_066E: ; original 0x235ce; IDA unnamed; logical 0x135ce; file 0x1493e
    call L_0382
; Original operand: call    sub_23386
L_0671: ; original 0x235d1; IDA unnamed; logical 0x135d1; file 0x14941
    call L_0426
; Original operand: mov     al, ds:0B3h
L_0674: ; original 0x235d4; IDA unnamed; logical 0x135d4; file 0x14944
    mov al, BYTE PTR ds:[0B3h]
; Original operand: cmp     al, 61h ; 'a'
L_0677: ; original 0x235d7; IDA unnamed; logical 0x135d7; file 0x14947
    cmp al, 061h
; Original operand: jb      short loc_235E6
L_0679: ; original 0x235d9; IDA unnamed; logical 0x135d9; file 0x14949
    jb SHORT L_0686
; Original operand: cmp     al, 67h ; 'g'
L_067B: ; original 0x235db; IDA unnamed; logical 0x135db; file 0x1494b
    cmp al, 067h
; Original operand: ja      short loc_235E6
L_067D: ; original 0x235dd; IDA unnamed; logical 0x135dd; file 0x1494d
    ja SHORT L_0686
; Original operand: or      byte ptr ds:0B3h, 8
L_067F: ; original 0x235df; IDA unnamed; logical 0x135df; file 0x1494f
    or BYTE PTR ds:[0B3h], 08h
; Original operand: jmp     short loc_235EB
L_0684: ; original 0x235e4; IDA unnamed; logical 0x135e4; file 0x14954
    jmp SHORT L_068B
; Original operand: and     byte ptr ds:0B3h, 0FEh
L_0686: ; original 0x235e6; IDA loc_235E6; logical 0x135e6; file 0x14956
    and BYTE PTR ds:[0B3h], 0FEh
; Original operand: clc
L_068B: ; original 0x235eb; IDA loc_235EB; logical 0x135eb; file 0x1495b
    clc 
; Original operand: retn
L_068C: ; original 0x235ec; IDA unnamed; logical 0x135ec; file 0x1495c
    ret 
; Original operand: mov     ax, 4
L_068D: ; original 0x235ed; IDA unnamed; logical 0x135ed; file 0x1495d
    mov ax, 04h
; Original operand: call    sub_232E2
L_0690: ; original 0x235f0; IDA unnamed; logical 0x135f0; file 0x14960
    call L_0382
; Original operand: mov     al, ds:0B3h
L_0693: ; original 0x235f3; IDA unnamed; logical 0x135f3; file 0x14963
    mov al, BYTE PTR ds:[0B3h]
; Original operand: cmp     al, 61h ; 'a'
L_0696: ; original 0x235f6; IDA unnamed; logical 0x135f6; file 0x14966
    cmp al, 061h
; Original operand: jb      short loc_23605
L_0698: ; original 0x235f8; IDA unnamed; logical 0x135f8; file 0x14968
    jb SHORT L_06A5
; Original operand: cmp     al, 67h ; 'g'
L_069A: ; original 0x235fa; IDA unnamed; logical 0x135fa; file 0x1496a
    cmp al, 067h
; Original operand: ja      short loc_23605
L_069C: ; original 0x235fc; IDA unnamed; logical 0x135fc; file 0x1496c
    ja SHORT L_06A5
; Original operand: or      byte ptr ds:0B3h, 8
L_069E: ; original 0x235fe; IDA unnamed; logical 0x135fe; file 0x1496e
    or BYTE PTR ds:[0B3h], 08h
; Original operand: jmp     short loc_2360A
L_06A3: ; original 0x23603; IDA unnamed; logical 0x13603; file 0x14973
    jmp SHORT L_06AA
; Original operand: and     byte ptr ds:0B3h, 0FEh
L_06A5: ; original 0x23605; IDA loc_23605; logical 0x13605; file 0x14975
    and BYTE PTR ds:[0B3h], 0FEh
; Original operand: call    sub_23386
L_06AA: ; original 0x2360a; IDA loc_2360A; logical 0x1360a; file 0x1497a
    call L_0426
; Original operand: clc
L_06AD: ; original 0x2360d; IDA unnamed; logical 0x1360d; file 0x1497d
    clc 
; Original operand: retn
L_06AE: ; original 0x2360e; IDA unnamed; logical 0x1360e; file 0x1497e
    ret 
; Original operand: mov     dx, ds:30h
L_06AF: ; original 0x2360f; IDA unnamed; logical 0x1360f; file 0x1497f
    mov dx, WORD PTR ds:[030h]
; Original operand: add     dl, 0Ch
L_06B3: ; original 0x23613; IDA unnamed; logical 0x13613; file 0x14983
    add dl, 0Ch
; Original operand: mov     al, 40h ; '@'
L_06B6: ; original 0x23616; IDA unnamed; logical 0x13616; file 0x14986
    mov al, 040h
; Original operand: call    sub_23077
L_06B8: ; original 0x23618; IDA unnamed; logical 0x13618; file 0x14988
    call L_0117
; Original operand: push    es
L_06BB: ; original 0x2361b; IDA unnamed; logical 0x1361b; file 0x1498b
    push es
; Original operand: push    bx
L_06BC: ; original 0x2361c; IDA unnamed; logical 0x1361c; file 0x1498c
    push bx
; Original operand: les     bx, ds:0CAh
L_06BD: ; original 0x2361d; IDA unnamed; logical 0x1361d; file 0x1498d
    les bx, DWORD PTR ds:[0CAh]
; Original operand: mov     al, es:[bx+6]
L_06C1: ; original 0x23621; IDA unnamed; logical 0x13621; file 0x14991
    mov al, BYTE PTR es:[bx+06h]
; Original operand: call    sub_23077
L_06C5: ; original 0x23625; IDA unnamed; logical 0x13625; file 0x14995
    call L_0117
; Original operand: mov     ax, es:[bx+4]
L_06C8: ; original 0x23628; IDA unnamed; logical 0x13628; file 0x14998
    mov ax, WORD PTR es:[bx+04h]
; Original operand: pop     bx
L_06CC: ; original 0x2362c; IDA unnamed; logical 0x1362c; file 0x1499c
    pop bx
; Original operand: pop     es
L_06CD: ; original 0x2362d; IDA unnamed; logical 0x1362d; file 0x1499d
    pop es
; Original operand: mov     bx, ax
L_06CE: ; original 0x2362e; IDA unnamed; logical 0x1362e; file 0x1499e
    mov bx, ax
; Original operand: mov     al, 80h
L_06D0: ; original 0x23630; IDA unnamed; logical 0x13630; file 0x149a0
    mov al, 080h
; Original operand: call    sub_23077
L_06D2: ; original 0x23632; IDA unnamed; logical 0x13632; file 0x149a2
    call L_0117
; Original operand: mov     al, bl
L_06D5: ; original 0x23635; IDA unnamed; logical 0x13635; file 0x149a5
    mov al, bl
; Original operand: call    sub_23077
L_06D7: ; original 0x23637; IDA unnamed; logical 0x13637; file 0x149a7
    call L_0117
; Original operand: mov     al, bh
L_06DA: ; original 0x2363a; IDA unnamed; logical 0x1363a; file 0x149aa
    mov al, bh
; Original operand: call    sub_23077
L_06DC: ; original 0x2363c; IDA unnamed; logical 0x1363c; file 0x149ac
    call L_0117
; Original operand: clc
L_06DF: ; original 0x2363f; IDA unnamed; logical 0x1363f; file 0x149af
    clc 
; Original operand: retn
L_06E0: ; original 0x23640; IDA unnamed; logical 0x13640; file 0x149b0
    ret 
; Original operand: push    ds
L_06E1: ; original 0x23641; IDA unnamed; logical 0x13641; file 0x149b1
    push ds
; Original operand: push    bx
L_06E2: ; original 0x23642; IDA unnamed; logical 0x13642; file 0x149b2
    push bx
; Original operand: lds     bx, ds:0CAh
L_06E3: ; original 0x23643; IDA unnamed; logical 0x13643; file 0x149b3
    lds bx, DWORD PTR ds:[0CAh]
; Original operand: mov     ax, [bx+4]
L_06E7: ; original 0x23647; IDA unnamed; logical 0x13647; file 0x149b7
    mov ax, WORD PTR [bx+04h]
; Original operand: pop     bx
L_06EA: ; original 0x2364a; IDA unnamed; logical 0x1364a; file 0x149ba
    pop bx
; Original operand: pop     ds
L_06EB: ; original 0x2364b; IDA unnamed; logical 0x1364b; file 0x149bb
    pop ds
; Original operand: call    sub_232CA
L_06EC: ; original 0x2364c; IDA unnamed; logical 0x1364c; file 0x149bc
    call L_036A
; Original operand: call    sub_23546
L_06EF: ; original 0x2364f; IDA unnamed; logical 0x1364f; file 0x149bf
    call L_05E6
; Original operand: stc
L_06F2: ; original 0x23652; IDA unnamed; logical 0x13652; file 0x149c2
    stc 
; Original operand: retn
L_06F3: ; original 0x23653; IDA unnamed; logical 0x13653; file 0x149c3
    ret 
; Original operand: call    sub_23546
L_06F4: ; original 0x23654; IDA unnamed; logical 0x13654; file 0x149c4
    call L_05E6
; Original operand: stc
L_06F7: ; original 0x23657; IDA unnamed; logical 0x13657; file 0x149c7
    stc 
; Original operand: retn
L_06F8: ; original 0x23658; IDA unnamed; logical 0x13658; file 0x149c8
    ret 
; Original operand: push    ds
L_06F9: ; original 0x23659; IDA unnamed; logical 0x13659; file 0x149c9
    push ds
; Original operand: push    bx
L_06FA: ; original 0x2365a; IDA unnamed; logical 0x1365a; file 0x149ca
    push bx
; Original operand: lds     bx, ds:0CAh
L_06FB: ; original 0x2365b; IDA unnamed; logical 0x1365b; file 0x149cb
    lds bx, DWORD PTR ds:[0CAh]
; Original operand: mov     ax, [bx+4]
L_06FF: ; original 0x2365f; IDA unnamed; logical 0x1365f; file 0x149cf
    mov ax, WORD PTR [bx+04h]
; Original operand: pop     bx
L_0702: ; original 0x23662; IDA unnamed; logical 0x13662; file 0x149d2
    pop bx
; Original operand: pop     ds
L_0703: ; original 0x23663; IDA unnamed; logical 0x13663; file 0x149d3
    pop ds
; Original operand: mov     ds:0D2h, ax
L_0704: ; original 0x23664; IDA unnamed; logical 0x13664; file 0x149d4
    mov WORD PTR ds:[0D2h], ax
; Original operand: call    sub_23546
L_0707: ; original 0x23667; IDA unnamed; logical 0x13667; file 0x149d7
    call L_05E6
; Original operand: mov     ax, ds:0CAh
L_070A: ; original 0x2366a; IDA unnamed; logical 0x1366a; file 0x149da
    mov ax, WORD PTR ds:[0CAh]
; Original operand: mov     ds:0CEh, ax
L_070D: ; original 0x2366d; IDA unnamed; logical 0x1366d; file 0x149dd
    mov WORD PTR ds:[0CEh], ax
; Original operand: mov     ax, ds:0CCh
L_0710: ; original 0x23670; IDA unnamed; logical 0x13670; file 0x149e0
    mov ax, WORD PTR ds:[0CCh]
; Original operand: mov     ds:0D0h, ax
L_0713: ; original 0x23673; IDA unnamed; logical 0x13673; file 0x149e3
    mov WORD PTR ds:[0D0h], ax
; Original operand: mov     word ptr ds:0D4h, 1
L_0716: ; original 0x23676; IDA unnamed; logical 0x13676; file 0x149e6
    mov WORD PTR ds:[0D4h], 01h
; Original operand: stc
L_071C: ; original 0x2367c; IDA unnamed; logical 0x1367c; file 0x149ec
    stc 
; Original operand: retn
L_071D: ; original 0x2367d; IDA unnamed; logical 0x1367d; file 0x149ed
    ret 
; Original operand: cmp     word ptr ds:0D2h, 0
L_071E: ; original 0x2367e; IDA unnamed; logical 0x1367e; file 0x149ee
    cmp WORD PTR ds:[0D2h], 00h
; Original operand: jz      short loc_2369E
L_0723: ; original 0x23683; IDA unnamed; logical 0x13683; file 0x149f3
    jz SHORT L_073E
; Original operand: mov     ax, ds:0CEh
L_0725: ; original 0x23685; IDA unnamed; logical 0x13685; file 0x149f5
    mov ax, WORD PTR ds:[0CEh]
; Original operand: mov     ds:0CAh, ax
L_0728: ; original 0x23688; IDA unnamed; logical 0x13688; file 0x149f8
    mov WORD PTR ds:[0CAh], ax
; Original operand: mov     ax, ds:0D0h
L_072B: ; original 0x2368b; IDA unnamed; logical 0x1368b; file 0x149fb
    mov ax, WORD PTR ds:[0D0h]
; Original operand: mov     ds:0CCh, ax
L_072E: ; original 0x2368e; IDA unnamed; logical 0x1368e; file 0x149fe
    mov WORD PTR ds:[0CCh], ax
; Original operand: cmp     word ptr ds:0D2h, 0FFFFh
L_0731: ; original 0x23691; IDA unnamed; logical 0x13691; file 0x14a01
    cmp WORD PTR ds:[0D2h], 0FFFFh
; Original operand: jz      short loc_236A7
L_0736: ; original 0x23696; IDA unnamed; logical 0x13696; file 0x14a06
    jz SHORT L_0747
; Original operand: dec     word ptr ds:0D2h
L_0738: ; original 0x23698; IDA unnamed; logical 0x13698; file 0x14a08
    dec WORD PTR ds:[0D2h]
; Original operand: jmp     short loc_236A7
L_073C: ; original 0x2369c; IDA unnamed; logical 0x1369c; file 0x14a0c
    jmp SHORT L_0747
; Original operand: mov     word ptr ds:0D4h, 0
L_073E: ; original 0x2369e; IDA loc_2369E; logical 0x1369e; file 0x14a0e
    mov WORD PTR ds:[0D4h], 00h
; Original operand: call    sub_23546
L_0744: ; original 0x236a4; IDA unnamed; logical 0x136a4; file 0x14a14
    call L_05E6
; Original operand: stc
L_0747: ; original 0x236a7; IDA loc_236A7; logical 0x136a7; file 0x14a17
    stc 
; Original operand: retn
L_0748: ; original 0x236a8; IDA unnamed; logical 0x136a8; file 0x14a18
    ret 
; Unknown data skipped: 0x236a9..0x236b4
ORG 0754h
; Original operand: pushf
L_0754: ; original 0x236b4; IDA sub_236B4; logical 0x136b4; file 0x14a24
    pushf 
; Original operand: push    ds
L_0755: ; original 0x236b5; IDA unnamed; logical 0x136b5; file 0x14a25
    push ds
; Original operand: push    es
L_0756: ; original 0x236b6; IDA unnamed; logical 0x136b6; file 0x14a26
    push es
; Original operand: push    ax
L_0757: ; original 0x236b7; IDA unnamed; logical 0x136b7; file 0x14a27
    push ax
; Original operand: push    bx
L_0758: ; original 0x236b8; IDA unnamed; logical 0x136b8; file 0x14a28
    push bx
; Original operand: push    cx
L_0759: ; original 0x236b9; IDA unnamed; logical 0x136b9; file 0x14a29
    push cx
; Original operand: push    dx
L_075A: ; original 0x236ba; IDA unnamed; logical 0x136ba; file 0x14a2a
    push dx
; Original operand: push    di
L_075B: ; original 0x236bb; IDA unnamed; logical 0x136bb; file 0x14a2b
    push di
; Original operand: push    si
L_075C: ; original 0x236bc; IDA unnamed; logical 0x136bc; file 0x14a2c
    push si
; Original operand: push    bp
L_075D: ; original 0x236bd; IDA unnamed; logical 0x136bd; file 0x14a2d
    push bp
; Original operand: mov     bp, sp
L_075E: ; original 0x236be; IDA unnamed; logical 0x136be; file 0x14a2e
    mov bp, sp
; Original operand: push    ax
L_0760: ; original 0x236c0; IDA unnamed; logical 0x136c0; file 0x14a30
    push ax
; Original operand: mov     ax, cs
L_0761: ; original 0x236c1; IDA unnamed; logical 0x136c1; file 0x14a31
    mov ax, cs
; Original operand: mov     ds, ax
L_0763: ; original 0x236c3; IDA unnamed; logical 0x136c3; file 0x14a33
    mov ds, ax
; Original operand: mov     es, ax
L_0765: ; original 0x236c5; IDA unnamed; logical 0x136c5; file 0x14a35
    mov es, ax
; Original operand: pop     ax
L_0767: ; original 0x236c7; IDA unnamed; logical 0x136c7; file 0x14a37
    pop ax
; Original operand: cld
L_0768: ; original 0x236c8; IDA unnamed; logical 0x136c8; file 0x14a38
    cld 
; Original operand: cmp     bx, 0Eh
L_0769: ; original 0x236c9; IDA unnamed; logical 0x136c9; file 0x14a39
    cmp bx, 0Eh
; Original operand: jnb     short loc_236E5
L_076C: ; original 0x236cc; IDA unnamed; logical 0x136cc; file 0x14a3c
    jnb SHORT L_0785
; Original operand: cmp     bx, 4
L_076E: ; original 0x236ce; IDA unnamed; logical 0x136ce; file 0x14a3e
    cmp bx, 04h
; Original operand: jb      short loc_236DA
L_0771: ; original 0x236d1; IDA unnamed; logical 0x136d1; file 0x14a41
    jb SHORT L_077A
; Original operand: cmp     byte_23011, 0
L_0773: ; original 0x236d3; IDA unnamed; logical 0x136d3; file 0x14a43
    cmp BYTE PTR ds:[0B1h], 00h
; Original operand: jz      short loc_236E5
L_0778: ; original 0x236d8; IDA unnamed; logical 0x136d8; file 0x14a48
    jz SHORT L_0785
; Original operand: shl     bx, 1
L_077A: ; original 0x236da; IDA loc_236DA; logical 0x136da; file 0x14a4a
    shl bx, 01h
; Original operand: call    funcs_236DC[bx]
L_077C: ; original 0x236dc; IDA unnamed; logical 0x136dc; file 0x14a4c
    call WORD PTR [bx+085h]
; Original operand: mov     [bp+0Ch], ax
L_0780: ; original 0x236e0; IDA unnamed; logical 0x136e0; file 0x14a50
    mov WORD PTR [bp+0Ch], ax
; Original operand: jmp     short loc_236EA
L_0783: ; original 0x236e3; IDA unnamed; logical 0x136e3; file 0x14a53
    jmp SHORT L_078A
; Original operand: mov     word ptr [bp+0Ch], 0FFFFh
L_0785: ; original 0x236e5; IDA loc_236E5; logical 0x136e5; file 0x14a55
    mov WORD PTR [bp+0Ch], 0FFFFh
; Original operand: pop     bp
L_078A: ; original 0x236ea; IDA loc_236EA; logical 0x136ea; file 0x14a5a
    pop bp
; Original operand: pop     si
L_078B: ; original 0x236eb; IDA unnamed; logical 0x136eb; file 0x14a5b
    pop si
; Original operand: pop     di
L_078C: ; original 0x236ec; IDA unnamed; logical 0x136ec; file 0x14a5c
    pop di
; Original operand: pop     dx
L_078D: ; original 0x236ed; IDA unnamed; logical 0x136ed; file 0x14a5d
    pop dx
; Original operand: pop     cx
L_078E: ; original 0x236ee; IDA unnamed; logical 0x136ee; file 0x14a5e
    pop cx
; Original operand: pop     bx
L_078F: ; original 0x236ef; IDA unnamed; logical 0x136ef; file 0x14a5f
    pop bx
; Original operand: pop     ax
L_0790: ; original 0x236f0; IDA unnamed; logical 0x136f0; file 0x14a60
    pop ax
; Original operand: pop     es
L_0791: ; original 0x236f1; IDA unnamed; logical 0x136f1; file 0x14a61
    pop es
; Original operand: pop     ds
L_0792: ; original 0x236f2; IDA unnamed; logical 0x136f2; file 0x14a62
    pop ds
; Original operand: popf
L_0793: ; original 0x236f3; IDA unnamed; logical 0x136f3; file 0x14a63
    popf 
; Original operand: retf
L_0794: ; original 0x236f4; IDA unnamed; logical 0x136f4; file 0x14a64
    retf 
; Original operand: mov     ax, ds:33h
L_0795: ; original 0x236f5; IDA sub_236F5; logical 0x136f5; file 0x14a65
    mov ax, WORD PTR ds:[033h]
; Original operand: retn
L_0798: ; original 0x236f8; IDA unnamed; logical 0x136f8; file 0x14a68
    ret 
; Original operand: mov     bx, ax
L_0799: ; original 0x236f9; IDA sub_236F9; logical 0x136f9; file 0x14a69
    mov bx, ax
; Original operand: mov     cl, 4
L_079B: ; original 0x236fb; IDA unnamed; logical 0x136fb; file 0x14a6b
    mov cl, 04h
; Original operand: ror     ax, cl
L_079D: ; original 0x236fd; IDA unnamed; logical 0x136fd; file 0x14a6d
    ror ax, cl
; Original operand: cmp     ax, 21h ; '!'
L_079F: ; original 0x236ff; IDA unnamed; logical 0x136ff; file 0x14a6f
    cmp ax, IMM_0021
; Original operand: jb      short loc_23711
L_07A2: ; original 0x23702; IDA unnamed; logical 0x13702; file 0x14a72
    jb SHORT L_07B1
; Original operand: cmp     ax, 26h ; '&'
L_07A4: ; original 0x23704; IDA unnamed; logical 0x13704; file 0x14a74
    cmp ax, IMM_0026
; Original operand: ja      short loc_23711
L_07A7: ; original 0x23707; IDA unnamed; logical 0x13707; file 0x14a77
    ja SHORT L_07B1
; Original operand: mov     ds:30h, bx
L_07A9: ; original 0x23709; IDA unnamed; logical 0x13709; file 0x14a79
    mov WORD PTR ds:[030h], bx
; Original operand: sub     ax, ax
L_07AD: ; original 0x2370d; IDA unnamed; logical 0x1370d; file 0x14a7d
    sub ax, ax
; Original operand: jmp     short locret_23714
L_07AF: ; original 0x2370f; IDA unnamed; logical 0x1370f; file 0x14a7f
    jmp SHORT L_07B4
; Original operand: mov     ax, 0FFFFh
L_07B1: ; original 0x23711; IDA loc_23711; logical 0x13711; file 0x14a81
    mov ax, 0FFFFh
; Original operand: retn
L_07B4: ; original 0x23714; IDA locret_23714; logical 0x13714; file 0x14a84
    ret 
; Original operand: cmp     al, 7
L_07B5: ; original 0x23715; IDA sub_23715; logical 0x13715; file 0x14a85
    cmp al, 07h
; Original operand: jz      short loc_23725
L_07B7: ; original 0x23717; IDA unnamed; logical 0x13717; file 0x14a87
    jz SHORT L_07C5
; Original operand: cmp     al, 5
L_07B9: ; original 0x23719; IDA unnamed; logical 0x13719; file 0x14a89
    cmp al, 05h
; Original operand: jz      short loc_23725
L_07BB: ; original 0x2371b; IDA unnamed; logical 0x1371b; file 0x14a8b
    jz SHORT L_07C5
; Original operand: cmp     al, 3
L_07BD: ; original 0x2371d; IDA unnamed; logical 0x1371d; file 0x14a8d
    cmp al, 03h
; Original operand: jz      short loc_23725
L_07BF: ; original 0x2371f; IDA unnamed; logical 0x1371f; file 0x14a8f
    jz SHORT L_07C5
; Original operand: cmp     al, 2
L_07C1: ; original 0x23721; IDA unnamed; logical 0x13721; file 0x14a91
    cmp al, 02h
; Original operand: jnz     short loc_2372C
L_07C3: ; original 0x23723; IDA unnamed; logical 0x13723; file 0x14a93
    jnz SHORT L_07CC
; Original operand: mov     ds:32h, al
L_07C5: ; original 0x23725; IDA loc_23725; logical 0x13725; file 0x14a95
    mov BYTE PTR ds:[032h], al
; Original operand: sub     ax, ax
L_07C8: ; original 0x23728; IDA unnamed; logical 0x13728; file 0x14a98
    sub ax, ax
; Original operand: jmp     short locret_2372F
L_07CA: ; original 0x2372a; IDA unnamed; logical 0x1372a; file 0x14a9a
    jmp SHORT L_07CF
; Original operand: mov     ax, 0FFFFh
L_07CC: ; original 0x2372c; IDA loc_2372C; logical 0x1372c; file 0x14a9c
    mov ax, 0FFFFh
; Original operand: retn
L_07CF: ; original 0x2372f; IDA locret_2372F; logical 0x1372f; file 0x14a9f
    ret 
; Original operand: mov     word ptr ds:0C4h, cs
L_07D0: ; original 0x23730; IDA sub_23730; logical 0x13730; file 0x14aa0
    mov WORD PTR ds:[0C4h], cs
; Original operand: mov     word ptr ds:0C2h, 0B6h
L_07D4: ; original 0x23734; IDA unnamed; logical 0x13734; file 0x14aa4
    mov WORD PTR ds:[0C2h], 0B6h
; Original operand: mov     ax, 83Ah
L_07DA: ; original 0x2373a; IDA unnamed; logical 0x1373a; file 0x14aaa
    mov ax, 083Ah
; Original operand: cmp     ax, ds:91h
L_07DD: ; original 0x2373d; IDA unnamed; logical 0x1373d; file 0x14aad
    cmp ax, WORD PTR ds:[091h]
; Original operand: jz      short loc_23768
L_07E1: ; original 0x23741; IDA unnamed; logical 0x13741; file 0x14ab1
    jz SHORT L_0808
; Original operand: mov     dx, ds:30h
L_07E3: ; original 0x23743; IDA unnamed; logical 0x13743; file 0x14ab3
    mov dx, WORD PTR ds:[030h]
; Original operand: add     dl, 0Ch
L_07E7: ; original 0x23747; IDA unnamed; logical 0x13747; file 0x14ab7
    add dl, 0Ch
; Original operand: mov     al, 0D3h
L_07EA: ; original 0x2374a; IDA unnamed; logical 0x1374a; file 0x14aba
    mov al, 0D3h
; Original operand: call    sub_23043
L_07EC: ; original 0x2374c; IDA unnamed; logical 0x1374c; file 0x14abc
    call L_00E3
; Original operand: mov     cx, 0FFFFh
L_07EF: ; original 0x2374f; IDA unnamed; logical 0x1374f; file 0x14abf
    mov cx, 0FFFFh
; Original operand: loop    loc_23752
L_07F2: ; original 0x23752; IDA loc_23752; logical 0x13752; file 0x14ac2
    loop L_07F2
; Original operand: call    sub_23099
L_07F4: ; original 0x23754; IDA unnamed; logical 0x13754; file 0x14ac4
    call L_0139
; Original operand: jnz     short locret_2376F
L_07F7: ; original 0x23757; IDA unnamed; logical 0x13757; file 0x14ac7
    jnz SHORT L_080F
; Original operand: call    sub_230C2
L_07F9: ; original 0x23759; IDA unnamed; logical 0x13759; file 0x14ac9
    call L_0162
; Original operand: jnz     short locret_2376F
L_07FC: ; original 0x2375c; IDA unnamed; logical 0x1375c; file 0x14acc
    jnz SHORT L_080F
; Original operand: call    sub_23190
L_07FE: ; original 0x2375e; IDA unnamed; logical 0x1375e; file 0x14ace
    call L_0230
; Original operand: jnz     short locret_2376F
L_0801: ; original 0x23761; IDA unnamed; logical 0x13761; file 0x14ad1
    jnz SHORT L_080F
; Original operand: call    sub_230EA
L_0803: ; original 0x23763; IDA unnamed; logical 0x13763; file 0x14ad3
    call L_018A
; Original operand: jnz     short locret_2376F
L_0806: ; original 0x23766; IDA unnamed; logical 0x13766; file 0x14ad6
    jnz SHORT L_080F
; Original operand: mov     al, 1
L_0808: ; original 0x23768; IDA loc_23768; logical 0x13768; file 0x14ad8
    mov al, 01h
; Original operand: call    sub_23770
L_080A: ; original 0x2376a; IDA unnamed; logical 0x1376a; file 0x14ada
    call L_0810
; Original operand: sub     ax, ax
L_080D: ; original 0x2376d; IDA unnamed; logical 0x1376d; file 0x14add
    sub ax, ax
; Original operand: retn
L_080F: ; original 0x2376f; IDA locret_2376F; logical 0x1376f; file 0x14adf
    ret 
; Original operand: mov     dx, ds:30h
L_0810: ; original 0x23770; IDA sub_23770; logical 0x13770; file 0x14ae0
    mov dx, WORD PTR ds:[030h]
; Original operand: add     dx, 0Ch
L_0814: ; original 0x23774; IDA unnamed; logical 0x13774; file 0x14ae4
    add dx, 0Ch
; Original operand: mov     ah, 0D1h
L_0817: ; original 0x23777; IDA unnamed; logical 0x13777; file 0x14ae7
    mov ah, 0D1h
; Original operand: or      al, al
L_0819: ; original 0x23779; IDA unnamed; logical 0x13779; file 0x14ae9
    or al, al
; Original operand: jnz     short loc_2377F
L_081B: ; original 0x2377b; IDA unnamed; logical 0x1377b; file 0x14aeb
    jnz SHORT L_081F
; Original operand: mov     ah, 0D3h
L_081D: ; original 0x2377d; IDA unnamed; logical 0x1377d; file 0x14aed
    mov ah, 0D3h
; Original operand: mov     al, ah
L_081F: ; original 0x2377f; IDA loc_2377F; logical 0x1377f; file 0x14aef
    mov al, ah
; Original operand: call    sub_23077
L_0821: ; original 0x23781; IDA unnamed; logical 0x13781; file 0x14af1
    call L_0117
; Original operand: sub     ax, ax
L_0824: ; original 0x23784; IDA unnamed; logical 0x13784; file 0x14af4
    sub ax, ax
; Original operand: retn
L_0826: ; original 0x23786; IDA unnamed; logical 0x13786; file 0x14af6
    ret 
; Original operand: sub     ax, ax
L_0827: ; original 0x23787; IDA sub_23787; logical 0x13787; file 0x14af7
    sub ax, ax
; Original operand: push    es
L_0829: ; original 0x23789; IDA unnamed; logical 0x13789; file 0x14af9
    push es
; Original operand: mov     es, word ptr [bp+0Eh]
L_082A: ; original 0x2378a; IDA unnamed; logical 0x1378a; file 0x14afa
    mov es, WORD PTR [bp+0Eh]
; Original operand: mov     es:[di], ax
L_082D: ; original 0x2378d; IDA unnamed; logical 0x1378d; file 0x14afd
    mov WORD PTR es:[di], ax
; Original operand: mov     word ptr ds:0C4h, es
L_0830: ; original 0x23790; IDA unnamed; logical 0x13790; file 0x14b00
    mov WORD PTR ds:[0C4h], es
; Original operand: mov     ds:0C2h, di
L_0834: ; original 0x23794; IDA unnamed; logical 0x13794; file 0x14b04
    mov WORD PTR ds:[0C2h], di
; Original operand: pop     es
L_0838: ; original 0x23798; IDA unnamed; logical 0x13798; file 0x14b08
    pop es
; Original operand: retn
L_0839: ; original 0x23799; IDA unnamed; logical 0x13799; file 0x14b09
    ret 
; Original operand: cmp     byte ptr ds:0B2h, 0
L_083A: ; original 0x2379a; IDA unnamed; logical 0x1379a; file 0x14b0a
    cmp BYTE PTR ds:[0B2h], 00h
; Original operand: jz      short loc_237A6
L_083F: ; original 0x2379f; IDA unnamed; logical 0x1379f; file 0x14b0f
    jz SHORT L_0846
; Original operand: mov     ax, 1
L_0841: ; original 0x237a1; IDA unnamed; logical 0x137a1; file 0x14b11
    mov ax, 01h
; Original operand: jmp     short locret_237EB
L_0844: ; original 0x237a4; IDA unnamed; logical 0x137a4; file 0x14b14
    jmp SHORT L_088B
; Original operand: mov     byte ptr ds:0B2h, 1
L_0846: ; original 0x237a6; IDA loc_237A6; logical 0x137a6; file 0x14b16
    mov BYTE PTR ds:[0B2h], 01h
; Original operand: mov     dx, [bp+0Eh]
L_084B: ; original 0x237ab; IDA unnamed; logical 0x137ab; file 0x14b1b
    mov dx, WORD PTR [bp+0Eh]
; Original operand: mov     ax, di
L_084E: ; original 0x237ae; IDA unnamed; logical 0x137ae; file 0x14b1e
    mov ax, di
; Original operand: call    sub_2358C
L_0850: ; original 0x237b0; IDA unnamed; logical 0x137b0; file 0x14b20
    call L_062C
; Original operand: mov     ds:0CCh, dx
L_0853: ; original 0x237b3; IDA unnamed; logical 0x137b3; file 0x14b23
    mov WORD PTR ds:[0CCh], dx
; Original operand: mov     ds:0CAh, ax
L_0857: ; original 0x237b7; IDA unnamed; logical 0x137b7; file 0x14b27
    mov WORD PTR ds:[0CAh], ax
; Original operand: sub     ax, ax
L_085A: ; original 0x237ba; IDA unnamed; logical 0x137ba; file 0x14b2a
    sub ax, ax
; Original operand: mov     ds:0E2h, al
L_085C: ; original 0x237bc; IDA unnamed; logical 0x137bc; file 0x14b2c
    mov BYTE PTR ds:[0E2h], al
; Original operand: mov     ds:0D2h, ax
L_085F: ; original 0x237bf; IDA unnamed; logical 0x137bf; file 0x14b2f
    mov WORD PTR ds:[0D2h], ax
; Original operand: mov     ds:0D4h, ax
L_0862: ; original 0x237c2; IDA unnamed; logical 0x137c2; file 0x14b32
    mov WORD PTR ds:[0D4h], ax
; Original operand: mov     ds:0DEh, ax
L_0865: ; original 0x237c5; IDA unnamed; logical 0x137c5; file 0x14b35
    mov WORD PTR ds:[0DEh], ax
; Original operand: mov     ds:0DCh, ax
L_0868: ; original 0x237c8; IDA unnamed; logical 0x137c8; file 0x14b38
    mov WORD PTR ds:[0DCh], ax
; Original operand: call    sub_232CA
L_086B: ; original 0x237cb; IDA unnamed; logical 0x137cb; file 0x14b3b
    call L_036A
; Original operand: mov     ax, 3E1h
L_086E: ; original 0x237ce; IDA unnamed; logical 0x137ce; file 0x14b3e
    mov ax, 03E1h
; Original operand: call    sub_2323B
L_0871: ; original 0x237d1; IDA unnamed; logical 0x137d1; file 0x14b41
    call L_02DB
; Original operand: mov     ax, 0FFFFh
L_0874: ; original 0x237d4; IDA unnamed; logical 0x137d4; file 0x14b44
    mov ax, 0FFFFh
; Original operand: call    sub_232CA
L_0877: ; original 0x237d7; IDA unnamed; logical 0x137d7; file 0x14b47
    call L_036A
; Original operand: call    sub_23511
L_087A: ; original 0x237da; IDA unnamed; logical 0x137da; file 0x14b4a
    call L_05B1
; Original operand: cmp     byte ptr ds:0E2h, 1
L_087D: ; original 0x237dd; IDA unnamed; logical 0x137dd; file 0x14b4d
    cmp BYTE PTR ds:[0E2h], 01h
; Original operand: jnz     short loc_237E9
L_0882: ; original 0x237e2; IDA unnamed; logical 0x137e2; file 0x14b52
    jnz SHORT L_0889
; Original operand: sub     ax, ax
L_0884: ; original 0x237e4; IDA unnamed; logical 0x137e4; file 0x14b54
    sub ax, ax
; Original operand: call    sub_232CA
L_0886: ; original 0x237e6; IDA unnamed; logical 0x137e6; file 0x14b56
    call L_036A
; Original operand: sub     ax, ax
L_0889: ; original 0x237e9; IDA loc_237E9; logical 0x137e9; file 0x14b59
    sub ax, ax
; Original operand: retn
L_088B: ; original 0x237eb; IDA locret_237EB; logical 0x137eb; file 0x14b5b
    ret 
; Original operand: mov     ax, 1
L_088C: ; original 0x237ec; IDA sub_237EC; logical 0x137ec; file 0x14b5c
    mov ax, 01h
; Original operand: cmp     byte ptr ds:0B2h, 0
L_088F: ; original 0x237ef; IDA unnamed; logical 0x137ef; file 0x14b5f
    cmp BYTE PTR ds:[0B2h], 00h
; Original operand: jz      short locret_237FE
L_0894: ; original 0x237f4; IDA unnamed; logical 0x137f4; file 0x14b64
    jz SHORT L_089E
; Original operand: call    sub_231AD
L_0896: ; original 0x237f6; IDA unnamed; logical 0x137f6; file 0x14b66
    call L_024D
; Original operand: call    sub_233EC
L_0899: ; original 0x237f9; IDA unnamed; logical 0x137f9; file 0x14b69
    call L_048C
; Original operand: sub     ax, ax
L_089C: ; original 0x237fc; IDA unnamed; logical 0x137fc; file 0x14b6c
    sub ax, ax
; Original operand: retn
L_089E: ; original 0x237fe; IDA locret_237FE; logical 0x137fe; file 0x14b6e
    ret 
; Original operand: cmp     byte ptr ds:0B2h, 0
L_089F: ; original 0x237ff; IDA unnamed; logical 0x137ff; file 0x14b6f
    cmp BYTE PTR ds:[0B2h], 00h
; Original operand: jz      short loc_2380C
L_08A4: ; original 0x23804; IDA unnamed; logical 0x13804; file 0x14b74
    jz SHORT L_08AC
; Original operand: mov     ax, 1
L_08A6: ; original 0x23806; IDA unnamed; logical 0x13806; file 0x14b76
    mov ax, 01h
; Original operand: jmp     locret_2389E
L_08A9: ; original 0x23809; IDA unnamed; logical 0x13809; file 0x14b79
    jmp NEAR PTR L_093E
; Original operand: or      dx, dx
L_08AC: ; original 0x2380c; IDA loc_2380C; logical 0x1380c; file 0x14b7c
    or dx, dx
; Original operand: jnz     short loc_2381B
L_08AE: ; original 0x2380e; IDA unnamed; logical 0x1380e; file 0x14b7e
    jnz SHORT L_08BB
; Original operand: cmp     cx, 7
L_08B0: ; original 0x23810; IDA unnamed; logical 0x13810; file 0x14b80
    cmp cx, 07h
; Original operand: ja      short loc_2381B
L_08B3: ; original 0x23813; IDA unnamed; logical 0x13813; file 0x14b83
    ja SHORT L_08BB
; Original operand: mov     ax, 2
L_08B5: ; original 0x23815; IDA unnamed; logical 0x13815; file 0x14b85
    mov ax, 02h
; Original operand: jmp     locret_2389E
L_08B8: ; original 0x23818; IDA unnamed; logical 0x13818; file 0x14b88
    jmp NEAR PTR L_093E
; Original operand: mov     byte ptr ds:0B2h, 2
L_08BB: ; original 0x2381b; IDA loc_2381B; logical 0x1381b; file 0x14b8b
    mov BYTE PTR ds:[0B2h], 02h
; Original operand: sub     cx, 7
L_08C0: ; original 0x23820; IDA unnamed; logical 0x13820; file 0x14b90
    sub cx, 07h
; Original operand: sbb     dx, 0
L_08C3: ; original 0x23823; IDA unnamed; logical 0x13823; file 0x14b93
    sbb dx, 00h
; Original operand: mov     ds:0DCh, cx
L_08C6: ; original 0x23826; IDA unnamed; logical 0x13826; file 0x14b96
    mov WORD PTR ds:[0DCh], cx
; Original operand: mov     ds:0DEh, dx
L_08CA: ; original 0x2382a; IDA unnamed; logical 0x1382a; file 0x14b9a
    mov WORD PTR ds:[0DEh], dx
; Original operand: mov     cx, ax
L_08CE: ; original 0x2382e; IDA unnamed; logical 0x1382e; file 0x14b9e
    mov cx, ax
; Original operand: mov     dx, 0Fh
L_08D0: ; original 0x23830; IDA unnamed; logical 0x13830; file 0x14ba0
    mov dx, 0Fh
; Original operand: mov     ax, 4240h
L_08D3: ; original 0x23833; IDA unnamed; logical 0x13833; file 0x14ba3
    mov ax, 04240h
; Original operand: div     cx
L_08D6: ; original 0x23836; IDA unnamed; logical 0x13836; file 0x14ba6
    div cx
; Original operand: neg     al
L_08D8: ; original 0x23838; IDA unnamed; logical 0x13838; file 0x14ba8
    neg al
; Original operand: mov     ds:0B5h, al
L_08DA: ; original 0x2383a; IDA unnamed; logical 0x1383a; file 0x14baa
    mov BYTE PTR ds:[0B5h], al
; Original operand: mov     dx, ds:30h
L_08DD: ; original 0x2383d; IDA unnamed; logical 0x1383d; file 0x14bad
    mov dx, WORD PTR ds:[030h]
; Original operand: add     dl, 0Ch
L_08E1: ; original 0x23841; IDA unnamed; logical 0x13841; file 0x14bb1
    add dl, 0Ch
; Original operand: mov     al, 40h ; '@'
L_08E4: ; original 0x23844; IDA unnamed; logical 0x13844; file 0x14bb4
    mov al, 040h
; Original operand: call    sub_23077
L_08E6: ; original 0x23846; IDA unnamed; logical 0x13846; file 0x14bb6
    call L_0117
; Original operand: mov     al, ds:0B5h
L_08E9: ; original 0x23849; IDA unnamed; logical 0x13849; file 0x14bb9
    mov al, BYTE PTR ds:[0B5h]
; Original operand: call    sub_23077
L_08EC: ; original 0x2384c; IDA unnamed; logical 0x1384c; file 0x14bbc
    call L_0117
; Original operand: mov     dx, [bp+0Eh]
L_08EF: ; original 0x2384f; IDA unnamed; logical 0x1384f; file 0x14bbf
    mov dx, WORD PTR [bp+0Eh]
; Original operand: mov     ax, di
L_08F2: ; original 0x23852; IDA unnamed; logical 0x13852; file 0x14bc2
    mov ax, di
; Original operand: call    sub_2358C
L_08F4: ; original 0x23854; IDA unnamed; logical 0x13854; file 0x14bc4
    call L_062C
; Original operand: mov     ds:0C6h, ax
L_08F7: ; original 0x23857; IDA unnamed; logical 0x13857; file 0x14bc7
    mov WORD PTR ds:[0C6h], ax
; Original operand: mov     ds:0C8h, dx
L_08FA: ; original 0x2385a; IDA unnamed; logical 0x1385a; file 0x14bca
    mov WORD PTR ds:[0C8h], dx
; Original operand: add     ax, 6
L_08FE: ; original 0x2385e; IDA unnamed; logical 0x1385e; file 0x14bce
    add ax, IMM_0006
; Original operand: mov     ds:0CAh, ax
L_0901: ; original 0x23861; IDA unnamed; logical 0x13861; file 0x14bd1
    mov WORD PTR ds:[0CAh], ax
; Original operand: mov     ds:0CCh, dx
L_0904: ; original 0x23864; IDA unnamed; logical 0x13864; file 0x14bd4
    mov WORD PTR ds:[0CCh], dx
; Original operand: call    sub_2331D
L_0908: ; original 0x23868; IDA unnamed; logical 0x13868; file 0x14bd8
    call L_03BD
; Original operand: push    ds
L_090B: ; original 0x2386b; IDA unnamed; logical 0x1386b; file 0x14bdb
    push ds
; Original operand: push    bx
L_090C: ; original 0x2386c; IDA unnamed; logical 0x1386c; file 0x14bdc
    push bx
; Original operand: lds     bx, ds:0C6h
L_090D: ; original 0x2386d; IDA unnamed; logical 0x1386d; file 0x14bdd
    lds bx, DWORD PTR ds:[0C6h]
; Original operand: mov     byte ptr [bx], 1
L_0911: ; original 0x23871; IDA unnamed; logical 0x13871; file 0x14be1
    mov BYTE PTR [bx], 01h
; Original operand: sub     ax, ax
L_0914: ; original 0x23874; IDA unnamed; logical 0x13874; file 0x14be4
    sub ax, ax
; Original operand: mov     [bx+1], ax
L_0916: ; original 0x23876; IDA unnamed; logical 0x13876; file 0x14be6
    mov WORD PTR [bx+01h], ax
; Original operand: mov     [bx+3], al
L_0919: ; original 0x23879; IDA unnamed; logical 0x13879; file 0x14be9
    mov BYTE PTR [bx+03h], al
; Original operand: mov     [bx+5], al
L_091C: ; original 0x2387c; IDA unnamed; logical 0x1387c; file 0x14bec
    mov BYTE PTR [bx+05h], al
; Original operand: mov     al, cs:byte_23012+3
L_091F: ; original 0x2387f; IDA unnamed; logical 0x1387f; file 0x14bef
    mov al, BYTE PTR cs:[0B5h]
; Original operand: mov     [bx+4], al
L_0923: ; original 0x23883; IDA unnamed; logical 0x13883; file 0x14bf3
    mov BYTE PTR [bx+04h], al
; Original operand: pop     bx
L_0926: ; original 0x23886; IDA unnamed; logical 0x13886; file 0x14bf6
    pop bx
; Original operand: pop     ds
L_0927: ; original 0x23887; IDA unnamed; logical 0x13887; file 0x14bf7
    pop ds
; Original operand: sub     ax, ax
L_0928: ; original 0x23888; IDA unnamed; logical 0x13888; file 0x14bf8
    sub ax, ax
; Original operand: call    sub_232CA
L_092A: ; original 0x2388a; IDA unnamed; logical 0x1388a; file 0x14bfa
    call L_036A
; Original operand: mov     ax, 4F7h
L_092D: ; original 0x2388d; IDA unnamed; logical 0x1388d; file 0x14bfd
    mov ax, 04F7h
; Original operand: call    sub_2323B
L_0930: ; original 0x23890; IDA unnamed; logical 0x13890; file 0x14c00
    call L_02DB
; Original operand: mov     ax, 0FFFFh
L_0933: ; original 0x23893; IDA unnamed; logical 0x13893; file 0x14c03
    mov ax, 0FFFFh
; Original operand: call    sub_232CA
L_0936: ; original 0x23896; IDA unnamed; logical 0x13896; file 0x14c06
    call L_036A
; Original operand: call    sub_2348F
L_0939: ; original 0x23899; IDA unnamed; logical 0x13899; file 0x14c09
    call L_052F
; Original operand: sub     ax, ax
L_093C: ; original 0x2389c; IDA unnamed; logical 0x1389c; file 0x14c0c
    sub ax, ax
; Original operand: retn
L_093E: ; original 0x2389e; IDA locret_2389E; logical 0x1389e; file 0x14c0e
    ret 
; Original operand: mov     ax, 1
L_093F: ; original 0x2389f; IDA unnamed; logical 0x1389f; file 0x14c0f
    mov ax, 01h
; Original operand: cmp     byte ptr ds:0B2h, 1
L_0942: ; original 0x238a2; IDA unnamed; logical 0x138a2; file 0x14c12
    cmp BYTE PTR ds:[0B2h], 01h
; Original operand: jnz     short locret_238AE
L_0947: ; original 0x238a7; IDA unnamed; logical 0x138a7; file 0x14c17
    jnz SHORT L_094E
; Original operand: call    sub_231AD
L_0949: ; original 0x238a9; IDA unnamed; logical 0x138a9; file 0x14c19
    call L_024D
; Original operand: sub     ax, ax
L_094C: ; original 0x238ac; IDA unnamed; logical 0x138ac; file 0x14c1c
    sub ax, ax
; Original operand: retn
L_094E: ; original 0x238ae; IDA locret_238AE; logical 0x138ae; file 0x14c1e
    ret 
; Original operand: mov     ax, 1
L_094F: ; original 0x238af; IDA unnamed; logical 0x138af; file 0x14c1f
    mov ax, 01h
; Original operand: cmp     byte ptr ds:0B2h, 1
L_0952: ; original 0x238b2; IDA unnamed; logical 0x138b2; file 0x14c22
    cmp BYTE PTR ds:[0B2h], 01h
; Original operand: jnz     short locret_238C7
L_0957: ; original 0x238b7; IDA unnamed; logical 0x138b7; file 0x14c27
    jnz SHORT L_0967
; Original operand: mov     dx, ds:30h
L_0959: ; original 0x238b9; IDA unnamed; logical 0x138b9; file 0x14c29
    mov dx, WORD PTR ds:[030h]
; Original operand: add     dl, 0Ch
L_095D: ; original 0x238bd; IDA unnamed; logical 0x138bd; file 0x14c2d
    add dl, 0Ch
; Original operand: mov     al, 0D4h
L_0960: ; original 0x238c0; IDA unnamed; logical 0x138c0; file 0x14c30
    mov al, 0D4h
; Original operand: call    sub_23077
L_0962: ; original 0x238c2; IDA unnamed; logical 0x138c2; file 0x14c32
    call L_0117
; Original operand: sub     ax, ax
L_0965: ; original 0x238c5; IDA unnamed; logical 0x138c5; file 0x14c35
    sub ax, ax
; Original operand: retn
L_0967: ; original 0x238c7; IDA locret_238C7; logical 0x138c7; file 0x14c37
    ret 
; Original operand: call    sub_237EC
L_0968: ; original 0x238c8; IDA unnamed; logical 0x138c8; file 0x14c38
    call L_088C
; Original operand: sub     al, al
L_096B: ; original 0x238cb; IDA unnamed; logical 0x138cb; file 0x14c3b
    sub al, al
; Original operand: call    sub_23770
L_096D: ; original 0x238cd; IDA unnamed; logical 0x138cd; file 0x14c3d
    call L_0810
; Original operand: sub     ax, ax
L_0970: ; original 0x238d0; IDA unnamed; logical 0x138d0; file 0x14c40
    sub ax, ax
; Original operand: retn
L_0972: ; original 0x238d2; IDA unnamed; logical 0x138d2; file 0x14c42
    ret 
; Original operand: mov     bx, ax
L_0973: ; original 0x238d3; IDA unnamed; logical 0x138d3; file 0x14c43
    mov bx, ax
; Original operand: mov     ax, 1
L_0975: ; original 0x238d5; IDA unnamed; logical 0x138d5; file 0x14c45
    mov ax, 01h
; Original operand: pushf
L_0978: ; original 0x238d8; IDA unnamed; logical 0x138d8; file 0x14c48
    pushf 
; Original operand: cli
L_0979: ; original 0x238d9; IDA unnamed; logical 0x138d9; file 0x14c49
    cli 
; Original operand: cmp     word ptr ds:0D4h, 0
L_097A: ; original 0x238da; IDA unnamed; logical 0x138da; file 0x14c4a
    cmp WORD PTR ds:[0D4h], 00h
; Original operand: jz      short loc_23910
L_097F: ; original 0x238df; IDA unnamed; logical 0x138df; file 0x14c4f
    jz SHORT L_09B0
; Original operand: mov     word ptr ds:0D2h, 0
L_0981: ; original 0x238e1; IDA unnamed; logical 0x138e1; file 0x14c51
    mov WORD PTR ds:[0D2h], 00h
; Original operand: mov     word ptr ds:0D4h, 0
L_0987: ; original 0x238e7; IDA unnamed; logical 0x138e7; file 0x14c57
    mov WORD PTR ds:[0D4h], 00h
; Original operand: or      bx, bx
L_098D: ; original 0x238ed; IDA unnamed; logical 0x138ed; file 0x14c5d
    or bx, bx
; Original operand: jz      short loc_2390E
L_098F: ; original 0x238ef; IDA unnamed; logical 0x138ef; file 0x14c5f
    jz SHORT L_09AE
; Original operand: call    sub_231AD
L_0991: ; original 0x238f1; IDA unnamed; logical 0x138f1; file 0x14c61
    call L_024D
; Original operand: call    sub_23546
L_0994: ; original 0x238f4; IDA loc_238F4; logical 0x138f4; file 0x14c64
    call L_05E6
; Original operand: call    sub_232D5
L_0997: ; original 0x238f7; IDA unnamed; logical 0x138f7; file 0x14c67
    call L_0375
; Original operand: cmp     al, 7
L_099A: ; original 0x238fa; IDA unnamed; logical 0x138fa; file 0x14c6a
    cmp al, 07h
; Original operand: jnz     short loc_238F4
L_099C: ; original 0x238fc; IDA unnamed; logical 0x138fc; file 0x14c6c
    jnz SHORT L_0994
; Original operand: call    sub_23546
L_099E: ; original 0x238fe; IDA unnamed; logical 0x138fe; file 0x14c6e
    call L_05E6
; Original operand: call    sub_23511
L_09A1: ; original 0x23901; IDA unnamed; logical 0x13901; file 0x14c71
    call L_05B1
; Original operand: cmp     byte ptr ds:0E2h, 0
L_09A4: ; original 0x23904; IDA unnamed; logical 0x13904; file 0x14c74
    cmp BYTE PTR ds:[0E2h], 00h
; Original operand: jz      short loc_2390E
L_09A9: ; original 0x23909; IDA unnamed; logical 0x13909; file 0x14c79
    jz SHORT L_09AE
; Original operand: call    sub_233EC
L_09AB: ; original 0x2390b; IDA unnamed; logical 0x1390b; file 0x14c7b
    call L_048C
; Original operand: sub     ax, ax
L_09AE: ; original 0x2390e; IDA loc_2390E; logical 0x1390e; file 0x14c7e
    sub ax, ax
; Original operand: popf
L_09B0: ; original 0x23910; IDA loc_23910; logical 0x13910; file 0x14c80
    popf 
; Original operand: retn
L_09B1: ; original 0x23911; IDA unnamed; logical 0x13911; file 0x14c81
    ret 
; Original operand: pushf
L_09B2: ; original 0x23912; IDA unnamed; logical 0x13912; file 0x14c82
    pushf 
; Original operand: cli
L_09B3: ; original 0x23913; IDA unnamed; logical 0x13913; file 0x14c83
    cli 
; Original operand: mov     ds:0BCh, dx
L_09B4: ; original 0x23914; IDA unnamed; logical 0x13914; file 0x14c84
    mov WORD PTR ds:[0BCh], dx
; Original operand: mov     ds:0BAh, ax
L_09B8: ; original 0x23918; IDA unnamed; logical 0x13918; file 0x14c88
    mov WORD PTR ds:[0BAh], ax
; Original operand: popf
L_09BB: ; original 0x2391b; IDA unnamed; logical 0x1391b; file 0x14c8b
    popf 
; Original operand: retn
L_09BC: ; original 0x2391c; IDA unnamed; logical 0x1391c; file 0x14c8c
    ret 
_TEXT ENDS
END module_entry
