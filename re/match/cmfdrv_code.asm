; Verified instruction source only. Original SDK data is not reconstructed.
; Input DQ3.EXE 115282 bytes SHA256 5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c
; CMFDRV.ASM original IDA linear 0x23920..0x24dd0
; BYTE/WORD qualifiers and absolute EQU symbols preserve assembler encoding.
.8086
EXTRN IMM_0064:ABS
EXTRN IMM_01A0:ABS
EXTRN IMM_0300:ABS
EXTRN DISP_000E:ABS
EXTRN DISP_0010:ABS
EXTRN IMM_FFFF:ABS
_TEXT SEGMENT PARA PUBLIC USE16 'CODE'
ASSUME CS:_TEXT, DS:_TEXT
PUBLIC module_entry
module_entry LABEL NEAR
; Original operand: jmp     near ptr sub_24A2F
L_0000: ; original 0x23920; IDA sub_23920; logical 0x13920; file 0x14c90
    jmp NEAR PTR L_110F
; Unknown data skipped: 0x23923..0x24225
ORG 0905h
; Original operand: pushf
L_0905: ; original 0x24225; IDA sub_24225; logical 0x14225; file 0x15595
    pushf 
; Original operand: push    ds
L_0906: ; original 0x24226; IDA unnamed; logical 0x14226; file 0x15596
    push ds
; Original operand: push    ax
L_0907: ; original 0x24227; IDA unnamed; logical 0x14227; file 0x15597
    push ax
; Original operand: sub     ax, ax
L_0908: ; original 0x24228; IDA unnamed; logical 0x14228; file 0x15598
    sub ax, ax
; Original operand: mov     ds, ax
L_090A: ; original 0x2422a; IDA unnamed; logical 0x1422a; file 0x1559a
    mov ds, ax
; Original operand: pop     ax
L_090C: ; original 0x2422c; IDA unnamed; logical 0x1422c; file 0x1559c
    pop ax
; Original operand: shl     bx, 1
L_090D: ; original 0x2422d; IDA unnamed; logical 0x1422d; file 0x1559d
    shl bx, 01h
; Original operand: shl     bx, 1
L_090F: ; original 0x2422f; IDA unnamed; logical 0x1422f; file 0x1559f
    shl bx, 01h
; Original operand: cli
L_0911: ; original 0x24231; IDA unnamed; logical 0x14231; file 0x155a1
    cli 
; Original operand: mov     [bx], ax
L_0912: ; original 0x24232; IDA unnamed; logical 0x14232; file 0x155a2
    mov WORD PTR [bx], ax
; Original operand: mov     [bx+2], dx
L_0914: ; original 0x24234; IDA unnamed; logical 0x14234; file 0x155a4
    mov WORD PTR [bx+02h], dx
; Original operand: pop     ds
L_0917: ; original 0x24237; IDA unnamed; logical 0x14237; file 0x155a7
    pop ds
; Original operand: popf
L_0918: ; original 0x24238; IDA unnamed; logical 0x14238; file 0x155a8
    popf 
; Original operand: retn
L_0919: ; original 0x24239; IDA unnamed; logical 0x14239; file 0x155a9
    ret 
; Original operand: pushf
L_091A: ; original 0x2423a; IDA sub_2423A; logical 0x1423a; file 0x155aa
    pushf 
; Original operand: push    ds
L_091B: ; original 0x2423b; IDA unnamed; logical 0x1423b; file 0x155ab
    push ds
; Original operand: sub     ax, ax
L_091C: ; original 0x2423c; IDA unnamed; logical 0x1423c; file 0x155ac
    sub ax, ax
; Original operand: mov     ds, ax
L_091E: ; original 0x2423e; IDA unnamed; logical 0x1423e; file 0x155ae
    mov ds, ax
; Original operand: shl     bx, 1
L_0920: ; original 0x24240; IDA unnamed; logical 0x14240; file 0x155b0
    shl bx, 01h
; Original operand: shl     bx, 1
L_0922: ; original 0x24242; IDA unnamed; logical 0x14242; file 0x155b2
    shl bx, 01h
; Original operand: cli
L_0924: ; original 0x24244; IDA unnamed; logical 0x14244; file 0x155b4
    cli 
; Original operand: mov     ax, [bx]
L_0925: ; original 0x24245; IDA unnamed; logical 0x14245; file 0x155b5
    mov ax, WORD PTR [bx]
; Original operand: mov     dx, [bx+2]
L_0927: ; original 0x24247; IDA unnamed; logical 0x14247; file 0x155b7
    mov dx, WORD PTR [bx+02h]
; Original operand: pop     ds
L_092A: ; original 0x2424a; IDA unnamed; logical 0x1424a; file 0x155ba
    pop ds
; Original operand: popf
L_092B: ; original 0x2424b; IDA unnamed; logical 0x1424b; file 0x155bb
    popf 
; Original operand: retn
L_092C: ; original 0x2424c; IDA unnamed; logical 0x1424c; file 0x155bc
    ret 
; Original operand: mov     ds:36h, ax
L_092D: ; original 0x2424d; IDA sub_2424D; logical 0x1424d; file 0x155bd
    mov WORD PTR ds:[036h], ax
; Original operand: mov     al, 36h ; '6'
L_0930: ; original 0x24250; IDA unnamed; logical 0x14250; file 0x155c0
    mov al, 036h
; Original operand: out     43h, al; Timer 8253-5 (AT: 8254.2).
L_0932: ; original 0x24252; IDA unnamed; logical 0x14252; file 0x155c2
    out 043h, al
; Original operand: mov     al, ds:36h
L_0934: ; original 0x24254; IDA unnamed; logical 0x14254; file 0x155c4
    mov al, BYTE PTR ds:[036h]
; Original operand: out     40h, al; Timer 8253-5 (AT: 8254.2).
L_0937: ; original 0x24257; IDA unnamed; logical 0x14257; file 0x155c7
    out 040h, al
; Original operand: mov     al, ah
L_0939: ; original 0x24259; IDA unnamed; logical 0x14259; file 0x155c9
    mov al, ah
; Original operand: out     40h, al; Timer 8253-5 (AT: 8254.2).
L_093B: ; original 0x2425b; IDA unnamed; logical 0x1425b; file 0x155cb
    out 040h, al
; Original operand: retn
L_093D: ; original 0x2425d; IDA unnamed; logical 0x1425d; file 0x155cd
    ret 
; Original operand: push    ax
L_093E: ; original 0x2425e; IDA sub_2425E; logical 0x1425e; file 0x155ce
    push ax
; Original operand: push    cx
L_093F: ; original 0x2425f; IDA unnamed; logical 0x1425f; file 0x155cf
    push cx
; Original operand: push    dx
L_0940: ; original 0x24260; IDA unnamed; logical 0x14260; file 0x155d0
    push dx
; Original operand: mov     dx, ds:9
L_0941: ; original 0x24261; IDA unnamed; logical 0x14261; file 0x155d1
    mov dx, WORD PTR ds:[09h]
; Original operand: xchg    ah, al
L_0945: ; original 0x24265; IDA unnamed; logical 0x14265; file 0x155d5
    xchg al, ah
; Original operand: out     dx, al
L_0947: ; original 0x24267; IDA unnamed; logical 0x14267; file 0x155d7
    out dx, al
; Original operand: mov     cx, ds:2Eh
L_0948: ; original 0x24268; IDA unnamed; logical 0x14268; file 0x155d8
    mov cx, WORD PTR ds:[02Eh]
; Original operand: nop
L_094C: ; original 0x2426c; IDA loc_2426C; logical 0x1426c; file 0x155dc
    nop 
; Original operand: dec     cx
L_094D: ; original 0x2426d; IDA unnamed; logical 0x1426d; file 0x155dd
    dec cx
; Original operand: or      cx, cx
L_094E: ; original 0x2426e; IDA unnamed; logical 0x1426e; file 0x155de
    or cx, cx
; Original operand: jnz     short loc_2426C
L_0950: ; original 0x24270; IDA unnamed; logical 0x14270; file 0x155e0
    jnz SHORT L_094C
; Original operand: inc     dx
L_0952: ; original 0x24272; IDA unnamed; logical 0x14272; file 0x155e2
    inc dx
; Original operand: mov     al, ah
L_0953: ; original 0x24273; IDA unnamed; logical 0x14273; file 0x155e3
    mov al, ah
; Original operand: out     dx, al
L_0955: ; original 0x24275; IDA unnamed; logical 0x14275; file 0x155e5
    out dx, al
; Original operand: mov     cx, ds:30h
L_0956: ; original 0x24276; IDA unnamed; logical 0x14276; file 0x155e6
    mov cx, WORD PTR ds:[030h]
; Original operand: nop
L_095A: ; original 0x2427a; IDA loc_2427A; logical 0x1427a; file 0x155ea
    nop 
; Original operand: dec     cx
L_095B: ; original 0x2427b; IDA unnamed; logical 0x1427b; file 0x155eb
    dec cx
; Original operand: or      cx, cx
L_095C: ; original 0x2427c; IDA unnamed; logical 0x1427c; file 0x155ec
    or cx, cx
; Original operand: jnz     short loc_2427A
L_095E: ; original 0x2427e; IDA unnamed; logical 0x1427e; file 0x155ee
    jnz SHORT L_095A
; Original operand: pop     dx
L_0960: ; original 0x24280; IDA unnamed; logical 0x14280; file 0x155f0
    pop dx
; Original operand: pop     cx
L_0961: ; original 0x24281; IDA unnamed; logical 0x14281; file 0x155f1
    pop cx
; Original operand: pop     ax
L_0962: ; original 0x24282; IDA unnamed; logical 0x14282; file 0x155f2
    pop ax
; Original operand: retn
L_0963: ; original 0x24283; IDA unnamed; logical 0x14283; file 0x155f3
    ret 
; Original operand: pushf
L_0964: ; original 0x24284; IDA sub_24284; logical 0x14284; file 0x155f4
    pushf 
; Original operand: cli
L_0965: ; original 0x24285; IDA unnamed; logical 0x14285; file 0x155f5
    cli 
; Original operand: mov     bx, 8
L_0966: ; original 0x24286; IDA unnamed; logical 0x14286; file 0x155f6
    mov bx, 08h
; Original operand: mov     ah, 83h
L_0969: ; original 0x24289; IDA loc_24289; logical 0x14289; file 0x155f9
    mov ah, 083h
; Original operand: add     al, bl
L_096B: ; original 0x2428b; IDA unnamed; logical 0x1428b; file 0x155fb
    add al, bl
; Original operand: mov     al, 13h
L_096D: ; original 0x2428d; IDA unnamed; logical 0x1428d; file 0x155fd
    mov al, 013h
; Original operand: call    sub_2425E
L_096F: ; original 0x2428f; IDA unnamed; logical 0x1428f; file 0x155ff
    call L_093E
; Original operand: cmp     byte ptr [bx+148h], 7Fh
L_0972: ; original 0x24292; IDA unnamed; logical 0x14292; file 0x15602
    cmp BYTE PTR [bx+0148h], 07Fh
; Original operand: ja      short loc_242B3
L_0977: ; original 0x24297; IDA unnamed; logical 0x14297; file 0x15607
    ja SHORT L_0993
; Original operand: shl     bx, 1
L_0979: ; original 0x24299; IDA unnamed; logical 0x14299; file 0x15609
    shl bx, 01h
; Original operand: mov     dx, [bx+18Ah]
L_097B: ; original 0x2429b; IDA unnamed; logical 0x1429b; file 0x1560b
    mov dx, WORD PTR [bx+018Ah]
; Original operand: shr     bx, 1
L_097F: ; original 0x2429f; IDA unnamed; logical 0x1429f; file 0x1560f
    shr bx, 01h
; Original operand: mov     ah, 0A0h
L_0981: ; original 0x242a1; IDA unnamed; logical 0x142a1; file 0x15611
    mov ah, 0A0h
; Original operand: add     ah, bl
L_0983: ; original 0x242a3; IDA unnamed; logical 0x142a3; file 0x15613
    add ah, bl
; Original operand: mov     al, dl
L_0985: ; original 0x242a5; IDA unnamed; logical 0x142a5; file 0x15615
    mov al, dl
; Original operand: call    sub_2425E
L_0987: ; original 0x242a7; IDA unnamed; logical 0x142a7; file 0x15617
    call L_093E
; Original operand: mov     ah, 0B0h
L_098A: ; original 0x242aa; IDA unnamed; logical 0x142aa; file 0x1561a
    mov ah, 0B0h
; Original operand: add     ah, bl
L_098C: ; original 0x242ac; IDA unnamed; logical 0x142ac; file 0x1561c
    add ah, bl
; Original operand: mov     al, dh
L_098E: ; original 0x242ae; IDA unnamed; logical 0x142ae; file 0x1561e
    mov al, dh
; Original operand: call    sub_2425E
L_0990: ; original 0x242b0; IDA unnamed; logical 0x142b0; file 0x15620
    call L_093E
; Original operand: dec     bl
L_0993: ; original 0x242b3; IDA loc_242B3; logical 0x142b3; file 0x15623
    dec bl
; Original operand: jns     short loc_24289
L_0995: ; original 0x242b5; IDA unnamed; logical 0x142b5; file 0x15625
    jns SHORT L_0969
; Original operand: cmp     byte ptr ds:1E7h, 0
L_0997: ; original 0x242b7; IDA unnamed; logical 0x142b7; file 0x15627
    cmp BYTE PTR ds:[01E7h], 00h
; Original operand: jz      short loc_242CB
L_099C: ; original 0x242bc; IDA unnamed; logical 0x142bc; file 0x1562c
    jz SHORT L_09AB
; Original operand: and     byte ptr ds:1E6h, 0E0h
L_099E: ; original 0x242be; IDA unnamed; logical 0x142be; file 0x1562e
    and BYTE PTR ds:[01E6h], 0E0h
; Original operand: mov     al, ds:1E6h
L_09A3: ; original 0x242c3; IDA unnamed; logical 0x142c3; file 0x15633
    mov al, BYTE PTR ds:[01E6h]
; Original operand: mov     ah, 0BDh
L_09A6: ; original 0x242c6; IDA unnamed; logical 0x142c6; file 0x15636
    mov ah, 0BDh
; Original operand: call    sub_2425E
L_09A8: ; original 0x242c8; IDA unnamed; logical 0x142c8; file 0x15638
    call L_093E
; Original operand: popf
L_09AB: ; original 0x242cb; IDA loc_242CB; logical 0x142cb; file 0x1563b
    popf 
; Original operand: retn
L_09AC: ; original 0x242cc; IDA unnamed; logical 0x142cc; file 0x1563c
    ret 
; Original operand: push    es
L_09AD: ; original 0x242cd; IDA sub_242CD; logical 0x142cd; file 0x1563d
    push es
; Original operand: push    di
L_09AE: ; original 0x242ce; IDA unnamed; logical 0x142ce; file 0x1563e
    push di
; Original operand: push    si
L_09AF: ; original 0x242cf; IDA unnamed; logical 0x142cf; file 0x1563f
    push si
; Original operand: mov     cx, cs
L_09B0: ; original 0x242d0; IDA unnamed; logical 0x142d0; file 0x15640
    mov cx, cs
; Original operand: mov     es, cx
L_09B2: ; original 0x242d2; IDA unnamed; logical 0x142d2; file 0x15642
    mov es, cx
; Original operand: mov     cx, 10h
L_09B4: ; original 0x242d4; IDA unnamed; logical 0x142d4; file 0x15644
    mov cx, 010h
; Original operand: mov     di, 1E9h
L_09B7: ; original 0x242d7; IDA unnamed; logical 0x142d7; file 0x15647
    mov di, 01E9h
; Original operand: sub     al, al
L_09BA: ; original 0x242da; IDA unnamed; logical 0x142da; file 0x1564a
    sub al, al
; Original operand: rep stosb
L_09BC: ; original 0x242dc; IDA unnamed; logical 0x142dc; file 0x1564c
    rep stosb
; Original operand: mov     cx, 0Bh
L_09BE: ; original 0x242de; IDA unnamed; logical 0x142de; file 0x1564e
    mov cx, 0Bh
; Original operand: mov     al, 0FFh
L_09C1: ; original 0x242e1; IDA unnamed; logical 0x142e1; file 0x15651
    mov al, 0FFh
; Original operand: mov     di, 148h
L_09C3: ; original 0x242e3; IDA unnamed; logical 0x142e3; file 0x15653
    mov di, 0148h
; Original operand: rep stosb
L_09C6: ; original 0x242e6; IDA unnamed; logical 0x142e6; file 0x15656
    rep stosb
; Original operand: mov     di, 1BDh
L_09C8: ; original 0x242e8; IDA unnamed; logical 0x142e8; file 0x15658
    mov di, 01BDh
; Original operand: mov     bl, 0
L_09CB: ; original 0x242eb; IDA unnamed; logical 0x142eb; file 0x1565b
    mov bl, 00h
; Original operand: call    sub_2450E
L_09CD: ; original 0x242ed; IDA loc_242ED; logical 0x142ed; file 0x1565d
    call L_0BEE
; Original operand: mov     ax, 800h
L_09D0: ; original 0x242f0; IDA unnamed; logical 0x142f0; file 0x15660
    mov ax, 0800h
; Original operand: call    sub_2425E
L_09D3: ; original 0x242f3; IDA unnamed; logical 0x142f3; file 0x15663
    call L_093E
; Original operand: mov     ah, [di]
L_09D6: ; original 0x242f6; IDA unnamed; logical 0x142f6; file 0x15666
    mov ah, BYTE PTR [di]
; Original operand: mov     cx, 4
L_09D8: ; original 0x242f8; IDA unnamed; logical 0x142f8; file 0x15668
    mov cx, 04h
; Original operand: mov     si, 1C6h
L_09DB: ; original 0x242fb; IDA unnamed; logical 0x142fb; file 0x1566b
    mov si, 01C6h
; Original operand: add     ah, 20h ; ' '
L_09DE: ; original 0x242fe; IDA loc_242FE; logical 0x142fe; file 0x1566e
    add ah, 020h
; Original operand: lodsb
L_09E1: ; original 0x24301; IDA unnamed; logical 0x14301; file 0x15671
    lodsb
; Original operand: call    sub_2425E
L_09E2: ; original 0x24302; IDA unnamed; logical 0x14302; file 0x15672
    call L_093E
; Original operand: add     ah, 3
L_09E5: ; original 0x24305; IDA unnamed; logical 0x14305; file 0x15675
    add ah, 03h
; Original operand: lodsb
L_09E8: ; original 0x24308; IDA unnamed; logical 0x14308; file 0x15678
    lodsb
; Original operand: call    sub_2425E
L_09E9: ; original 0x24309; IDA unnamed; logical 0x14309; file 0x15679
    call L_093E
; Original operand: sub     ah, 3
L_09EC: ; original 0x2430c; IDA unnamed; logical 0x1430c; file 0x1567c
    sub ah, 03h
; Original operand: loop    loc_242FE
L_09EF: ; original 0x2430f; IDA unnamed; logical 0x1430f; file 0x1567f
    loop L_09DE
; Original operand: add     ah, 60h ; '`'
L_09F1: ; original 0x24311; IDA unnamed; logical 0x14311; file 0x15681
    add ah, 060h
; Original operand: lodsb
L_09F4: ; original 0x24314; IDA unnamed; logical 0x14314; file 0x15684
    lodsb
; Original operand: call    sub_2425E
L_09F5: ; original 0x24315; IDA unnamed; logical 0x14315; file 0x15685
    call L_093E
; Original operand: add     ah, 3
L_09F8: ; original 0x24318; IDA unnamed; logical 0x14318; file 0x15688
    add ah, 03h
; Original operand: lodsb
L_09FB: ; original 0x2431b; IDA unnamed; logical 0x1431b; file 0x1568b
    lodsb
; Original operand: call    sub_2425E
L_09FC: ; original 0x2431c; IDA unnamed; logical 0x1431c; file 0x1568c
    call L_093E
; Original operand: mov     ah, [di]
L_09FF: ; original 0x2431f; IDA unnamed; logical 0x1431f; file 0x1568f
    mov ah, BYTE PTR [di]
; Original operand: add     ah, bl
L_0A01: ; original 0x24321; IDA unnamed; logical 0x14321; file 0x15691
    add ah, bl
; Original operand: lodsb
L_0A03: ; original 0x24323; IDA unnamed; logical 0x14323; file 0x15693
    lodsb
; Original operand: call    sub_2425E
L_0A04: ; original 0x24324; IDA unnamed; logical 0x14324; file 0x15694
    call L_093E
; Original operand: inc     di
L_0A07: ; original 0x24327; IDA unnamed; logical 0x14327; file 0x15697
    inc di
; Original operand: inc     bl
L_0A08: ; original 0x24328; IDA unnamed; logical 0x14328; file 0x15698
    inc bl
; Original operand: cmp     bl, 9
L_0A0A: ; original 0x2432a; IDA unnamed; logical 0x1432a; file 0x1569a
    cmp bl, 09h
; Original operand: jb      short loc_242ED
L_0A0D: ; original 0x2432d; IDA unnamed; logical 0x1432d; file 0x1569d
    jb SHORT L_09CD
; Original operand: pop     si
L_0A0F: ; original 0x2432f; IDA unnamed; logical 0x1432f; file 0x1569f
    pop si
; Original operand: pop     di
L_0A10: ; original 0x24330; IDA unnamed; logical 0x14330; file 0x156a0
    pop di
; Original operand: pop     es
L_0A11: ; original 0x24331; IDA unnamed; logical 0x14331; file 0x156a1
    pop es
; Original operand: retn
L_0A12: ; original 0x24332; IDA unnamed; logical 0x14332; file 0x156a2
    ret 
; Original operand: push    ds
L_0A13: ; original 0x24333; IDA sub_24333; logical 0x14333; file 0x156a3
    push ds
; Original operand: push    bx
L_0A14: ; original 0x24334; IDA unnamed; logical 0x14334; file 0x156a4
    push bx
; Original operand: lds     bx, ds:26h
L_0A15: ; original 0x24335; IDA unnamed; logical 0x14335; file 0x156a5
    lds bx, DWORD PTR ds:[026h]
; Original operand: mov     [bx], al
L_0A19: ; original 0x24339; IDA unnamed; logical 0x14339; file 0x156a9
    mov BYTE PTR [bx], al
; Original operand: pop     bx
L_0A1B: ; original 0x2433b; IDA unnamed; logical 0x1433b; file 0x156ab
    pop bx
; Original operand: pop     ds
L_0A1C: ; original 0x2433c; IDA unnamed; logical 0x1433c; file 0x156ac
    pop ds
; Original operand: retn
L_0A1D: ; original 0x2433d; IDA unnamed; logical 0x1433d; file 0x156ad
    ret 
; Original operand: push    es
L_0A1E: ; original 0x2433e; IDA sub_2433E; logical 0x1433e; file 0x156ae
    push es
; Original operand: push    cx
L_0A1F: ; original 0x2433f; IDA unnamed; logical 0x1433f; file 0x156af
    push cx
; Original operand: push    di
L_0A20: ; original 0x24340; IDA unnamed; logical 0x14340; file 0x156b0
    push di
; Original operand: cmp     al, ds:1E0h
L_0A21: ; original 0x24341; IDA unnamed; logical 0x14341; file 0x156b1
    cmp al, BYTE PTR ds:[01E0h]
; Original operand: jb      short loc_2434A
L_0A25: ; original 0x24345; IDA unnamed; logical 0x14345; file 0x156b5
    jb SHORT L_0A2A
; Original operand: jmp     loc_2443B
L_0A27: ; original 0x24347; IDA unnamed; logical 0x14347; file 0x156b7
    jmp NEAR PTR L_0B1B
; Original operand: cbw
L_0A2A: ; original 0x2434a; IDA loc_2434A; logical 0x1434a; file 0x156ba
    cbw 
; Original operand: shl     ax, 1
L_0A2B: ; original 0x2434b; IDA unnamed; logical 0x1434b; file 0x156bb
    shl ax, 01h
; Original operand: shl     ax, 1
L_0A2D: ; original 0x2434d; IDA unnamed; logical 0x1434d; file 0x156bd
    shl ax, 01h
; Original operand: shl     ax, 1
L_0A2F: ; original 0x2434f; IDA unnamed; logical 0x1434f; file 0x156bf
    shl ax, 01h
; Original operand: shl     ax, 1
L_0A31: ; original 0x24351; IDA unnamed; logical 0x14351; file 0x156c1
    shl ax, 01h
; Original operand: les     di, ds:16h
L_0A33: ; original 0x24353; IDA unnamed; logical 0x14353; file 0x156c3
    les di, DWORD PTR ds:[016h]
; Original operand: add     di, ax
L_0A37: ; original 0x24357; IDA unnamed; logical 0x14357; file 0x156c7
    add di, ax
; Original operand: mov     al, es:[di+3]
L_0A39: ; original 0x24359; IDA unnamed; logical 0x14359; file 0x156c9
    mov al, BYTE PTR es:[di+03h]
; Original operand: cmp     byte ptr ds:1E7h, 0
L_0A3D: ; original 0x2435d; IDA unnamed; logical 0x1435d; file 0x156cd
    cmp BYTE PTR ds:[01E7h], 00h
; Original operand: jz      short loc_2436D
L_0A42: ; original 0x24362; IDA unnamed; logical 0x14362; file 0x156d2
    jz SHORT L_0A4D
; Original operand: cmp     bl, 7
L_0A44: ; original 0x24364; IDA unnamed; logical 0x14364; file 0x156d4
    cmp bl, 07h
; Original operand: jb      short loc_2436D
L_0A47: ; original 0x24367; IDA unnamed; logical 0x14367; file 0x156d7
    jb SHORT L_0A4D
; Original operand: mov     al, es:[di+2]
L_0A49: ; original 0x24369; IDA unnamed; logical 0x14369; file 0x156d9
    mov al, BYTE PTR es:[di+02h]
; Original operand: mov     ah, al
L_0A4D: ; original 0x2436d; IDA loc_2436D; logical 0x1436d; file 0x156dd
    mov ah, al
; Original operand: and     ax, 0C03Fh
L_0A4F: ; original 0x2436f; IDA unnamed; logical 0x1436f; file 0x156df
    and ax, 0C03Fh
; Original operand: mov     [bx+169h], ah
L_0A52: ; original 0x24372; IDA unnamed; logical 0x14372; file 0x156e2
    mov BYTE PTR [bx+0169h], ah
; Original operand: sub     al, 3Fh ; '?'
L_0A56: ; original 0x24376; IDA unnamed; logical 0x14376; file 0x156e6
    sub al, 03Fh
; Original operand: neg     al
L_0A58: ; original 0x24378; IDA unnamed; logical 0x14378; file 0x156e8
    neg al
; Original operand: mov     [bx+17Fh], al
L_0A5A: ; original 0x2437a; IDA unnamed; logical 0x1437a; file 0x156ea
    mov BYTE PTR [bx+017Fh], al
; Original operand: mul     byte ptr ds:1B6h
L_0A5E: ; original 0x2437e; IDA unnamed; logical 0x1437e; file 0x156ee
    mul BYTE PTR ds:[01B6h]
; Original operand: add     al, al
L_0A62: ; original 0x24382; IDA unnamed; logical 0x14382; file 0x156f2
    add al, al
; Original operand: adc     ah, 0
L_0A64: ; original 0x24384; IDA unnamed; logical 0x14384; file 0x156f4
    adc ah, 00h
; Original operand: mov     [bx+174h], ah
L_0A67: ; original 0x24387; IDA unnamed; logical 0x14387; file 0x156f7
    mov BYTE PTR [bx+0174h], ah
; Original operand: cmp     byte ptr ds:1E7h, 0
L_0A6B: ; original 0x2438b; IDA unnamed; logical 0x1438b; file 0x156fb
    cmp BYTE PTR ds:[01E7h], 00h
; Original operand: jz      short loc_24397
L_0A70: ; original 0x24390; IDA unnamed; logical 0x14390; file 0x15700
    jz SHORT L_0A77
; Original operand: cmp     bx, 6
L_0A72: ; original 0x24392; IDA unnamed; logical 0x14392; file 0x15702
    cmp bx, 06h
; Original operand: ja      short loc_243FD
L_0A75: ; original 0x24395; IDA unnamed; logical 0x14395; file 0x15705
    ja SHORT L_0ADD
; Original operand: mov     ah, [bx+1BDh]
L_0A77: ; original 0x24397; IDA loc_24397; logical 0x14397; file 0x15707
    mov ah, BYTE PTR [bx+01BDh]
; Original operand: add     ah, 20h ; ' '
L_0A7B: ; original 0x2439b; IDA unnamed; logical 0x1439b; file 0x1570b
    add ah, 020h
; Original operand: mov     al, es:[di]
L_0A7E: ; original 0x2439e; IDA unnamed; logical 0x1439e; file 0x1570e
    mov al, BYTE PTR es:[di]
; Original operand: inc     di
L_0A81: ; original 0x243a1; IDA unnamed; logical 0x143a1; file 0x15711
    inc di
; Original operand: call    sub_2425E
L_0A82: ; original 0x243a2; IDA unnamed; logical 0x143a2; file 0x15712
    call L_093E
; Original operand: add     ah, 3
L_0A85: ; original 0x243a5; IDA unnamed; logical 0x143a5; file 0x15715
    add ah, 03h
; Original operand: mov     al, es:[di]
L_0A88: ; original 0x243a8; IDA unnamed; logical 0x143a8; file 0x15718
    mov al, BYTE PTR es:[di]
; Original operand: inc     di
L_0A8B: ; original 0x243ab; IDA unnamed; logical 0x143ab; file 0x1571b
    inc di
; Original operand: call    sub_2425E
L_0A8C: ; original 0x243ac; IDA unnamed; logical 0x143ac; file 0x1571c
    call L_093E
; Original operand: sub     ah, 3
L_0A8F: ; original 0x243af; IDA unnamed; logical 0x143af; file 0x1571f
    sub ah, 03h
; Original operand: add     ah, 20h ; ' '
L_0A92: ; original 0x243b2; IDA unnamed; logical 0x143b2; file 0x15722
    add ah, 020h
; Original operand: mov     al, es:[di]
L_0A95: ; original 0x243b5; IDA unnamed; logical 0x143b5; file 0x15725
    mov al, BYTE PTR es:[di]
; Original operand: inc     di
L_0A98: ; original 0x243b8; IDA unnamed; logical 0x143b8; file 0x15728
    inc di
; Original operand: call    sub_2425E
L_0A99: ; original 0x243b9; IDA unnamed; logical 0x143b9; file 0x15729
    call L_093E
; Original operand: inc     di
L_0A9C: ; original 0x243bc; IDA unnamed; logical 0x143bc; file 0x1572c
    inc di
; Original operand: mov     cx, 2
L_0A9D: ; original 0x243bd; IDA unnamed; logical 0x143bd; file 0x1572d
    mov cx, 02h
; Original operand: add     ah, 20h ; ' '
L_0AA0: ; original 0x243c0; IDA loc_243C0; logical 0x143c0; file 0x15730
    add ah, 020h
; Original operand: mov     al, es:[di]
L_0AA3: ; original 0x243c3; IDA unnamed; logical 0x143c3; file 0x15733
    mov al, BYTE PTR es:[di]
; Original operand: inc     di
L_0AA6: ; original 0x243c6; IDA unnamed; logical 0x143c6; file 0x15736
    inc di
; Original operand: call    sub_2425E
L_0AA7: ; original 0x243c7; IDA unnamed; logical 0x143c7; file 0x15737
    call L_093E
; Original operand: add     ah, 3
L_0AAA: ; original 0x243ca; IDA unnamed; logical 0x143ca; file 0x1573a
    add ah, 03h
; Original operand: mov     al, es:[di]
L_0AAD: ; original 0x243cd; IDA unnamed; logical 0x143cd; file 0x1573d
    mov al, BYTE PTR es:[di]
; Original operand: inc     di
L_0AB0: ; original 0x243d0; IDA unnamed; logical 0x143d0; file 0x15740
    inc di
; Original operand: call    sub_2425E
L_0AB1: ; original 0x243d1; IDA unnamed; logical 0x143d1; file 0x15741
    call L_093E
; Original operand: sub     ah, 3
L_0AB4: ; original 0x243d4; IDA unnamed; logical 0x143d4; file 0x15744
    sub ah, 03h
; Original operand: loop    loc_243C0
L_0AB7: ; original 0x243d7; IDA unnamed; logical 0x143d7; file 0x15747
    loop L_0AA0
; Original operand: add     ah, 60h ; '`'
L_0AB9: ; original 0x243d9; IDA unnamed; logical 0x143d9; file 0x15749
    add ah, 060h
; Original operand: mov     al, es:[di]
L_0ABC: ; original 0x243dc; IDA unnamed; logical 0x143dc; file 0x1574c
    mov al, BYTE PTR es:[di]
; Original operand: inc     di
L_0ABF: ; original 0x243df; IDA unnamed; logical 0x143df; file 0x1574f
    inc di
; Original operand: call    sub_2425E
L_0AC0: ; original 0x243e0; IDA unnamed; logical 0x143e0; file 0x15750
    call L_093E
; Original operand: add     ah, 3
L_0AC3: ; original 0x243e3; IDA unnamed; logical 0x143e3; file 0x15753
    add ah, 03h
; Original operand: mov     al, es:[di]
L_0AC6: ; original 0x243e6; IDA unnamed; logical 0x143e6; file 0x15756
    mov al, BYTE PTR es:[di]
; Original operand: inc     di
L_0AC9: ; original 0x243e9; IDA unnamed; logical 0x143e9; file 0x15759
    inc di
; Original operand: call    sub_2425E
L_0ACA: ; original 0x243ea; IDA unnamed; logical 0x143ea; file 0x1575a
    call L_093E
; Original operand: sub     ah, 3
L_0ACD: ; original 0x243ed; IDA unnamed; logical 0x143ed; file 0x1575d
    sub ah, 03h
; Original operand: mov     ah, bl
L_0AD0: ; original 0x243f0; IDA unnamed; logical 0x143f0; file 0x15760
    mov ah, bl
; Original operand: add     ah, 0C0h
L_0AD2: ; original 0x243f2; IDA unnamed; logical 0x143f2; file 0x15762
    add ah, 0C0h
; Original operand: mov     al, es:[di]
L_0AD5: ; original 0x243f5; IDA unnamed; logical 0x143f5; file 0x15765
    mov al, BYTE PTR es:[di]
; Original operand: call    sub_2425E
L_0AD8: ; original 0x243f8; IDA unnamed; logical 0x143f8; file 0x15768
    call L_093E
; Original operand: jmp     short loc_2443B
L_0ADB: ; original 0x243fb; IDA unnamed; logical 0x143fb; file 0x1576b
    jmp SHORT L_0B1B
; Original operand: mov     ah, [bx+1CBh]
L_0ADD: ; original 0x243fd; IDA loc_243FD; logical 0x143fd; file 0x1576d
    mov ah, BYTE PTR [bx+01CBh]
; Original operand: add     ah, 20h ; ' '
L_0AE1: ; original 0x24401; IDA unnamed; logical 0x14401; file 0x15771
    add ah, 020h
; Original operand: mov     al, es:[di]
L_0AE4: ; original 0x24404; IDA unnamed; logical 0x14404; file 0x15774
    mov al, BYTE PTR es:[di]
; Original operand: inc     di
L_0AE7: ; original 0x24407; IDA unnamed; logical 0x14407; file 0x15777
    inc di
; Original operand: inc     di
L_0AE8: ; original 0x24408; IDA unnamed; logical 0x14408; file 0x15778
    inc di
; Original operand: call    sub_2425E
L_0AE9: ; original 0x24409; IDA unnamed; logical 0x14409; file 0x15779
    call L_093E
; Original operand: add     ah, 20h ; ' '
L_0AEC: ; original 0x2440c; IDA unnamed; logical 0x1440c; file 0x1577c
    add ah, 020h
; Original operand: inc     di
L_0AEF: ; original 0x2440f; IDA unnamed; logical 0x1440f; file 0x1577f
    inc di
; Original operand: inc     di
L_0AF0: ; original 0x24410; IDA unnamed; logical 0x14410; file 0x15780
    inc di
; Original operand: mov     cx, 2
L_0AF1: ; original 0x24411; IDA unnamed; logical 0x14411; file 0x15781
    mov cx, 02h
; Original operand: add     ah, 20h ; ' '
L_0AF4: ; original 0x24414; IDA loc_24414; logical 0x14414; file 0x15784
    add ah, 020h
; Original operand: mov     al, es:[di]
L_0AF7: ; original 0x24417; IDA unnamed; logical 0x14417; file 0x15787
    mov al, BYTE PTR es:[di]
; Original operand: inc     di
L_0AFA: ; original 0x2441a; IDA unnamed; logical 0x1441a; file 0x1578a
    inc di
; Original operand: inc     di
L_0AFB: ; original 0x2441b; IDA unnamed; logical 0x1441b; file 0x1578b
    inc di
; Original operand: call    sub_2425E
L_0AFC: ; original 0x2441c; IDA unnamed; logical 0x1441c; file 0x1578c
    call L_093E
; Original operand: loop    loc_24414
L_0AFF: ; original 0x2441f; IDA unnamed; logical 0x1441f; file 0x1578f
    loop L_0AF4
; Original operand: add     ah, 60h ; '`'
L_0B01: ; original 0x24421; IDA unnamed; logical 0x14421; file 0x15791
    add ah, 060h
; Original operand: mov     al, es:[di]
L_0B04: ; original 0x24424; IDA unnamed; logical 0x14424; file 0x15794
    mov al, BYTE PTR es:[di]
; Original operand: inc     di
L_0B07: ; original 0x24427; IDA unnamed; logical 0x14427; file 0x15797
    inc di
; Original operand: inc     di
L_0B08: ; original 0x24428; IDA unnamed; logical 0x14428; file 0x15798
    inc di
; Original operand: call    sub_2425E
L_0B09: ; original 0x24429; IDA unnamed; logical 0x14429; file 0x15799
    call L_093E
; Original operand: mov     ah, [bx+1D5h]
L_0B0C: ; original 0x2442c; IDA unnamed; logical 0x1442c; file 0x1579c
    mov ah, BYTE PTR [bx+01D5h]
; Original operand: add     ah, 0C0h
L_0B10: ; original 0x24430; IDA unnamed; logical 0x14430; file 0x157a0
    add ah, 0C0h
; Original operand: mov     al, es:[di]
L_0B13: ; original 0x24433; IDA unnamed; logical 0x14433; file 0x157a3
    mov al, BYTE PTR es:[di]
; Original operand: call    sub_2425E
L_0B16: ; original 0x24436; IDA unnamed; logical 0x14436; file 0x157a6
    call L_093E
; Original operand: jmp     short $+2
L_0B19: ; original 0x24439; IDA unnamed; logical 0x14439; file 0x157a9
    jmp SHORT L_0B1B
; Original operand: pop     di
L_0B1B: ; original 0x2443b; IDA loc_2443B; logical 0x1443b; file 0x157ab
    pop di
; Original operand: pop     cx
L_0B1C: ; original 0x2443c; IDA unnamed; logical 0x1443c; file 0x157ac
    pop cx
; Original operand: pop     es
L_0B1D: ; original 0x2443d; IDA unnamed; logical 0x1443d; file 0x157ad
    pop es
; Original operand: retn
L_0B1E: ; original 0x2443e; IDA unnamed; logical 0x1443e; file 0x157ae
    ret 
; Original operand: push    si
L_0B1F: ; original 0x2443f; IDA sub_2443F; logical 0x1443f; file 0x157af
    push si
; Original operand: push    bx
L_0B20: ; original 0x24440; IDA unnamed; logical 0x14440; file 0x157b0
    push bx
; Original operand: push    ds
L_0B21: ; original 0x24441; IDA unnamed; logical 0x14441; file 0x157b1
    push ds
; Original operand: sub     bx, bx
L_0B22: ; original 0x24442; IDA unnamed; logical 0x14442; file 0x157b2
    sub bx, bx
; Original operand: sub     dx, dx
L_0B24: ; original 0x24444; IDA unnamed; logical 0x14444; file 0x157b4
    sub dx, dx
; Original operand: lds     si, ds:1Ah
L_0B26: ; original 0x24446; IDA unnamed; logical 0x14446; file 0x157b6
    lds si, DWORD PTR ds:[01Ah]
; Original operand: lodsb
L_0B2A: ; original 0x2444a; IDA loc_2444A; logical 0x1444a; file 0x157ba
    lodsb
; Original operand: push    ax
L_0B2B: ; original 0x2444b; IDA unnamed; logical 0x1444b; file 0x157bb
    push ax
; Original operand: and     al, 7Fh
L_0B2C: ; original 0x2444c; IDA unnamed; logical 0x1444c; file 0x157bc
    and al, 07Fh
; Original operand: cbw
L_0B2E: ; original 0x2444e; IDA unnamed; logical 0x1444e; file 0x157be
    cbw 
; Original operand: add     bx, ax
L_0B2F: ; original 0x2444f; IDA unnamed; logical 0x1444f; file 0x157bf
    add bx, ax
; Original operand: adc     dx, 0
L_0B31: ; original 0x24451; IDA unnamed; logical 0x14451; file 0x157c1
    adc dx, 00h
; Original operand: pop     ax
L_0B34: ; original 0x24454; IDA unnamed; logical 0x14454; file 0x157c4
    pop ax
; Original operand: or      al, al
L_0B35: ; original 0x24455; IDA unnamed; logical 0x14455; file 0x157c5
    or al, al
; Original operand: jns     short loc_24465
L_0B37: ; original 0x24457; IDA unnamed; logical 0x14457; file 0x157c7
    jns SHORT L_0B45
; Original operand: mov     al, 7
L_0B39: ; original 0x24459; IDA unnamed; logical 0x14459; file 0x157c9
    mov al, 07h
; Original operand: shl     bx, 1
L_0B3B: ; original 0x2445b; IDA loc_2445B; logical 0x1445b; file 0x157cb
    shl bx, 01h
; Original operand: rcl     dx, 1
L_0B3D: ; original 0x2445d; IDA unnamed; logical 0x1445d; file 0x157cd
    rcl dx, 01h
; Original operand: dec     al
L_0B3F: ; original 0x2445f; IDA unnamed; logical 0x1445f; file 0x157cf
    dec al
; Original operand: jnz     short loc_2445B
L_0B41: ; original 0x24461; IDA unnamed; logical 0x14461; file 0x157d1
    jnz SHORT L_0B3B
; Original operand: jmp     short loc_2444A
L_0B43: ; original 0x24463; IDA unnamed; logical 0x14463; file 0x157d3
    jmp SHORT L_0B2A
; Original operand: pop     ds
L_0B45: ; original 0x24465; IDA loc_24465; logical 0x14465; file 0x157d5
    pop ds
; Original operand: mov     ax, bx
L_0B46: ; original 0x24466; IDA unnamed; logical 0x14466; file 0x157d6
    mov ax, bx
; Original operand: mov     ds:1Ah, si
L_0B48: ; original 0x24468; IDA unnamed; logical 0x14468; file 0x157d8
    mov WORD PTR ds:[01Ah], si
; Original operand: pop     bx
L_0B4C: ; original 0x2446c; IDA unnamed; logical 0x1446c; file 0x157dc
    pop bx
; Original operand: pop     si
L_0B4D: ; original 0x2446d; IDA unnamed; logical 0x1446d; file 0x157dd
    pop si
; Original operand: retn
L_0B4E: ; original 0x2446e; IDA unnamed; logical 0x1446e; file 0x157de
    ret 
; Original operand: push    cx
L_0B4F: ; original 0x2446f; IDA sub_2446F; logical 0x1446f; file 0x157df
    push cx
; Original operand: add     ds:1Ah, ax
L_0B50: ; original 0x24470; IDA unnamed; logical 0x14470; file 0x157e0
    add WORD PTR ds:[01Ah], ax
; Original operand: jnb     short loc_2447C
L_0B54: ; original 0x24474; IDA unnamed; logical 0x14474; file 0x157e4
    jnb SHORT L_0B5C
; Original operand: add     word ptr ds:1Ch, 1000h
L_0B56: ; original 0x24476; IDA unnamed; logical 0x14476; file 0x157e6
    add WORD PTR ds:[01Ch], 01000h
; Original operand: mov     cl, 4
L_0B5C: ; original 0x2447c; IDA loc_2447C; logical 0x1447c; file 0x157ec
    mov cl, 04h
; Original operand: mov     dh, dl
L_0B5E: ; original 0x2447e; IDA unnamed; logical 0x1447e; file 0x157ee
    mov dh, dl
; Original operand: sub     dl, dl
L_0B60: ; original 0x24480; IDA unnamed; logical 0x14480; file 0x157f0
    sub dl, dl
; Original operand: shl     dx, cl
L_0B62: ; original 0x24482; IDA unnamed; logical 0x14482; file 0x157f2
    shl dx, cl
; Original operand: add     ds:1Ch, dx
L_0B64: ; original 0x24484; IDA unnamed; logical 0x14484; file 0x157f4
    add WORD PTR ds:[01Ch], dx
; Original operand: mov     si, ds:1Ah
L_0B68: ; original 0x24488; IDA unnamed; logical 0x14488; file 0x157f8
    mov si, WORD PTR ds:[01Ah]
; Original operand: pop     cx
L_0B6C: ; original 0x2448c; IDA unnamed; logical 0x1448c; file 0x157fc
    pop cx
; Original operand: retn
L_0B6D: ; original 0x2448d; IDA unnamed; logical 0x1448d; file 0x157fd
    ret 
; Original operand: push    es
L_0B6E: ; original 0x2448e; IDA sub_2448E; logical 0x1448e; file 0x157fe
    push es
; Original operand: mov     cx, ds
L_0B6F: ; original 0x2448f; IDA unnamed; logical 0x1448f; file 0x157ff
    mov cx, ds
; Original operand: mov     es, cx
L_0B71: ; original 0x24491; IDA unnamed; logical 0x14491; file 0x15801
    mov es, cx
; Original operand: mov     cx, ds:3Ah
L_0B73: ; original 0x24493; IDA unnamed; logical 0x14493; file 0x15803
    mov cx, WORD PTR ds:[03Ah]
; Original operand: mov     ah, al
L_0B77: ; original 0x24497; IDA unnamed; logical 0x14497; file 0x15807
    mov ah, al
; Original operand: or      al, 80h
L_0B79: ; original 0x24499; IDA unnamed; logical 0x14499; file 0x15809
    or al, 080h
; Original operand: mov     di, 148h
L_0B7B: ; original 0x2449b; IDA unnamed; logical 0x1449b; file 0x1580b
    mov di, 0148h
; Original operand: repne scasb
L_0B7E: ; original 0x2449e; IDA unnamed; logical 0x1449e; file 0x1580e
    repne scasb
; Original operand: jz      short loc_24506
L_0B80: ; original 0x244a0; IDA unnamed; logical 0x144a0; file 0x15810
    jz SHORT L_0BE6
; Original operand: mov     cx, ds:3Ah
L_0B82: ; original 0x244a2; IDA unnamed; logical 0x144a2; file 0x15812
    mov cx, WORD PTR ds:[03Ah]
; Original operand: mov     al, 0FFh
L_0B86: ; original 0x244a6; IDA unnamed; logical 0x144a6; file 0x15816
    mov al, 0FFh
; Original operand: mov     di, 148h
L_0B88: ; original 0x244a8; IDA unnamed; logical 0x144a8; file 0x15818
    mov di, 0148h
; Original operand: repne scasb
L_0B8B: ; original 0x244ab; IDA unnamed; logical 0x144ab; file 0x1581b
    repne scasb
; Original operand: jz      short loc_24506
L_0B8D: ; original 0x244ad; IDA unnamed; logical 0x144ad; file 0x1581d
    jz SHORT L_0BE6
; Original operand: mov     cx, ds:3Ah
L_0B8F: ; original 0x244af; IDA unnamed; logical 0x144af; file 0x1581f
    mov cx, WORD PTR ds:[03Ah]
; Original operand: mov     al, 7Fh
L_0B93: ; original 0x244b3; IDA unnamed; logical 0x144b3; file 0x15823
    mov al, 07Fh
; Original operand: mov     di, 148h
L_0B95: ; original 0x244b5; IDA unnamed; logical 0x144b5; file 0x15825
    mov di, 0148h
; Original operand: cmp     al, [di]
L_0B98: ; original 0x244b8; IDA loc_244B8; logical 0x144b8; file 0x15828
    cmp al, BYTE PTR [di]
; Original operand: jb      short loc_244C1
L_0B9A: ; original 0x244ba; IDA unnamed; logical 0x144ba; file 0x1582a
    jb SHORT L_0BA1
; Original operand: inc     di
L_0B9C: ; original 0x244bc; IDA unnamed; logical 0x144bc; file 0x1582c
    inc di
; Original operand: loop    loc_244B8
L_0B9D: ; original 0x244bd; IDA unnamed; logical 0x144bd; file 0x1582d
    loop L_0B98
; Original operand: jmp     short loc_244C4
L_0B9F: ; original 0x244bf; IDA unnamed; logical 0x144bf; file 0x1582f
    jmp SHORT L_0BA4
; Original operand: inc     di
L_0BA1: ; original 0x244c1; IDA loc_244C1; logical 0x144c1; file 0x15831
    inc di
; Original operand: jmp     short loc_24506
L_0BA2: ; original 0x244c2; IDA unnamed; logical 0x144c2; file 0x15832
    jmp SHORT L_0BE6
; Original operand: push    si
L_0BA4: ; original 0x244c4; IDA loc_244C4; logical 0x144c4; file 0x15834
    push si
; Original operand: sub     si, si
L_0BA5: ; original 0x244c5; IDA unnamed; logical 0x144c5; file 0x15835
    sub si, si
; Original operand: mov     cx, ds:3Ah
L_0BA7: ; original 0x244c7; IDA unnamed; logical 0x144c7; file 0x15837
    mov cx, WORD PTR ds:[03Ah]
; Original operand: mov     di, 1A0h
L_0BAB: ; original 0x244cb; IDA unnamed; logical 0x144cb; file 0x1583b
    mov di, 01A0h
; Original operand: mov     ax, di
L_0BAE: ; original 0x244ce; IDA unnamed; logical 0x144ce; file 0x1583e
    mov ax, di
; Original operand: mov     dx, [di]
L_0BB0: ; original 0x244d0; IDA loc_244D0; logical 0x144d0; file 0x15840
    mov dx, WORD PTR [di]
; Original operand: sub     dx, ds:40h
L_0BB2: ; original 0x244d2; IDA unnamed; logical 0x144d2; file 0x15842
    sub dx, WORD PTR ds:[040h]
; Original operand: neg     dx
L_0BB6: ; original 0x244d6; IDA unnamed; logical 0x144d6; file 0x15846
    neg dx
; Original operand: cmp     dx, si
L_0BB8: ; original 0x244d8; IDA unnamed; logical 0x144d8; file 0x15848
    cmp dx, si
; Original operand: jbe     short loc_244E0
L_0BBA: ; original 0x244da; IDA unnamed; logical 0x144da; file 0x1584a
    jbe SHORT L_0BC0
; Original operand: mov     si, dx
L_0BBC: ; original 0x244dc; IDA unnamed; logical 0x144dc; file 0x1584c
    mov si, dx
; Original operand: mov     ax, di
L_0BBE: ; original 0x244de; IDA unnamed; logical 0x144de; file 0x1584e
    mov ax, di
; Original operand: inc     di
L_0BC0: ; original 0x244e0; IDA loc_244E0; logical 0x144e0; file 0x15850
    inc di
; Original operand: inc     di
L_0BC1: ; original 0x244e1; IDA unnamed; logical 0x144e1; file 0x15851
    inc di
; Original operand: dec     cx
L_0BC2: ; original 0x244e2; IDA unnamed; logical 0x144e2; file 0x15852
    dec cx
; Original operand: jnz     short loc_244D0
L_0BC3: ; original 0x244e3; IDA unnamed; logical 0x144e3; file 0x15853
    jnz SHORT L_0BB0
; Original operand: sub     ax, 1A0h
L_0BC5: ; original 0x244e5; IDA unnamed; logical 0x144e5; file 0x15855
    sub ax, IMM_01A0
; Original operand: mov     bx, ax
L_0BC8: ; original 0x244e8; IDA unnamed; logical 0x144e8; file 0x15858
    mov bx, ax
; Original operand: mov     ax, [bx+18Ah]
L_0BCA: ; original 0x244ea; IDA unnamed; logical 0x144ea; file 0x1585a
    mov ax, WORD PTR [bx+018Ah]
; Original operand: shr     bx, 1
L_0BCE: ; original 0x244ee; IDA unnamed; logical 0x144ee; file 0x1585e
    shr bx, 01h
; Original operand: mov     dh, 0A0h
L_0BD0: ; original 0x244f0; IDA unnamed; logical 0x144f0; file 0x15860
    mov dh, 0A0h
; Original operand: xchg    dh, ah
L_0BD2: ; original 0x244f2; IDA unnamed; logical 0x144f2; file 0x15862
    xchg ah, dh
; Original operand: add     ah, bl
L_0BD4: ; original 0x244f4; IDA unnamed; logical 0x144f4; file 0x15864
    add ah, bl
; Original operand: call    sub_2425E
L_0BD6: ; original 0x244f6; IDA unnamed; logical 0x144f6; file 0x15866
    call L_093E
; Original operand: add     ah, 10h
L_0BD9: ; original 0x244f9; IDA unnamed; logical 0x144f9; file 0x15869
    add ah, 010h
; Original operand: mov     al, dh
L_0BDC: ; original 0x244fc; IDA unnamed; logical 0x144fc; file 0x1586c
    mov al, dh
; Original operand: call    sub_2425E
L_0BDE: ; original 0x244fe; IDA unnamed; logical 0x144fe; file 0x1586e
    call L_093E
; Original operand: mov     ax, bx
L_0BE1: ; original 0x24501; IDA unnamed; logical 0x14501; file 0x15871
    mov ax, bx
; Original operand: pop     si
L_0BE3: ; original 0x24503; IDA unnamed; logical 0x14503; file 0x15873
    pop si
; Original operand: jmp     short loc_2450C
L_0BE4: ; original 0x24504; IDA unnamed; logical 0x14504; file 0x15874
    jmp SHORT L_0BEC
; Original operand: sub     di, 149h
L_0BE6: ; original 0x24506; IDA loc_24506; logical 0x14506; file 0x15876
    sub di, 0149h
; Original operand: mov     ax, di
L_0BEA: ; original 0x2450a; IDA unnamed; logical 0x1450a; file 0x1587a
    mov ax, di
; Original operand: pop     es
L_0BEC: ; original 0x2450c; IDA loc_2450C; logical 0x1450c; file 0x1587c
    pop es
; Original operand: retn
L_0BED: ; original 0x2450d; IDA unnamed; logical 0x1450d; file 0x1587d
    ret 
; Original operand: mov     word ptr ds:3Ah, 9
L_0BEE: ; original 0x2450e; IDA sub_2450E; logical 0x1450e; file 0x1587e
    mov WORD PTR ds:[03Ah], 09h
; Original operand: mov     ax, 0C0h
L_0BF4: ; original 0x24514; IDA unnamed; logical 0x14514; file 0x15884
    mov ax, 0C0h
; Original operand: mov     ds:1E6h, ax
L_0BF7: ; original 0x24517; IDA unnamed; logical 0x14517; file 0x15887
    mov WORD PTR ds:[01E6h], ax
; Original operand: mov     ah, 0BDh
L_0BFA: ; original 0x2451a; IDA unnamed; logical 0x1451a; file 0x1588a
    mov ah, 0BDh
; Original operand: call    sub_2425E
L_0BFC: ; original 0x2451c; IDA unnamed; logical 0x1451c; file 0x1588c
    call L_093E
; Original operand: retn
L_0BFF: ; original 0x2451f; IDA unnamed; logical 0x1451f; file 0x1588f
    ret 
; Original operand: mov     byte ptr ds:1B6h, 0FFh
L_0C00: ; original 0x24520; IDA sub_24520; logical 0x14520; file 0x15890
    mov BYTE PTR ds:[01B6h], 0FFh
; Original operand: mov     byte ptr ds:1B7h, 0FFh
L_0C05: ; original 0x24525; IDA unnamed; logical 0x14525; file 0x15895
    mov BYTE PTR ds:[01B7h], 0FFh
; Original operand: call    sub_24AB8
L_0C0A: ; original 0x2452a; IDA unnamed; logical 0x1452a; file 0x1589a
    call L_1198
; Original operand: call    sub_2450E
L_0C0D: ; original 0x2452d; IDA unnamed; logical 0x1452d; file 0x1589d
    call L_0BEE
; Original operand: mov     cx, 8
L_0C10: ; original 0x24530; IDA unnamed; logical 0x14530; file 0x158a0
    mov cx, 08h
; Original operand: mov     di, 219h
L_0C13: ; original 0x24533; IDA unnamed; logical 0x14533; file 0x158a3
    mov di, 0219h
; Original operand: mov     ax, 101h
L_0C16: ; original 0x24536; IDA unnamed; logical 0x14536; file 0x158a6
    mov ax, 0101h
; Original operand: rep stosw
L_0C19: ; original 0x24539; IDA unnamed; logical 0x14539; file 0x158a9
    rep stosw
; Original operand: mov     byte ptr ds:1E0h, 10h
L_0C1B: ; original 0x2453b; IDA unnamed; logical 0x1453b; file 0x158ab
    mov BYTE PTR ds:[01E0h], 010h
; Original operand: mov     word ptr ds:16h, 48h ; 'H'
L_0C20: ; original 0x24540; IDA unnamed; logical 0x14540; file 0x158b0
    mov WORD PTR ds:[016h], 048h
; Original operand: mov     word ptr ds:18h, ds
L_0C26: ; original 0x24546; IDA unnamed; logical 0x14546; file 0x158b6
    mov WORD PTR ds:[018h], ds
; Original operand: call    sub_242CD
L_0C2A: ; original 0x2454a; IDA unnamed; logical 0x1454a; file 0x158ba
    call L_09AD
; Original operand: mov     word ptr ds:34h, 48D3h
L_0C2D: ; original 0x2454d; IDA unnamed; logical 0x1454d; file 0x158bd
    mov WORD PTR ds:[034h], 048D3h
; Original operand: sub     ax, ax
L_0C33: ; original 0x24553; IDA unnamed; logical 0x14553; file 0x158c3
    sub ax, ax
; Original operand: mov     ds:42h, ax
L_0C35: ; original 0x24555; IDA unnamed; logical 0x14555; file 0x158c5
    mov WORD PTR ds:[042h], ax
; Original operand: retn
L_0C38: ; original 0x24558; IDA unnamed; logical 0x14558; file 0x158c8
    ret 
; Original operand: mov     cx, ax
L_0C39: ; original 0x24559; IDA sub_24559; logical 0x14559; file 0x158c9
    mov cx, ax
; Original operand: mov     ax, 0FFFEh
L_0C3B: ; original 0x2455b; IDA unnamed; logical 0x1455b; file 0x158cb
    mov ax, 0FFFEh
; Original operand: cmp     byte ptr ds:1E1h, 0
L_0C3E: ; original 0x2455e; IDA unnamed; logical 0x1455e; file 0x158ce
    cmp BYTE PTR ds:[01E1h], 00h
; Original operand: jnz     short locret_245B7
L_0C43: ; original 0x24563; IDA unnamed; logical 0x14563; file 0x158d3
    jnz SHORT L_0C97
; Original operand: mov     ds:1Eh, cx
L_0C45: ; original 0x24565; IDA unnamed; logical 0x14565; file 0x158d5
    mov WORD PTR ds:[01Eh], cx
; Original operand: mov     ds:20h, dx
L_0C49: ; original 0x24569; IDA unnamed; logical 0x14569; file 0x158d9
    mov WORD PTR ds:[020h], dx
; Original operand: mov     ds:1Ah, cx
L_0C4D: ; original 0x2456d; IDA unnamed; logical 0x1456d; file 0x158dd
    mov WORD PTR ds:[01Ah], cx
; Original operand: mov     ds:1Ch, dx
L_0C51: ; original 0x24571; IDA unnamed; logical 0x14571; file 0x158e1
    mov WORD PTR ds:[01Ch], dx
; Original operand: mov     cx, 10h
L_0C55: ; original 0x24575; IDA unnamed; logical 0x14575; file 0x158e5
    mov cx, 010h
; Original operand: sub     ax, ax
L_0C58: ; original 0x24578; IDA unnamed; logical 0x14578; file 0x158e8
    sub ax, ax
; Original operand: mov     di, 1F9h
L_0C5A: ; original 0x2457a; IDA unnamed; logical 0x1457a; file 0x158ea
    mov di, 01F9h
; Original operand: rep stosw
L_0C5D: ; original 0x2457d; IDA unnamed; logical 0x1457d; file 0x158ed
    rep stosw
; Original operand: mov     cx, 9
L_0C5F: ; original 0x2457f; IDA unnamed; logical 0x1457f; file 0x158ef
    mov cx, 09h
; Original operand: mov     al, 0FFh
L_0C62: ; original 0x24582; IDA unnamed; logical 0x14582; file 0x158f2
    mov al, 0FFh
; Original operand: mov     di, 148h
L_0C64: ; original 0x24584; IDA unnamed; logical 0x14584; file 0x158f4
    mov di, 0148h
; Original operand: rep stosb
L_0C67: ; original 0x24587; IDA unnamed; logical 0x14587; file 0x158f7
    rep stosb
; Original operand: call    sub_2443F
L_0C69: ; original 0x24589; IDA unnamed; logical 0x14589; file 0x158f9
    call L_0B1F
; Original operand: mov     ds:22h, ax
L_0C6C: ; original 0x2458c; IDA unnamed; logical 0x1458c; file 0x158fc
    mov WORD PTR ds:[022h], ax
; Original operand: mov     ds:24h, dx
L_0C6F: ; original 0x2458f; IDA unnamed; logical 0x1458f; file 0x158ff
    mov WORD PTR ds:[024h], dx
; Original operand: mov     word ptr ds:40h, 0
L_0C73: ; original 0x24593; IDA unnamed; logical 0x14593; file 0x15903
    mov WORD PTR ds:[040h], 00h
; Original operand: mov     ax, ds:34h
L_0C79: ; original 0x24599; IDA unnamed; logical 0x14599; file 0x15909
    mov ax, WORD PTR ds:[034h]
; Original operand: call    sub_2424D
L_0C7C: ; original 0x2459c; IDA unnamed; logical 0x1459c; file 0x1590c
    call L_092D
; Original operand: mov     word ptr ds:38h, 0
L_0C7F: ; original 0x2459f; IDA unnamed; logical 0x1459f; file 0x1590f
    mov WORD PTR ds:[038h], 00h
; Original operand: call    sub_2450E
L_0C85: ; original 0x245a5; IDA unnamed; logical 0x145a5; file 0x15915
    call L_0BEE
; Original operand: pushf
L_0C88: ; original 0x245a8; IDA unnamed; logical 0x145a8; file 0x15918
    pushf 
; Original operand: cli
L_0C89: ; original 0x245a9; IDA unnamed; logical 0x145a9; file 0x15919
    cli 
; Original operand: mov     byte ptr ds:1E1h, 1
L_0C8A: ; original 0x245aa; IDA unnamed; logical 0x145aa; file 0x1591a
    mov BYTE PTR ds:[01E1h], 01h
; Original operand: mov     al, 0FFh
L_0C8F: ; original 0x245af; IDA unnamed; logical 0x145af; file 0x1591f
    mov al, 0FFh
; Original operand: call    sub_24333
L_0C91: ; original 0x245b1; IDA unnamed; logical 0x145b1; file 0x15921
    call L_0A13
; Original operand: popf
L_0C94: ; original 0x245b4; IDA unnamed; logical 0x145b4; file 0x15924
    popf 
; Original operand: sub     ax, ax
L_0C95: ; original 0x245b5; IDA unnamed; logical 0x145b5; file 0x15925
    sub ax, ax
; Original operand: retn
L_0C97: ; original 0x245b7; IDA locret_245B7; logical 0x145b7; file 0x15927
    ret 
; Original operand: mov     ax, 0FFFDh
L_0C98: ; original 0x245b8; IDA sub_245B8; logical 0x145b8; file 0x15928
    mov ax, 0FFFDh
; Original operand: cmp     byte ptr ds:1E1h, 1
L_0C9B: ; original 0x245bb; IDA unnamed; logical 0x145bb; file 0x1592b
    cmp BYTE PTR ds:[01E1h], 01h
; Original operand: jnz     short locret_245CC
L_0CA0: ; original 0x245c0; IDA unnamed; logical 0x145c0; file 0x15930
    jnz SHORT L_0CAC
; Original operand: mov     byte ptr ds:1E1h, 2
L_0CA2: ; original 0x245c2; IDA unnamed; logical 0x145c2; file 0x15932
    mov BYTE PTR ds:[01E1h], 02h
; Original operand: call    sub_24284
L_0CA7: ; original 0x245c7; IDA unnamed; logical 0x145c7; file 0x15937
    call L_0964
; Original operand: sub     ax, ax
L_0CAA: ; original 0x245ca; IDA unnamed; logical 0x145ca; file 0x1593a
    sub ax, ax
; Original operand: retn
L_0CAC: ; original 0x245cc; IDA locret_245CC; logical 0x145cc; file 0x1593c
    ret 
; Original operand: mov     ax, 0FFFCh
L_0CAD: ; original 0x245cd; IDA sub_245CD; logical 0x145cd; file 0x1593d
    mov ax, 0FFFCh
; Original operand: cmp     byte ptr ds:1E1h, 2
L_0CB0: ; original 0x245d0; IDA unnamed; logical 0x145d0; file 0x15940
    cmp BYTE PTR ds:[01E1h], 02h
; Original operand: jnz     short locret_245DE
L_0CB5: ; original 0x245d5; IDA unnamed; logical 0x145d5; file 0x15945
    jnz SHORT L_0CBE
; Original operand: mov     byte ptr ds:1E1h, 1
L_0CB7: ; original 0x245d7; IDA unnamed; logical 0x145d7; file 0x15947
    mov BYTE PTR ds:[01E1h], 01h
; Original operand: sub     ax, ax
L_0CBC: ; original 0x245dc; IDA unnamed; logical 0x145dc; file 0x1594c
    sub ax, ax
; Original operand: retn
L_0CBE: ; original 0x245de; IDA locret_245DE; logical 0x145de; file 0x1594e
    ret 
; Original operand: push    es
L_0CBF: ; original 0x245df; IDA sub_245DF; logical 0x145df; file 0x1594f
    push es
; Original operand: inc     word ptr ds:40h
L_0CC0: ; original 0x245e0; IDA unnamed; logical 0x145e0; file 0x15950
    inc WORD PTR ds:[040h]
; Original operand: mov     ax, ds:1Ah
L_0CC4: ; original 0x245e4; IDA unnamed; logical 0x145e4; file 0x15954
    mov ax, WORD PTR ds:[01Ah]
; Original operand: shr     ax, 1
L_0CC7: ; original 0x245e7; IDA unnamed; logical 0x145e7; file 0x15957
    shr ax, 01h
; Original operand: shr     ax, 1
L_0CC9: ; original 0x245e9; IDA unnamed; logical 0x145e9; file 0x15959
    shr ax, 01h
; Original operand: shr     ax, 1
L_0CCB: ; original 0x245eb; IDA unnamed; logical 0x145eb; file 0x1595b
    shr ax, 01h
; Original operand: shr     ax, 1
L_0CCD: ; original 0x245ed; IDA unnamed; logical 0x145ed; file 0x1595d
    shr ax, 01h
; Original operand: add     ds:1Ch, ax
L_0CCF: ; original 0x245ef; IDA unnamed; logical 0x145ef; file 0x1595f
    add WORD PTR ds:[01Ch], ax
; Original operand: and     word ptr ds:1Ah, 0Fh
L_0CD3: ; original 0x245f3; IDA unnamed; logical 0x145f3; file 0x15963
    and WORD PTR ds:[01Ah], 0Fh
; Original operand: les     si, ds:1Ah
L_0CD8: ; original 0x245f8; IDA loc_245F8; logical 0x145f8; file 0x15968
    les si, DWORD PTR ds:[01Ah]
; Original operand: mov     al, es:[si]
L_0CDC: ; original 0x245fc; IDA unnamed; logical 0x145fc; file 0x1596c
    mov al, BYTE PTR es:[si]
; Original operand: or      al, al
L_0CDF: ; original 0x245ff; IDA unnamed; logical 0x145ff; file 0x1596f
    or al, al
; Original operand: jns     short loc_2461A
L_0CE1: ; original 0x24601; IDA unnamed; logical 0x14601; file 0x15971
    jns SHORT L_0CFA
; Original operand: inc     si
L_0CE3: ; original 0x24603; IDA unnamed; logical 0x14603; file 0x15973
    inc si
; Original operand: mov     ah, al
L_0CE4: ; original 0x24604; IDA unnamed; logical 0x14604; file 0x15974
    mov ah, al
; Original operand: and     al, 0Fh
L_0CE6: ; original 0x24606; IDA unnamed; logical 0x14606; file 0x15976
    and al, 0Fh
; Original operand: mov     ds:3Ch, al
L_0CE8: ; original 0x24608; IDA unnamed; logical 0x14608; file 0x15978
    mov BYTE PTR ds:[03Ch], al
; Original operand: shr     ah, 1
L_0CEB: ; original 0x2460b; IDA unnamed; logical 0x1460b; file 0x1597b
    shr ah, 01h
; Original operand: shr     ah, 1
L_0CED: ; original 0x2460d; IDA unnamed; logical 0x1460d; file 0x1597d
    shr ah, 01h
; Original operand: shr     ah, 1
L_0CEF: ; original 0x2460f; IDA unnamed; logical 0x1460f; file 0x1597f
    shr ah, 01h
; Original operand: shr     ah, 1
L_0CF1: ; original 0x24611; IDA unnamed; logical 0x14611; file 0x15981
    shr ah, 01h
; Original operand: sub     ah, 8
L_0CF3: ; original 0x24613; IDA unnamed; logical 0x14613; file 0x15983
    sub ah, 08h
; Original operand: mov     ds:3Eh, ah
L_0CF6: ; original 0x24616; IDA unnamed; logical 0x14616; file 0x15986
    mov BYTE PTR ds:[03Eh], ah
; Original operand: mov     bx, ds:3Eh
L_0CFA: ; original 0x2461a; IDA loc_2461A; logical 0x1461a; file 0x1598a
    mov bx, WORD PTR ds:[03Eh]
; Original operand: shl     bx, 1
L_0CFE: ; original 0x2461e; IDA unnamed; logical 0x1461e; file 0x1598e
    shl bx, 01h
; Original operand: call    word ptr [bx+229h]
L_0D00: ; original 0x24620; IDA unnamed; logical 0x14620; file 0x15990
    call WORD PTR [bx+0229h]
; Original operand: mov     ds:1Ah, si
L_0D04: ; original 0x24624; IDA unnamed; logical 0x14624; file 0x15994
    mov WORD PTR ds:[01Ah], si
; Original operand: cmp     byte ptr ds:1E1h, 0
L_0D08: ; original 0x24628; IDA unnamed; logical 0x14628; file 0x15998
    cmp BYTE PTR ds:[01E1h], 00h
; Original operand: jz      short loc_24647
L_0D0D: ; original 0x2462d; IDA unnamed; logical 0x1462d; file 0x1599d
    jz SHORT L_0D27
; Original operand: call    sub_2443F
L_0D0F: ; original 0x2462f; IDA unnamed; logical 0x1462f; file 0x1599f
    call L_0B1F
; Original operand: mov     ds:22h, ax
L_0D12: ; original 0x24632; IDA unnamed; logical 0x14632; file 0x159a2
    mov WORD PTR ds:[022h], ax
; Original operand: mov     ds:24h, dx
L_0D15: ; original 0x24635; IDA unnamed; logical 0x14635; file 0x159a5
    mov WORD PTR ds:[024h], dx
; Original operand: or      ax, dx
L_0D19: ; original 0x24639; IDA unnamed; logical 0x14639; file 0x159a9
    or ax, dx
; Original operand: jz      short loc_245F8
L_0D1B: ; original 0x2463b; IDA unnamed; logical 0x1463b; file 0x159ab
    jz SHORT L_0CD8
; Original operand: sub     word ptr ds:22h, 1
L_0D1D: ; original 0x2463d; IDA unnamed; logical 0x1463d; file 0x159ad
    sub WORD PTR ds:[022h], 01h
; Original operand: sbb     word ptr ds:24h, 0
L_0D22: ; original 0x24642; IDA unnamed; logical 0x14642; file 0x159b2
    sbb WORD PTR ds:[024h], 00h
; Original operand: pop     es
L_0D27: ; original 0x24647; IDA loc_24647; logical 0x14647; file 0x159b7
    pop es
; Original operand: retn
L_0D28: ; original 0x24648; IDA unnamed; logical 0x14648; file 0x159b8
    ret 
; Original operand: mov     ax, es:[si]
L_0D29: ; original 0x24649; IDA unnamed; logical 0x14649; file 0x159b9
    mov ax, WORD PTR es:[si]
; Original operand: inc     si
L_0D2C: ; original 0x2464c; IDA unnamed; logical 0x1464c; file 0x159bc
    inc si
; Original operand: inc     si
L_0D2D: ; original 0x2464d; IDA unnamed; logical 0x1464d; file 0x159bd
    inc si
; Original operand: mov     ds:1E2h, al
L_0D2E: ; original 0x2464e; IDA unnamed; logical 0x1464e; file 0x159be
    mov BYTE PTR ds:[01E2h], al
; Original operand: mov     ds:1E3h, ah
L_0D31: ; original 0x24651; IDA unnamed; logical 0x14651; file 0x159c1
    mov BYTE PTR ds:[01E3h], ah
; Original operand: push    es
L_0D35: ; original 0x24655; IDA loc_24655; logical 0x14655; file 0x159c5
    push es
; Original operand: mov     ax, ds
L_0D36: ; original 0x24656; IDA unnamed; logical 0x14656; file 0x159c6
    mov ax, ds
; Original operand: mov     es, ax
L_0D38: ; original 0x24658; IDA unnamed; logical 0x14658; file 0x159c8
    mov es, ax
; Original operand: mov     cx, ds:3Ah
L_0D3A: ; original 0x2465a; IDA unnamed; logical 0x1465a; file 0x159ca
    mov cx, WORD PTR ds:[03Ah]
; Original operand: mov     ah, ds:3Ch
L_0D3E: ; original 0x2465e; IDA unnamed; logical 0x1465e; file 0x159ce
    mov ah, BYTE PTR ds:[03Ch]
; Original operand: cmp     cl, 6
L_0D42: ; original 0x24662; IDA unnamed; logical 0x14662; file 0x159d2
    cmp cl, 06h
; Original operand: ja      short loc_24684
L_0D45: ; original 0x24665; IDA unnamed; logical 0x14665; file 0x159d5
    ja SHORT L_0D64
; Original operand: cmp     ah, 0Bh
L_0D47: ; original 0x24667; IDA unnamed; logical 0x14667; file 0x159d7
    cmp ah, 0Bh
; Original operand: jb      short loc_24684
L_0D4A: ; original 0x2466a; IDA unnamed; logical 0x1466a; file 0x159da
    jb SHORT L_0D64
; Original operand: sub     bh, bh
L_0D4C: ; original 0x2466c; IDA unnamed; logical 0x1466c; file 0x159dc
    sub bh, bh
; Original operand: mov     bl, ah
L_0D4E: ; original 0x2466e; IDA unnamed; logical 0x1466e; file 0x159de
    mov bl, ah
; Original operand: mov     al, [bx+1CBh]
L_0D50: ; original 0x24670; IDA unnamed; logical 0x14670; file 0x159e0
    mov al, BYTE PTR [bx+01CBh]
; Original operand: not     al
L_0D54: ; original 0x24674; IDA unnamed; logical 0x14674; file 0x159e4
    not al
; Original operand: and     al, ds:1E6h
L_0D56: ; original 0x24676; IDA unnamed; logical 0x14676; file 0x159e6
    and al, BYTE PTR ds:[01E6h]
; Original operand: mov     ds:1E6h, al
L_0D5A: ; original 0x2467a; IDA unnamed; logical 0x1467a; file 0x159ea
    mov BYTE PTR ds:[01E6h], al
; Original operand: mov     ah, 0BDh
L_0D5D: ; original 0x2467d; IDA unnamed; logical 0x1467d; file 0x159ed
    mov ah, 0BDh
; Original operand: call    sub_2425E
L_0D5F: ; original 0x2467f; IDA unnamed; logical 0x1467f; file 0x159ef
    call L_093E
; Original operand: jmp     short loc_246BC
L_0D62: ; original 0x24682; IDA unnamed; logical 0x14682; file 0x159f2
    jmp SHORT L_0D9C
; Original operand: mov     al, ds:1E2h
L_0D64: ; original 0x24684; IDA loc_24684; logical 0x14684; file 0x159f4
    mov al, BYTE PTR ds:[01E2h]
; Original operand: mov     di, 15Eh
L_0D67: ; original 0x24687; IDA unnamed; logical 0x14687; file 0x159f7
    mov di, 015Eh
; Original operand: repne scasb
L_0D6A: ; original 0x2468a; IDA loc_2468A; logical 0x1468a; file 0x159fa
    repne scasb
; Original operand: jnz     short loc_246BC
L_0D6C: ; original 0x2468c; IDA unnamed; logical 0x1468c; file 0x159fc
    jnz SHORT L_0D9C
; Original operand: mov     bx, di
L_0D6E: ; original 0x2468e; IDA unnamed; logical 0x1468e; file 0x159fe
    mov bx, di
; Original operand: sub     bx, 15Fh
L_0D70: ; original 0x24690; IDA unnamed; logical 0x14690; file 0x15a00
    sub bx, 015Fh
; Original operand: cmp     ah, [bx+148h]
L_0D74: ; original 0x24694; IDA unnamed; logical 0x14694; file 0x15a04
    cmp ah, BYTE PTR [bx+0148h]
; Original operand: jz      short loc_2469E
L_0D78: ; original 0x24698; IDA unnamed; logical 0x14698; file 0x15a08
    jz SHORT L_0D7E
; Original operand: jcxz    short loc_246BC
L_0D7A: ; original 0x2469a; IDA unnamed; logical 0x1469a; file 0x15a0a
    jcxz SHORT L_0D9C
; Original operand: jmp     short loc_2468A
L_0D7C: ; original 0x2469c; IDA unnamed; logical 0x1469c; file 0x15a0c
    jmp SHORT L_0D6A
; Original operand: or      byte ptr [bx+148h], 80h
L_0D7E: ; original 0x2469e; IDA loc_2469E; logical 0x1469e; file 0x15a0e
    or BYTE PTR [bx+0148h], 080h
; Original operand: shl     bx, 1
L_0D83: ; original 0x246a3; IDA unnamed; logical 0x146a3; file 0x15a13
    shl bx, 01h
; Original operand: mov     ax, [bx+18Ah]
L_0D85: ; original 0x246a5; IDA unnamed; logical 0x146a5; file 0x15a15
    mov ax, WORD PTR [bx+018Ah]
; Original operand: shr     bx, 1
L_0D89: ; original 0x246a9; IDA unnamed; logical 0x146a9; file 0x15a19
    shr bx, 01h
; Original operand: mov     dl, ah
L_0D8B: ; original 0x246ab; IDA unnamed; logical 0x146ab; file 0x15a1b
    mov dl, ah
; Original operand: mov     ah, 0A0h
L_0D8D: ; original 0x246ad; IDA unnamed; logical 0x146ad; file 0x15a1d
    mov ah, 0A0h
; Original operand: add     ah, bl
L_0D8F: ; original 0x246af; IDA unnamed; logical 0x146af; file 0x15a1f
    add ah, bl
; Original operand: call    sub_2425E
L_0D91: ; original 0x246b1; IDA unnamed; logical 0x146b1; file 0x15a21
    call L_093E
; Original operand: mov     al, dl
L_0D94: ; original 0x246b4; IDA unnamed; logical 0x146b4; file 0x15a24
    mov al, dl
; Original operand: add     ah, 10h
L_0D96: ; original 0x246b6; IDA unnamed; logical 0x146b6; file 0x15a26
    add ah, 010h
; Original operand: call    sub_2425E
L_0D99: ; original 0x246b9; IDA unnamed; logical 0x146b9; file 0x15a29
    call L_093E
; Original operand: pop     es
L_0D9C: ; original 0x246bc; IDA loc_246BC; logical 0x146bc; file 0x15a2c
    pop es
; Original operand: retn
L_0D9D: ; original 0x246bd; IDA unnamed; logical 0x146bd; file 0x15a2d
    ret 
; Original operand: jmp     short loc_24655
L_0D9E: ; original 0x246be; IDA loc_246BE; logical 0x146be; file 0x15a2e
    jmp SHORT L_0D35
; Original operand: mov     ax, es:[si]
L_0DA0: ; original 0x246c0; IDA unnamed; logical 0x146c0; file 0x15a30
    mov ax, WORD PTR es:[si]
; Original operand: inc     si
L_0DA3: ; original 0x246c3; IDA unnamed; logical 0x146c3; file 0x15a33
    inc si
; Original operand: inc     si
L_0DA4: ; original 0x246c4; IDA unnamed; logical 0x146c4; file 0x15a34
    inc si
; Original operand: mov     ds:1E2h, al
L_0DA5: ; original 0x246c5; IDA unnamed; logical 0x146c5; file 0x15a35
    mov BYTE PTR ds:[01E2h], al
; Original operand: mov     ds:1E3h, ah
L_0DA8: ; original 0x246c8; IDA unnamed; logical 0x146c8; file 0x15a38
    mov BYTE PTR ds:[01E3h], ah
; Original operand: or      ah, ah
L_0DAC: ; original 0x246cc; IDA unnamed; logical 0x146cc; file 0x15a3c
    or ah, ah
; Original operand: jz      short loc_246BE
L_0DAE: ; original 0x246ce; IDA unnamed; logical 0x146ce; file 0x15a3e
    jz SHORT L_0D9E
; Original operand: mov     al, ds:3Ch
L_0DB0: ; original 0x246d0; IDA unnamed; logical 0x146d0; file 0x15a40
    mov al, BYTE PTR ds:[03Ch]
; Original operand: mov     bl, al
L_0DB3: ; original 0x246d3; IDA unnamed; logical 0x146d3; file 0x15a43
    mov bl, al
; Original operand: sub     bh, bh
L_0DB5: ; original 0x246d5; IDA unnamed; logical 0x146d5; file 0x15a45
    sub bh, bh
; Original operand: cmp     bh, [bx+219h]
L_0DB7: ; original 0x246d7; IDA unnamed; logical 0x146d7; file 0x15a47
    cmp bh, BYTE PTR [bx+0219h]
; Original operand: jz      short locret_24741
L_0DBB: ; original 0x246db; IDA unnamed; logical 0x146db; file 0x15a4b
    jz SHORT L_0E21
; Original operand: cmp     byte ptr ds:1E7h, 0
L_0DBD: ; original 0x246dd; IDA unnamed; logical 0x146dd; file 0x15a4d
    cmp BYTE PTR ds:[01E7h], 00h
; Original operand: jz      short loc_246ED
L_0DC2: ; original 0x246e2; IDA unnamed; logical 0x146e2; file 0x15a52
    jz SHORT L_0DCD
; Original operand: cmp     al, 0Bh
L_0DC4: ; original 0x246e4; IDA unnamed; logical 0x146e4; file 0x15a54
    cmp al, 0Bh
; Original operand: jb      short loc_246ED
L_0DC6: ; original 0x246e6; IDA unnamed; logical 0x146e6; file 0x15a56
    jb SHORT L_0DCD
; Original operand: call    sub_24742
L_0DC8: ; original 0x246e8; IDA unnamed; logical 0x146e8; file 0x15a58
    call L_0E22
; Original operand: jmp     short locret_24741
L_0DCB: ; original 0x246eb; IDA unnamed; logical 0x146eb; file 0x15a5b
    jmp SHORT L_0E21
; Original operand: call    sub_2448E
L_0DCD: ; original 0x246ed; IDA loc_246ED; logical 0x146ed; file 0x15a5d
    call L_0B6E
; Original operand: mov     bx, ax
L_0DD0: ; original 0x246f0; IDA unnamed; logical 0x146f0; file 0x15a60
    mov bx, ax
; Original operand: mov     al, ds:3Ch
L_0DD2: ; original 0x246f2; IDA unnamed; logical 0x146f2; file 0x15a62
    mov al, BYTE PTR ds:[03Ch]
; Original operand: xchg    al, [bx+148h]
L_0DD5: ; original 0x246f5; IDA unnamed; logical 0x146f5; file 0x15a65
    xchg BYTE PTR [bx+0148h], al
; Original operand: and     al, 7Fh
L_0DD9: ; original 0x246f9; IDA unnamed; logical 0x146f9; file 0x15a69
    and al, 07Fh
; Original operand: cmp     al, ds:3Ch
L_0DDB: ; original 0x246fb; IDA unnamed; logical 0x146fb; file 0x15a6b
    cmp al, BYTE PTR ds:[03Ch]
; Original operand: jz      short loc_2470C
L_0DDF: ; original 0x246ff; IDA unnamed; logical 0x146ff; file 0x15a6f
    jz SHORT L_0DEC
; Original operand: mov     di, ds:3Ch
L_0DE1: ; original 0x24701; IDA unnamed; logical 0x14701; file 0x15a71
    mov di, WORD PTR ds:[03Ch]
; Original operand: mov     al, [di+1E9h]
L_0DE5: ; original 0x24705; IDA unnamed; logical 0x14705; file 0x15a75
    mov al, BYTE PTR [di+01E9h]
; Original operand: call    sub_2433E
L_0DE9: ; original 0x24709; IDA unnamed; logical 0x14709; file 0x15a79
    call L_0A1E
; Original operand: mov     cl, ds:1E3h
L_0DEC: ; original 0x2470c; IDA loc_2470C; logical 0x1470c; file 0x15a7c
    mov cl, BYTE PTR ds:[01E3h]
; Original operand: or      cl, 80h
L_0DF0: ; original 0x24710; IDA unnamed; logical 0x14710; file 0x15a80
    or cl, 080h
; Original operand: mov     al, [bx+174h]
L_0DF3: ; original 0x24713; IDA unnamed; logical 0x14713; file 0x15a83
    mov al, BYTE PTR [bx+0174h]
; Original operand: mul     cl
L_0DF7: ; original 0x24717; IDA unnamed; logical 0x14717; file 0x15a87
    mul cl
; Original operand: mov     al, 3Fh ; '?'
L_0DF9: ; original 0x24719; IDA unnamed; logical 0x14719; file 0x15a89
    mov al, 03Fh
; Original operand: sub     al, ah
L_0DFB: ; original 0x2471b; IDA unnamed; logical 0x1471b; file 0x15a8b
    sub al, ah
; Original operand: or      al, [bx+169h]
L_0DFD: ; original 0x2471d; IDA unnamed; logical 0x1471d; file 0x15a8d
    or al, BYTE PTR [bx+0169h]
; Original operand: mov     ah, [bx+1BDh]
L_0E01: ; original 0x24721; IDA unnamed; logical 0x14721; file 0x15a91
    mov ah, BYTE PTR [bx+01BDh]
; Original operand: add     ah, 43h ; 'C'
L_0E05: ; original 0x24725; IDA unnamed; logical 0x14725; file 0x15a95
    add ah, 043h
; Original operand: call    sub_2425E
L_0E08: ; original 0x24728; IDA unnamed; logical 0x14728; file 0x15a98
    call L_093E
; Original operand: call    sub_24795
L_0E0B: ; original 0x2472b; IDA unnamed; logical 0x1472b; file 0x15a9b
    call L_0E75
; Original operand: mov     dl, ah
L_0E0E: ; original 0x2472e; IDA unnamed; logical 0x1472e; file 0x15a9e
    mov dl, ah
; Original operand: mov     ah, 0A0h
L_0E10: ; original 0x24730; IDA unnamed; logical 0x14730; file 0x15aa0
    mov ah, 0A0h
; Original operand: add     ah, bl
L_0E12: ; original 0x24732; IDA unnamed; logical 0x14732; file 0x15aa2
    add ah, bl
; Original operand: call    sub_2425E
L_0E14: ; original 0x24734; IDA unnamed; logical 0x14734; file 0x15aa4
    call L_093E
; Original operand: mov     al, dl
L_0E17: ; original 0x24737; IDA unnamed; logical 0x14737; file 0x15aa7
    mov al, dl
; Original operand: or      al, 20h
L_0E19: ; original 0x24739; IDA unnamed; logical 0x14739; file 0x15aa9
    or al, 020h
; Original operand: add     ah, 10h
L_0E1B: ; original 0x2473b; IDA unnamed; logical 0x1473b; file 0x15aab
    add ah, 010h
; Original operand: call    sub_2425E
L_0E1E: ; original 0x2473e; IDA unnamed; logical 0x1473e; file 0x15aae
    call L_093E
; Original operand: retn
L_0E21: ; original 0x24741; IDA locret_24741; logical 0x14741; file 0x15ab1
    ret 
; Original operand: sub     al, 5
L_0E22: ; original 0x24742; IDA sub_24742; logical 0x14742; file 0x15ab2
    sub al, 05h
; Original operand: cbw
L_0E24: ; original 0x24744; IDA unnamed; logical 0x14744; file 0x15ab4
    cbw 
; Original operand: mov     bx, ax
L_0E25: ; original 0x24745; IDA unnamed; logical 0x14745; file 0x15ab5
    mov bx, ax
; Original operand: mov     al, [bx+1D0h]
L_0E27: ; original 0x24747; IDA unnamed; logical 0x14747; file 0x15ab7
    mov al, BYTE PTR [bx+01D0h]
; Original operand: or      ds:1E6h, al
L_0E2B: ; original 0x2474b; IDA unnamed; logical 0x1474b; file 0x15abb
    or BYTE PTR ds:[01E6h], al
; Original operand: mov     cl, ds:1E3h
L_0E2F: ; original 0x2474f; IDA unnamed; logical 0x1474f; file 0x15abf
    mov cl, BYTE PTR ds:[01E3h]
; Original operand: or      cl, 80h
L_0E33: ; original 0x24753; IDA unnamed; logical 0x14753; file 0x15ac3
    or cl, 080h
; Original operand: mov     al, [bx+174h]
L_0E36: ; original 0x24756; IDA unnamed; logical 0x14756; file 0x15ac6
    mov al, BYTE PTR [bx+0174h]
; Original operand: mul     cl
L_0E3A: ; original 0x2475a; IDA unnamed; logical 0x1475a; file 0x15aca
    mul cl
; Original operand: mov     al, 3Fh ; '?'
L_0E3C: ; original 0x2475c; IDA unnamed; logical 0x1475c; file 0x15acc
    mov al, 03Fh
; Original operand: sub     al, ah
L_0E3E: ; original 0x2475e; IDA unnamed; logical 0x1475e; file 0x15ace
    sub al, ah
; Original operand: or      al, [bx+169h]
L_0E40: ; original 0x24760; IDA unnamed; logical 0x14760; file 0x15ad0
    or al, BYTE PTR [bx+0169h]
; Original operand: mov     ah, [bx+1CBh]
L_0E44: ; original 0x24764; IDA unnamed; logical 0x14764; file 0x15ad4
    mov ah, BYTE PTR [bx+01CBh]
; Original operand: cmp     bl, 6
L_0E48: ; original 0x24768; IDA unnamed; logical 0x14768; file 0x15ad8
    cmp bl, 06h
; Original operand: jnz     short loc_24770
L_0E4B: ; original 0x2476b; IDA unnamed; logical 0x1476b; file 0x15adb
    jnz SHORT L_0E50
; Original operand: add     ah, 3
L_0E4D: ; original 0x2476d; IDA unnamed; logical 0x1476d; file 0x15add
    add ah, 03h
; Original operand: add     ah, 40h ; '@'
L_0E50: ; original 0x24770; IDA loc_24770; logical 0x14770; file 0x15ae0
    add ah, 040h
; Original operand: call    sub_2425E
L_0E53: ; original 0x24773; IDA unnamed; logical 0x14773; file 0x15ae3
    call L_093E
; Original operand: call    sub_24795
L_0E56: ; original 0x24776; IDA unnamed; logical 0x14776; file 0x15ae6
    call L_0E75
; Original operand: mov     dl, ah
L_0E59: ; original 0x24779; IDA unnamed; logical 0x14779; file 0x15ae9
    mov dl, ah
; Original operand: mov     ah, 0A0h
L_0E5B: ; original 0x2477b; IDA unnamed; logical 0x1477b; file 0x15aeb
    mov ah, 0A0h
; Original operand: add     ah, [bx+1D5h]
L_0E5D: ; original 0x2477d; IDA unnamed; logical 0x1477d; file 0x15aed
    add ah, BYTE PTR [bx+01D5h]
; Original operand: call    sub_2425E
L_0E61: ; original 0x24781; IDA unnamed; logical 0x14781; file 0x15af1
    call L_093E
; Original operand: mov     al, dl
L_0E64: ; original 0x24784; IDA unnamed; logical 0x14784; file 0x15af4
    mov al, dl
; Original operand: add     ah, 10h
L_0E66: ; original 0x24786; IDA unnamed; logical 0x14786; file 0x15af6
    add ah, 010h
; Original operand: call    sub_2425E
L_0E69: ; original 0x24789; IDA unnamed; logical 0x14789; file 0x15af9
    call L_093E
; Original operand: mov     al, ds:1E6h
L_0E6C: ; original 0x2478c; IDA unnamed; logical 0x1478c; file 0x15afc
    mov al, BYTE PTR ds:[01E6h]
; Original operand: mov     ah, 0BDh
L_0E6F: ; original 0x2478f; IDA unnamed; logical 0x1478f; file 0x15aff
    mov ah, 0BDh
; Original operand: call    sub_2425E
L_0E71: ; original 0x24791; IDA unnamed; logical 0x14791; file 0x15b01
    call L_093E
; Original operand: retn
L_0E74: ; original 0x24794; IDA unnamed; logical 0x14794; file 0x15b04
    ret 
; Original operand: mov     al, ds:1E2h
L_0E75: ; original 0x24795; IDA sub_24795; logical 0x14795; file 0x15b05
    mov al, BYTE PTR ds:[01E2h]
; Original operand: mov     [bx+15Eh], al
L_0E78: ; original 0x24798; IDA unnamed; logical 0x14798; file 0x15b08
    mov BYTE PTR [bx+015Eh], al
; Original operand: cbw
L_0E7C: ; original 0x2479c; IDA unnamed; logical 0x1479c; file 0x15b0c
    cbw 
; Original operand: mov     di, ds:42h
L_0E7D: ; original 0x2479d; IDA unnamed; logical 0x1479d; file 0x15b0d
    mov di, WORD PTR ds:[042h]
; Original operand: add     di, ax
L_0E81: ; original 0x247a1; IDA unnamed; logical 0x147a1; file 0x15b11
    add di, ax
; Original operand: jns     short loc_247A9
L_0E83: ; original 0x247a3; IDA unnamed; logical 0x147a3; file 0x15b13
    jns SHORT L_0E89
; Original operand: sub     di, di
L_0E85: ; original 0x247a5; IDA unnamed; logical 0x147a5; file 0x15b15
    sub di, di
; Original operand: jmp     short loc_247B2
L_0E87: ; original 0x247a7; IDA unnamed; logical 0x147a7; file 0x15b17
    jmp SHORT L_0E92
; Original operand: cmp     di, 80h
L_0E89: ; original 0x247a9; IDA loc_247A9; logical 0x147a9; file 0x15b19
    cmp di, 080h
; Original operand: jb      short loc_247B2
L_0E8D: ; original 0x247ad; IDA unnamed; logical 0x147ad; file 0x15b1d
    jb SHORT L_0E92
; Original operand: mov     di, 7Fh
L_0E8F: ; original 0x247af; IDA unnamed; logical 0x147af; file 0x15b1f
    mov di, 07Fh
; Original operand: mov     al, [di+285h]
L_0E92: ; original 0x247b2; IDA loc_247B2; logical 0x147b2; file 0x15b22
    mov al, BYTE PTR [di+0285h]
; Original operand: mov     [bx+153h], al
L_0E96: ; original 0x247b6; IDA unnamed; logical 0x147b6; file 0x15b26
    mov BYTE PTR [bx+0153h], al
; Original operand: call    sub_247BE
L_0E9A: ; original 0x247ba; IDA unnamed; logical 0x147ba; file 0x15b2a
    call L_0E9E
; Original operand: retn
L_0E9D: ; original 0x247bd; IDA unnamed; logical 0x147bd; file 0x15b2d
    ret 
; Original operand: mov     dl, al
L_0E9E: ; original 0x247be; IDA sub_247BE; logical 0x147be; file 0x15b2e
    mov dl, al
; Original operand: and     dl, 70h
L_0EA0: ; original 0x247c0; IDA unnamed; logical 0x147c0; file 0x15b30
    and dl, 070h
; Original operand: shr     dl, 1
L_0EA3: ; original 0x247c3; IDA unnamed; logical 0x147c3; file 0x15b33
    shr dl, 01h
; Original operand: shr     dl, 1
L_0EA5: ; original 0x247c5; IDA unnamed; logical 0x147c5; file 0x15b35
    shr dl, 01h
; Original operand: and     al, 0Fh
L_0EA7: ; original 0x247c7; IDA unnamed; logical 0x147c7; file 0x15b37
    and al, 0Fh
; Original operand: cbw
L_0EA9: ; original 0x247c9; IDA unnamed; logical 0x147c9; file 0x15b39
    cbw 
; Original operand: xchg    al, ah
L_0EAA: ; original 0x247ca; IDA unnamed; logical 0x147ca; file 0x15b3a
    xchg ah, al
; Original operand: shr     ax, 1
L_0EAC: ; original 0x247cc; IDA unnamed; logical 0x147cc; file 0x15b3c
    shr ax, 01h
; Original operand: shr     ax, 1
L_0EAE: ; original 0x247ce; IDA unnamed; logical 0x147ce; file 0x15b3e
    shr ax, 01h
; Original operand: mov     di, ds:3Ch
L_0EB0: ; original 0x247d0; IDA unnamed; logical 0x147d0; file 0x15b40
    mov di, WORD PTR ds:[03Ch]
; Original operand: shl     di, 1
L_0EB4: ; original 0x247d4; IDA unnamed; logical 0x147d4; file 0x15b44
    shl di, 01h
; Original operand: add     ax, [di+1F9h]
L_0EB6: ; original 0x247d6; IDA unnamed; logical 0x147d6; file 0x15b46
    add ax, WORD PTR [di+01F9h]
; Original operand: jns     short loc_247EA
L_0EBA: ; original 0x247da; IDA unnamed; logical 0x147da; file 0x15b4a
    jns SHORT L_0ECA
; Original operand: add     ax, 300h
L_0EBC: ; original 0x247dc; IDA unnamed; logical 0x147dc; file 0x15b4c
    add ax, IMM_0300
; Original operand: sub     dl, 4
L_0EBF: ; original 0x247df; IDA unnamed; logical 0x147df; file 0x15b4f
    sub dl, 04h
; Original operand: jns     short loc_247FF
L_0EC2: ; original 0x247e2; IDA unnamed; logical 0x147e2; file 0x15b52
    jns SHORT L_0EDF
; Original operand: sub     dl, dl
L_0EC4: ; original 0x247e4; IDA unnamed; logical 0x147e4; file 0x15b54
    sub dl, dl
; Original operand: sub     ax, ax
L_0EC6: ; original 0x247e6; IDA unnamed; logical 0x147e6; file 0x15b56
    sub ax, ax
; Original operand: jmp     short loc_247FF
L_0EC8: ; original 0x247e8; IDA unnamed; logical 0x147e8; file 0x15b58
    jmp SHORT L_0EDF
; Original operand: cmp     ax, 300h
L_0ECA: ; original 0x247ea; IDA loc_247EA; logical 0x147ea; file 0x15b5a
    cmp ax, IMM_0300
; Original operand: jb      short loc_247FF
L_0ECD: ; original 0x247ed; IDA unnamed; logical 0x147ed; file 0x15b5d
    jb SHORT L_0EDF
; Original operand: sub     ax, 300h
L_0ECF: ; original 0x247ef; IDA unnamed; logical 0x147ef; file 0x15b5f
    sub ax, IMM_0300
; Original operand: add     dl, 4
L_0ED2: ; original 0x247f2; IDA unnamed; logical 0x147f2; file 0x15b62
    add dl, 04h
; Original operand: cmp     dl, 1Ch
L_0ED5: ; original 0x247f5; IDA unnamed; logical 0x147f5; file 0x15b65
    cmp dl, 01Ch
; Original operand: jbe     short loc_247FF
L_0ED8: ; original 0x247f8; IDA unnamed; logical 0x147f8; file 0x15b68
    jbe SHORT L_0EDF
; Original operand: mov     ax, 2FFh
L_0EDA: ; original 0x247fa; IDA unnamed; logical 0x147fa; file 0x15b6a
    mov ax, 02FFh
; Original operand: mov     dl, 1Ch
L_0EDD: ; original 0x247fd; IDA unnamed; logical 0x147fd; file 0x15b6d
    mov dl, 01Ch
; Original operand: shl     ax, 1
L_0EDF: ; original 0x247ff; IDA loc_247FF; logical 0x147ff; file 0x15b6f
    shl ax, 01h
; Original operand: mov     di, ax
L_0EE1: ; original 0x24801; IDA unnamed; logical 0x14801; file 0x15b71
    mov di, ax
; Original operand: mov     ax, [di+305h]
L_0EE3: ; original 0x24803; IDA unnamed; logical 0x14803; file 0x15b73
    mov ax, WORD PTR [di+0305h]
; Original operand: or      ah, dl
L_0EE7: ; original 0x24807; IDA unnamed; logical 0x14807; file 0x15b77
    or ah, dl
; Original operand: shl     bx, 1
L_0EE9: ; original 0x24809; IDA unnamed; logical 0x14809; file 0x15b79
    shl bx, 01h
; Original operand: mov     [bx+18Ah], ax
L_0EEB: ; original 0x2480b; IDA unnamed; logical 0x1480b; file 0x15b7b
    mov WORD PTR [bx+018Ah], ax
; Original operand: mov     cx, ds:40h
L_0EEF: ; original 0x2480f; IDA unnamed; logical 0x1480f; file 0x15b7f
    mov cx, WORD PTR ds:[040h]
; Original operand: mov     [bx+1A0h], cx
L_0EF3: ; original 0x24813; IDA unnamed; logical 0x14813; file 0x15b83
    mov WORD PTR [bx+01A0h], cx
; Original operand: shr     bx, 1
L_0EF7: ; original 0x24817; IDA unnamed; logical 0x14817; file 0x15b87
    shr bx, 01h
; Original operand: retn
L_0EF9: ; original 0x24819; IDA unnamed; logical 0x14819; file 0x15b89
    ret 
; Original operand: mov     ax, es:[si]
L_0EFA: ; original 0x2481a; IDA unnamed; logical 0x1481a; file 0x15b8a
    mov ax, WORD PTR es:[si]
; Original operand: inc     si
L_0EFD: ; original 0x2481d; IDA unnamed; logical 0x1481d; file 0x15b8d
    inc si
; Original operand: inc     si
L_0EFE: ; original 0x2481e; IDA unnamed; logical 0x1481e; file 0x15b8e
    inc si
; Original operand: sub     al, 66h ; 'f'
L_0EFF: ; original 0x2481f; IDA unnamed; logical 0x1481f; file 0x15b8f
    sub al, 066h
; Original operand: cmp     al, 4
L_0F01: ; original 0x24821; IDA unnamed; logical 0x14821; file 0x15b91
    cmp al, 04h
; Original operand: jnb     short locret_2482F
L_0F03: ; original 0x24823; IDA unnamed; logical 0x14823; file 0x15b93
    jnb SHORT L_0F0F
; Original operand: mov     bl, al
L_0F05: ; original 0x24825; IDA unnamed; logical 0x14825; file 0x15b95
    mov bl, al
; Original operand: sub     bh, bh
L_0F07: ; original 0x24827; IDA unnamed; logical 0x14827; file 0x15b97
    sub bh, bh
; Original operand: shl     bx, 1
L_0F09: ; original 0x24829; IDA unnamed; logical 0x14829; file 0x15b99
    shl bx, 01h
; Original operand: call    word ptr [bx+239h]
L_0F0B: ; original 0x2482b; IDA unnamed; logical 0x1482b; file 0x15b9b
    call WORD PTR [bx+0239h]
; Original operand: retn
L_0F0F: ; original 0x2482f; IDA locret_2482F; logical 0x1482f; file 0x15b9f
    ret 
; Original operand: mov     al, ah
L_0F10: ; original 0x24830; IDA unnamed; logical 0x14830; file 0x15ba0
    mov al, ah
; Original operand: call    sub_24333
L_0F12: ; original 0x24832; IDA unnamed; logical 0x14832; file 0x15ba2
    call L_0A13
; Original operand: retn
L_0F15: ; original 0x24835; IDA unnamed; logical 0x14835; file 0x15ba5
    ret 
; Original operand: push    ax
L_0F16: ; original 0x24836; IDA unnamed; logical 0x14836; file 0x15ba6
    push ax
; Original operand: call    sub_242CD
L_0F17: ; original 0x24837; IDA unnamed; logical 0x14837; file 0x15ba7
    call L_09AD
; Original operand: pop     ax
L_0F1A: ; original 0x2483a; IDA unnamed; logical 0x1483a; file 0x15baa
    pop ax
; Original operand: mov     cx, 9C0h
L_0F1B: ; original 0x2483b; IDA unnamed; logical 0x1483b; file 0x15bab
    mov cx, 09C0h
; Original operand: mov     ds:1E7h, ah
L_0F1E: ; original 0x2483e; IDA unnamed; logical 0x1483e; file 0x15bae
    mov BYTE PTR ds:[01E7h], ah
; Original operand: or      ah, ah
L_0F22: ; original 0x24842; IDA unnamed; logical 0x14842; file 0x15bb2
    or ah, ah
; Original operand: jz      short loc_24849
L_0F24: ; original 0x24844; IDA unnamed; logical 0x14844; file 0x15bb4
    jz SHORT L_0F29
; Original operand: mov     cx, 6E0h
L_0F26: ; original 0x24846; IDA unnamed; logical 0x14846; file 0x15bb6
    mov cx, 06E0h
; Original operand: mov     ds:1E6h, cl
L_0F29: ; original 0x24849; IDA loc_24849; logical 0x14849; file 0x15bb9
    mov BYTE PTR ds:[01E6h], cl
; Original operand: mov     ds:3Ah, ch
L_0F2D: ; original 0x2484d; IDA unnamed; logical 0x1484d; file 0x15bbd
    mov BYTE PTR ds:[03Ah], ch
; Original operand: mov     al, ds:1E6h
L_0F31: ; original 0x24851; IDA unnamed; logical 0x14851; file 0x15bc1
    mov al, BYTE PTR ds:[01E6h]
; Original operand: mov     ah, 0BDh
L_0F34: ; original 0x24854; IDA unnamed; logical 0x14854; file 0x15bc4
    mov ah, 0BDh
; Original operand: call    sub_2425E
L_0F36: ; original 0x24856; IDA unnamed; logical 0x14856; file 0x15bc6
    call L_093E
; Original operand: retn
L_0F39: ; original 0x24859; IDA unnamed; logical 0x14859; file 0x15bc9
    ret 
; Original operand: neg     ah
L_0F3A: ; original 0x2485a; IDA unnamed; logical 0x1485a; file 0x15bca
    neg ah
; Original operand: push    es
L_0F3C: ; original 0x2485c; IDA unnamed; logical 0x1485c; file 0x15bcc
    push es
; Original operand: mov     bx, cs
L_0F3D: ; original 0x2485d; IDA unnamed; logical 0x1485d; file 0x15bcd
    mov bx, cs
; Original operand: mov     es, bx
L_0F3F: ; original 0x2485f; IDA unnamed; logical 0x1485f; file 0x15bcf
    mov es, bx
; Original operand: mov     al, ah
L_0F41: ; original 0x24861; IDA unnamed; logical 0x14861; file 0x15bd1
    mov al, ah
; Original operand: cbw
L_0F43: ; original 0x24863; IDA unnamed; logical 0x14863; file 0x15bd3
    cbw 
; Original operand: sar     ax, 1
L_0F44: ; original 0x24864; IDA unnamed; logical 0x14864; file 0x15bd4
    sar ax, 01h
; Original operand: sar     ax, 1
L_0F46: ; original 0x24866; IDA unnamed; logical 0x14866; file 0x15bd6
    sar ax, 01h
; Original operand: mov     bx, ds:3Ch
L_0F48: ; original 0x24868; IDA unnamed; logical 0x14868; file 0x15bd8
    mov bx, WORD PTR ds:[03Ch]
; Original operand: shl     bx, 1
L_0F4C: ; original 0x2486c; IDA unnamed; logical 0x1486c; file 0x15bdc
    shl bx, 01h
; Original operand: mov     [bx+1F9h], ax
L_0F4E: ; original 0x2486e; IDA unnamed; logical 0x1486e; file 0x15bde
    mov WORD PTR [bx+01F9h], ax
; Original operand: shr     bx, 1
L_0F52: ; original 0x24872; IDA unnamed; logical 0x14872; file 0x15be2
    shr bx, 01h
; Original operand: mov     cx, ds:3Ah
L_0F54: ; original 0x24874; IDA unnamed; logical 0x14874; file 0x15be4
    mov cx, WORD PTR ds:[03Ah]
; Original operand: mov     di, 148h
L_0F58: ; original 0x24878; IDA unnamed; logical 0x14878; file 0x15be8
    mov di, 0148h
; Original operand: mov     al, bl
L_0F5B: ; original 0x2487b; IDA unnamed; logical 0x1487b; file 0x15beb
    mov al, bl
; Original operand: repne scasb
L_0F5D: ; original 0x2487d; IDA loc_2487D; logical 0x1487d; file 0x15bed
    repne scasb
; Original operand: jnz     short loc_248AB
L_0F5F: ; original 0x2487f; IDA unnamed; logical 0x1487f; file 0x15bef
    jnz SHORT L_0F8B
; Original operand: push    ax
L_0F61: ; original 0x24881; IDA unnamed; logical 0x14881; file 0x15bf1
    push ax
; Original operand: push    cx
L_0F62: ; original 0x24882; IDA unnamed; logical 0x14882; file 0x15bf2
    push cx
; Original operand: push    di
L_0F63: ; original 0x24883; IDA unnamed; logical 0x14883; file 0x15bf3
    push di
; Original operand: sub     di, 149h
L_0F64: ; original 0x24884; IDA unnamed; logical 0x14884; file 0x15bf4
    sub di, 0149h
; Original operand: mov     bx, di
L_0F68: ; original 0x24888; IDA unnamed; logical 0x14888; file 0x15bf8
    mov bx, di
; Original operand: mov     al, [bx+153h]
L_0F6A: ; original 0x2488a; IDA unnamed; logical 0x1488a; file 0x15bfa
    mov al, BYTE PTR [bx+0153h]
; Original operand: call    sub_247BE
L_0F6E: ; original 0x2488e; IDA unnamed; logical 0x1488e; file 0x15bfe
    call L_0E9E
; Original operand: mov     dl, ah
L_0F71: ; original 0x24891; IDA unnamed; logical 0x14891; file 0x15c01
    mov dl, ah
; Original operand: mov     ah, 0A0h
L_0F73: ; original 0x24893; IDA unnamed; logical 0x14893; file 0x15c03
    mov ah, 0A0h
; Original operand: add     ah, bl
L_0F75: ; original 0x24895; IDA unnamed; logical 0x14895; file 0x15c05
    add ah, bl
; Original operand: call    sub_2425E
L_0F77: ; original 0x24897; IDA unnamed; logical 0x14897; file 0x15c07
    call L_093E
; Original operand: mov     al, dl
L_0F7A: ; original 0x2489a; IDA unnamed; logical 0x1489a; file 0x15c0a
    mov al, dl
; Original operand: or      al, 20h
L_0F7C: ; original 0x2489c; IDA unnamed; logical 0x1489c; file 0x15c0c
    or al, 020h
; Original operand: add     ah, 10h
L_0F7E: ; original 0x2489e; IDA unnamed; logical 0x1489e; file 0x15c0e
    add ah, 010h
; Original operand: call    sub_2425E
L_0F81: ; original 0x248a1; IDA unnamed; logical 0x148a1; file 0x15c11
    call L_093E
; Original operand: pop     di
L_0F84: ; original 0x248a4; IDA unnamed; logical 0x148a4; file 0x15c14
    pop di
; Original operand: pop     cx
L_0F85: ; original 0x248a5; IDA unnamed; logical 0x148a5; file 0x15c15
    pop cx
; Original operand: pop     ax
L_0F86: ; original 0x248a6; IDA unnamed; logical 0x148a6; file 0x15c16
    pop ax
; Original operand: or      cx, cx
L_0F87: ; original 0x248a7; IDA unnamed; logical 0x148a7; file 0x15c17
    or cx, cx
; Original operand: jnz     short loc_2487D
L_0F89: ; original 0x248a9; IDA unnamed; logical 0x148a9; file 0x15c19
    jnz SHORT L_0F5D
; Original operand: pop     es
L_0F8B: ; original 0x248ab; IDA loc_248AB; logical 0x148ab; file 0x15c1b
    pop es
; Original operand: retn
L_0F8C: ; original 0x248ac; IDA unnamed; logical 0x148ac; file 0x15c1c
    ret 
; Original operand: inc     si
L_0F8D: ; original 0x248ad; IDA unnamed; logical 0x148ad; file 0x15c1d
    inc si
; Original operand: inc     si
L_0F8E: ; original 0x248ae; IDA unnamed; logical 0x148ae; file 0x15c1e
    inc si
; Original operand: retn
L_0F8F: ; original 0x248af; IDA unnamed; logical 0x148af; file 0x15c1f
    ret 
; Original operand: mov     al, es:[si]
L_0F90: ; original 0x248b0; IDA unnamed; logical 0x148b0; file 0x15c20
    mov al, BYTE PTR es:[si]
; Original operand: inc     si
L_0F93: ; original 0x248b3; IDA unnamed; logical 0x148b3; file 0x15c23
    inc si
; Original operand: cmp     al, ds:1E0h
L_0F94: ; original 0x248b4; IDA loc_248B4; logical 0x148b4; file 0x15c24
    cmp al, BYTE PTR ds:[01E0h]
; Original operand: jb      short loc_248C0
L_0F98: ; original 0x248b8; IDA unnamed; logical 0x148b8; file 0x15c28
    jb SHORT L_0FA0
; Original operand: sub     al, ds:1E0h
L_0F9A: ; original 0x248ba; IDA unnamed; logical 0x148ba; file 0x15c2a
    sub al, BYTE PTR ds:[01E0h]
; Original operand: jmp     short loc_248B4
L_0F9E: ; original 0x248be; IDA unnamed; logical 0x148be; file 0x15c2e
    jmp SHORT L_0F94
; Original operand: mov     bx, ds:3Ch
L_0FA0: ; original 0x248c0; IDA loc_248C0; logical 0x148c0; file 0x15c30
    mov bx, WORD PTR ds:[03Ch]
; Original operand: add     bx, 1E9h
L_0FA4: ; original 0x248c4; IDA unnamed; logical 0x148c4; file 0x15c34
    add bx, 01E9h
; Original operand: mov     [bx], al
L_0FA8: ; original 0x248c8; IDA unnamed; logical 0x148c8; file 0x15c38
    mov BYTE PTR [bx], al
; Original operand: push    es
L_0FAA: ; original 0x248ca; IDA unnamed; logical 0x148ca; file 0x15c3a
    push es
; Original operand: mov     ax, ds
L_0FAB: ; original 0x248cb; IDA unnamed; logical 0x148cb; file 0x15c3b
    mov ax, ds
; Original operand: mov     es, ax
L_0FAD: ; original 0x248cd; IDA unnamed; logical 0x148cd; file 0x15c3d
    mov es, ax
; Original operand: cmp     byte ptr ds:1E7h, 0
L_0FAF: ; original 0x248cf; IDA unnamed; logical 0x148cf; file 0x15c3f
    cmp BYTE PTR ds:[01E7h], 00h
; Original operand: jz      short loc_248ED
L_0FB4: ; original 0x248d4; IDA unnamed; logical 0x148d4; file 0x15c44
    jz SHORT L_0FCD
; Original operand: cmp     word ptr ds:3Ch, 0Bh
L_0FB6: ; original 0x248d6; IDA unnamed; logical 0x148d6; file 0x15c46
    cmp WORD PTR ds:[03Ch], 0Bh
; Original operand: jb      short loc_248ED
L_0FBB: ; original 0x248db; IDA unnamed; logical 0x148db; file 0x15c4b
    jb SHORT L_0FCD
; Original operand: mov     bx, ds:3Ch
L_0FBD: ; original 0x248dd; IDA unnamed; logical 0x148dd; file 0x15c4d
    mov bx, WORD PTR ds:[03Ch]
; Original operand: mov     al, [bx+1E9h]
L_0FC1: ; original 0x248e1; IDA unnamed; logical 0x148e1; file 0x15c51
    mov al, BYTE PTR [bx+01E9h]
; Original operand: sub     bl, 5
L_0FC5: ; original 0x248e5; IDA unnamed; logical 0x148e5; file 0x15c55
    sub bl, 05h
; Original operand: call    sub_2433E
L_0FC8: ; original 0x248e8; IDA unnamed; logical 0x148e8; file 0x15c58
    call L_0A1E
; Original operand: jmp     short loc_24928
L_0FCB: ; original 0x248eb; IDA unnamed; logical 0x148eb; file 0x15c5b
    jmp SHORT L_1008
; Original operand: mov     cx, ds:3Ah
L_0FCD: ; original 0x248ed; IDA loc_248ED; logical 0x148ed; file 0x15c5d
    mov cx, WORD PTR ds:[03Ah]
; Original operand: mov     al, ds:3Ch
L_0FD1: ; original 0x248f1; IDA unnamed; logical 0x148f1; file 0x15c61
    mov al, BYTE PTR ds:[03Ch]
; Original operand: or      al, 80h
L_0FD4: ; original 0x248f4; IDA unnamed; logical 0x148f4; file 0x15c64
    or al, 080h
; Original operand: mov     di, 148h
L_0FD6: ; original 0x248f6; IDA unnamed; logical 0x148f6; file 0x15c66
    mov di, 0148h
; Original operand: repne scasb
L_0FD9: ; original 0x248f9; IDA loc_248F9; logical 0x148f9; file 0x15c69
    repne scasb
; Original operand: jnz     short loc_24905
L_0FDB: ; original 0x248fb; IDA unnamed; logical 0x148fb; file 0x15c6b
    jnz SHORT L_0FE5
; Original operand: mov     byte ptr [di-1], 0FFh
L_0FDD: ; original 0x248fd; IDA unnamed; logical 0x148fd; file 0x15c6d
    mov BYTE PTR [di-01h], 0FFh
; Original operand: or      cx, cx
L_0FE1: ; original 0x24901; IDA unnamed; logical 0x14901; file 0x15c71
    or cx, cx
; Original operand: jnz     short loc_248F9
L_0FE3: ; original 0x24903; IDA unnamed; logical 0x14903; file 0x15c73
    jnz SHORT L_0FD9
; Original operand: mov     cx, ds:3Ah
L_0FE5: ; original 0x24905; IDA loc_24905; logical 0x14905; file 0x15c75
    mov cx, WORD PTR ds:[03Ah]
; Original operand: mov     di, 148h
L_0FE9: ; original 0x24909; IDA unnamed; logical 0x14909; file 0x15c79
    mov di, 0148h
; Original operand: mov     al, ds:3Ch
L_0FEC: ; original 0x2490c; IDA loc_2490C; logical 0x1490c; file 0x15c7c
    mov al, BYTE PTR ds:[03Ch]
; Original operand: repne scasb
L_0FEF: ; original 0x2490f; IDA unnamed; logical 0x1490f; file 0x15c7f
    repne scasb
; Original operand: jnz     short loc_24928
L_0FF1: ; original 0x24911; IDA unnamed; logical 0x14911; file 0x15c81
    jnz SHORT L_1008
; Original operand: mov     bx, ds:3Ch
L_0FF3: ; original 0x24913; IDA unnamed; logical 0x14913; file 0x15c83
    mov bx, WORD PTR ds:[03Ch]
; Original operand: mov     al, [bx+1E9h]
L_0FF7: ; original 0x24917; IDA unnamed; logical 0x14917; file 0x15c87
    mov al, BYTE PTR [bx+01E9h]
; Original operand: mov     bx, di
L_0FFB: ; original 0x2491b; IDA unnamed; logical 0x1491b; file 0x15c8b
    mov bx, di
; Original operand: sub     bx, 149h
L_0FFD: ; original 0x2491d; IDA unnamed; logical 0x1491d; file 0x15c8d
    sub bx, 0149h
; Original operand: call    sub_2433E
L_1001: ; original 0x24921; IDA unnamed; logical 0x14921; file 0x15c91
    call L_0A1E
; Original operand: or      cx, cx
L_1004: ; original 0x24924; IDA unnamed; logical 0x14924; file 0x15c94
    or cx, cx
; Original operand: jnz     short loc_2490C
L_1006: ; original 0x24926; IDA unnamed; logical 0x14926; file 0x15c96
    jnz SHORT L_0FEC
; Original operand: pop     es
L_1008: ; original 0x24928; IDA loc_24928; logical 0x14928; file 0x15c98
    pop es
; Original operand: retn
L_1009: ; original 0x24929; IDA unnamed; logical 0x14929; file 0x15c99
    ret 
; Original operand: mov     bl, ds:3Ch
L_100A: ; original 0x2492a; IDA unnamed; logical 0x1492a; file 0x15c9a
    mov bl, BYTE PTR ds:[03Ch]
; Original operand: sub     bh, bh
L_100E: ; original 0x2492e; IDA unnamed; logical 0x1492e; file 0x15c9e
    sub bh, bh
; Original operand: shl     bx, 1
L_1010: ; original 0x24930; IDA unnamed; logical 0x14930; file 0x15ca0
    shl bx, 01h
; Original operand: call    word ptr [bx+241h]
L_1012: ; original 0x24932; IDA unnamed; logical 0x14932; file 0x15ca2
    call WORD PTR [bx+0241h]
; Original operand: retn
L_1016: ; original 0x24936; IDA unnamed; logical 0x14936; file 0x15ca6
    ret 
; Original operand: mov     ds:1Ah, si
L_1017: ; original 0x24937; IDA unnamed; logical 0x14937; file 0x15ca7
    mov WORD PTR ds:[01Ah], si
; Original operand: cmp     word ptr ds:2Ch, 0
L_101B: ; original 0x2493b; IDA unnamed; logical 0x1493b; file 0x15cab
    cmp WORD PTR ds:[02Ch], 00h
; Original operand: jz      short loc_24946
L_1020: ; original 0x24940; IDA unnamed; logical 0x14940; file 0x15cb0
    jz SHORT L_1026
; Original operand: call    dword ptr ds:2Ah
L_1022: ; original 0x24942; IDA unnamed; logical 0x14942; file 0x15cb2
    call DWORD PTR ds:[02Ah]
; Original operand: call    sub_2443F
L_1026: ; original 0x24946; IDA loc_24946; logical 0x14946; file 0x15cb6
    call L_0B1F
; Original operand: call    sub_2446F
L_1029: ; original 0x24949; IDA unnamed; logical 0x14949; file 0x15cb9
    call L_0B4F
; Original operand: retn
L_102C: ; original 0x2494c; IDA unnamed; logical 0x1494c; file 0x15cbc
    ret 
; Original operand: mov     ds:1Ah, si
L_102D: ; original 0x2494d; IDA unnamed; logical 0x1494d; file 0x15cbd
    mov WORD PTR ds:[01Ah], si
; Original operand: call    sub_2443F
L_1031: ; original 0x24951; IDA unnamed; logical 0x14951; file 0x15cc1
    call L_0B1F
; Original operand: call    sub_2446F
L_1034: ; original 0x24954; IDA unnamed; logical 0x14954; file 0x15cc4
    call L_0B4F
; Original operand: retn
L_1037: ; original 0x24957; IDA unnamed; logical 0x14957; file 0x15cc7
    ret 
; Original operand: call    sub_24AB8
L_1038: ; original 0x24958; IDA unnamed; logical 0x14958; file 0x15cc8
    call L_1198
; Original operand: retn
L_103B: ; original 0x2495b; IDA unnamed; logical 0x1495b; file 0x15ccb
    ret 
; Original operand: mov     al, es:[si]
L_103C: ; original 0x2495c; IDA unnamed; logical 0x1495c; file 0x15ccc
    mov al, BYTE PTR es:[si]
; Original operand: inc     si
L_103F: ; original 0x2495f; IDA unnamed; logical 0x1495f; file 0x15ccf
    inc si
; Original operand: cmp     al, 2Fh ; '/'
L_1040: ; original 0x24960; IDA unnamed; logical 0x14960; file 0x15cd0
    cmp al, 02Fh
; Original operand: jnz     short loc_24967
L_1042: ; original 0x24962; IDA unnamed; logical 0x14962; file 0x15cd2
    jnz SHORT L_1047
; Original operand: call    sub_24AB8
L_1044: ; original 0x24964; IDA unnamed; logical 0x14964; file 0x15cd4
    call L_1198
; Original operand: mov     ds:1Ah, si
L_1047: ; original 0x24967; IDA loc_24967; logical 0x14967; file 0x15cd7
    mov WORD PTR ds:[01Ah], si
; Original operand: call    sub_2443F
L_104B: ; original 0x2496b; IDA unnamed; logical 0x1496b; file 0x15cdb
    call L_0B1F
; Original operand: call    sub_2446F
L_104E: ; original 0x2496e; IDA unnamed; logical 0x1496e; file 0x15cde
    call L_0B4F
; Original operand: retn
L_1051: ; original 0x24971; IDA unnamed; logical 0x14971; file 0x15ce1
    ret 
; Original operand: push    ds
L_1052: ; original 0x24972; IDA unnamed; logical 0x14972; file 0x15ce2
    push ds
; Original operand: push    es
L_1053: ; original 0x24973; IDA unnamed; logical 0x14973; file 0x15ce3
    push es
; Original operand: push    ax
L_1054: ; original 0x24974; IDA unnamed; logical 0x14974; file 0x15ce4
    push ax
; Original operand: push    bx
L_1055: ; original 0x24975; IDA unnamed; logical 0x14975; file 0x15ce5
    push bx
; Original operand: push    cx
L_1056: ; original 0x24976; IDA unnamed; logical 0x14976; file 0x15ce6
    push cx
; Original operand: push    dx
L_1057: ; original 0x24977; IDA unnamed; logical 0x14977; file 0x15ce7
    push dx
; Original operand: push    di
L_1058: ; original 0x24978; IDA unnamed; logical 0x14978; file 0x15ce8
    push di
; Original operand: push    si
L_1059: ; original 0x24979; IDA unnamed; logical 0x14979; file 0x15ce9
    push si
; Original operand: push    bp
L_105A: ; original 0x2497a; IDA unnamed; logical 0x1497a; file 0x15cea
    push bp
; Original operand: mov     ax, cs
L_105B: ; original 0x2497b; IDA unnamed; logical 0x1497b; file 0x15ceb
    mov ax, cs
; Original operand: mov     ds, ax
L_105D: ; original 0x2497d; IDA unnamed; logical 0x1497d; file 0x15ced
    mov ds, ax
; Original operand: mov     es, ax
L_105F: ; original 0x2497f; IDA unnamed; logical 0x1497f; file 0x15cef
    mov es, ax
; Original operand: cmp     byte_23929+1D8h, 1
L_1061: ; original 0x24981; IDA unnamed; logical 0x14981; file 0x15cf1
    cmp BYTE PTR ds:[01E1h], 01h
; Original operand: jnz     short loc_249C8
L_1066: ; original 0x24986; IDA unnamed; logical 0x14986; file 0x15cf6
    jnz SHORT L_10A8
; Original operand: mov     ax, word ptr byte_23929+1ADh
L_1068: ; original 0x24988; IDA unnamed; logical 0x14988; file 0x15cf8
    mov ax, WORD PTR ds:[01B6h]
; Original operand: cmp     al, ah
L_106B: ; original 0x2498b; IDA unnamed; logical 0x1498b; file 0x15cfb
    cmp al, ah
; Original operand: jz      short loc_249A0
L_106D: ; original 0x2498d; IDA unnamed; logical 0x1498d; file 0x15cfd
    jz SHORT L_1080
; Original operand: dec     word ptr byte_23929+1AFh
L_106F: ; original 0x2498f; IDA unnamed; logical 0x1498f; file 0x15cff
    dec WORD PTR ds:[01B8h]
; Original operand: jnz     short loc_249A0
L_1073: ; original 0x24993; IDA unnamed; logical 0x14993; file 0x15d03
    jnz SHORT L_1080
; Original operand: mov     cx, word ptr byte_23929+1B1h
L_1075: ; original 0x24995; IDA unnamed; logical 0x14995; file 0x15d05
    mov cx, WORD PTR ds:[01BAh]
; Original operand: mov     word ptr byte_23929+1AFh, cx
L_1079: ; original 0x24999; IDA unnamed; logical 0x14999; file 0x15d09
    mov WORD PTR ds:[01B8h], cx
; Original operand: call    sub_24BC0
L_107D: ; original 0x2499d; IDA unnamed; logical 0x1499d; file 0x15d0d
    call L_12A0
; Original operand: sub     word ptr byte_23929+19h, 1
L_1080: ; original 0x249a0; IDA loc_249A0; logical 0x149a0; file 0x15d10
    sub WORD PTR ds:[022h], 01h
; Original operand: sbb     word ptr byte_23929+1Bh, 0
L_1085: ; original 0x249a5; IDA unnamed; logical 0x149a5; file 0x15d15
    sbb WORD PTR ds:[024h], 00h
; Original operand: jnb     short loc_249C8
L_108A: ; original 0x249aa; IDA unnamed; logical 0x149aa; file 0x15d1a
    jnb SHORT L_10A8
; Original operand: mov     word ptr byte_23929+3Dh, ss
L_108C: ; original 0x249ac; IDA unnamed; logical 0x149ac; file 0x15d1c
    mov WORD PTR ds:[046h], ss
; Original operand: mov     word ptr byte_23929+3Bh, sp
L_1090: ; original 0x249b0; IDA unnamed; logical 0x149b0; file 0x15d20
    mov WORD PTR ds:[044h], sp
; Original operand: cli
L_1094: ; original 0x249b4; IDA unnamed; logical 0x149b4; file 0x15d24
    cli 
; Original operand: mov     ax, cs
L_1095: ; original 0x249b5; IDA unnamed; logical 0x149b5; file 0x15d25
    mov ax, cs
; Original operand: mov     ss, ax
L_1097: ; original 0x249b7; IDA unnamed; logical 0x149b7; file 0x15d27
    mov ss, ax
; Original operand: mov     sp, 1337h
L_1099: ; original 0x249b9; IDA unnamed; logical 0x149b9; file 0x15d29
    mov sp, 01337h
; Original operand: cld
L_109C: ; original 0x249bc; IDA unnamed; logical 0x149bc; file 0x15d2c
    cld 
; Original operand: call    sub_245DF
L_109D: ; original 0x249bd; IDA unnamed; logical 0x149bd; file 0x15d2d
    call L_0CBF
; Original operand: mov     ss, word ptr byte_23929+3Dh
L_10A0: ; original 0x249c0; IDA unnamed; logical 0x149c0; file 0x15d30
    mov ss, WORD PTR ds:[046h]
; Original operand: mov     sp, word ptr byte_23929+3Bh
L_10A4: ; original 0x249c4; IDA unnamed; logical 0x149c4; file 0x15d34
    mov sp, WORD PTR ds:[044h]
; Original operand: mov     cx, word ptr byte_23929+29h
L_10A8: ; original 0x249c8; IDA loc_249C8; logical 0x149c8; file 0x15d38
    mov cx, WORD PTR ds:[032h]
; Original operand: mov     ax, word ptr byte_23929+2Dh
L_10AC: ; original 0x249cc; IDA unnamed; logical 0x149cc; file 0x15d3c
    mov ax, WORD PTR ds:[036h]
; Original operand: add     word ptr byte_23929+2Fh, ax
L_10AF: ; original 0x249cf; IDA unnamed; logical 0x149cf; file 0x15d3f
    add WORD PTR ds:[038h], ax
; Original operand: jb      short loc_249DB
L_10B3: ; original 0x249d3; IDA unnamed; logical 0x149d3; file 0x15d43
    jb SHORT L_10BB
; Original operand: cmp     word ptr byte_23929+2Fh, cx
L_10B5: ; original 0x249d5; IDA unnamed; logical 0x149d5; file 0x15d45
    cmp WORD PTR ds:[038h], cx
; Original operand: jb      short loc_249EC
L_10B9: ; original 0x249d9; IDA unnamed; logical 0x149d9; file 0x15d49
    jb SHORT L_10CC
; Original operand: sub     word ptr byte_23929+2Fh, cx
L_10BB: ; original 0x249db; IDA loc_249DB; logical 0x149db; file 0x15d4b
    sub WORD PTR ds:[038h], cx
; Original operand: pushf
L_10BF: ; original 0x249df; IDA unnamed; logical 0x149df; file 0x15d4f
    pushf 
; Original operand: call    dword ptr byte_23929+5
L_10C0: ; original 0x249e0; IDA unnamed; logical 0x149e0; file 0x15d50
    call DWORD PTR ds:[0Eh]
; Original operand: cmp     word ptr byte_23929+2Fh, cx
L_10C4: ; original 0x249e4; IDA unnamed; logical 0x149e4; file 0x15d54
    cmp WORD PTR ds:[038h], cx
; Original operand: ja      short loc_249DB
L_10C8: ; original 0x249e8; IDA unnamed; logical 0x149e8; file 0x15d58
    ja SHORT L_10BB
; Original operand: jmp     short loc_249F0
L_10CA: ; original 0x249ea; IDA unnamed; logical 0x149ea; file 0x15d5a
    jmp SHORT L_10D0
; Original operand: mov     al, 20h ; ' '
L_10CC: ; original 0x249ec; IDA loc_249EC; logical 0x149ec; file 0x15d5c
    mov al, 020h
; Original operand: out     20h, al; Interrupt controller, 8259A.
L_10CE: ; original 0x249ee; IDA unnamed; logical 0x149ee; file 0x15d5e
    out 020h, al
; Original operand: pop     bp
L_10D0: ; original 0x249f0; IDA loc_249F0; logical 0x149f0; file 0x15d60
    pop bp
; Original operand: pop     si
L_10D1: ; original 0x249f1; IDA unnamed; logical 0x149f1; file 0x15d61
    pop si
; Original operand: pop     di
L_10D2: ; original 0x249f2; IDA unnamed; logical 0x149f2; file 0x15d62
    pop di
; Original operand: pop     dx
L_10D3: ; original 0x249f3; IDA unnamed; logical 0x149f3; file 0x15d63
    pop dx
; Original operand: pop     cx
L_10D4: ; original 0x249f4; IDA unnamed; logical 0x149f4; file 0x15d64
    pop cx
; Original operand: pop     bx
L_10D5: ; original 0x249f5; IDA unnamed; logical 0x149f5; file 0x15d65
    pop bx
; Original operand: pop     ax
L_10D6: ; original 0x249f6; IDA unnamed; logical 0x149f6; file 0x15d66
    pop ax
; Original operand: pop     es
L_10D7: ; original 0x249f7; IDA unnamed; logical 0x149f7; file 0x15d67
    pop es
; Original operand: pop     ds
L_10D8: ; original 0x249f8; IDA unnamed; logical 0x149f8; file 0x15d68
    pop ds
; Original operand: iret
L_10D9: ; original 0x249f9; IDA unnamed; logical 0x149f9; file 0x15d69
    iret 
; Original operand: push    ds
L_10DA: ; original 0x249fa; IDA unnamed; logical 0x149fa; file 0x15d6a
    push ds
; Original operand: push    es
L_10DB: ; original 0x249fb; IDA unnamed; logical 0x149fb; file 0x15d6b
    push es
; Original operand: push    ax
L_10DC: ; original 0x249fc; IDA unnamed; logical 0x149fc; file 0x15d6c
    push ax
; Original operand: push    bx
L_10DD: ; original 0x249fd; IDA unnamed; logical 0x149fd; file 0x15d6d
    push bx
; Original operand: push    cx
L_10DE: ; original 0x249fe; IDA unnamed; logical 0x149fe; file 0x15d6e
    push cx
; Original operand: push    dx
L_10DF: ; original 0x249ff; IDA unnamed; logical 0x149ff; file 0x15d6f
    push dx
; Original operand: push    di
L_10E0: ; original 0x24a00; IDA unnamed; logical 0x14a00; file 0x15d70
    push di
; Original operand: push    si
L_10E1: ; original 0x24a01; IDA unnamed; logical 0x14a01; file 0x15d71
    push si
; Original operand: push    bp
L_10E2: ; original 0x24a02; IDA unnamed; logical 0x14a02; file 0x15d72
    push bp
; Original operand: mov     ax, cs
L_10E3: ; original 0x24a03; IDA unnamed; logical 0x14a03; file 0x15d73
    mov ax, cs
; Original operand: mov     ds, ax
L_10E5: ; original 0x24a05; IDA unnamed; logical 0x14a05; file 0x15d75
    mov ds, ax
; Original operand: mov     es, ax
L_10E7: ; original 0x24a07; IDA unnamed; logical 0x14a07; file 0x15d77
    mov es, ax
; Original operand: cld
L_10E9: ; original 0x24a09; IDA unnamed; logical 0x14a09; file 0x15d79
    cld 
; Original operand: in      al, 60h; 8042 keyboard controller data register
L_10EA: ; original 0x24a0a; IDA unnamed; logical 0x14a0a; file 0x15d7a
    in al, 060h
; Original operand: or      al, al
L_10EC: ; original 0x24a0c; IDA unnamed; logical 0x14a0c; file 0x15d7c
    or al, al
; Original operand: js      short loc_24A21
L_10EE: ; original 0x24a0e; IDA unnamed; logical 0x14a0e; file 0x15d7e
    js SHORT L_1101
; Original operand: cmp     al, 83h
L_10F0: ; original 0x24a10; IDA unnamed; logical 0x14a10; file 0x15d80
    cmp al, 083h
; Original operand: jnz     short loc_24A21
L_10F2: ; original 0x24a12; IDA unnamed; logical 0x14a12; file 0x15d82
    jnz SHORT L_1101
; Original operand: mov     ah, 2
L_10F4: ; original 0x24a14; IDA unnamed; logical 0x14a14; file 0x15d84
    mov ah, 02h
; Original operand: int     16h; KEYBOARD - GET SHIFT STATUS
L_10F6: ; original 0x24a16; IDA unnamed; logical 0x14a16; file 0x15d86
    int 016h
; Original operand: and     al, 0Ch
L_10F8: ; original 0x24a18; IDA unnamed; logical 0x14a18; file 0x15d88
    and al, 0Ch
; Original operand: cmp     al, 0Ch
L_10FA: ; original 0x24a1a; IDA unnamed; logical 0x14a1a; file 0x15d8a
    cmp al, 0Ch
; Original operand: jnz     short loc_24A21
L_10FC: ; original 0x24a1c; IDA unnamed; logical 0x14a1c; file 0x15d8c
    jnz SHORT L_1101
; Original operand: call    sub_24284
L_10FE: ; original 0x24a1e; IDA unnamed; logical 0x14a1e; file 0x15d8e
    call L_0964
; Original operand: pop     bp
L_1101: ; original 0x24a21; IDA loc_24A21; logical 0x14a21; file 0x15d91
    pop bp
; Original operand: pop     si
L_1102: ; original 0x24a22; IDA unnamed; logical 0x14a22; file 0x15d92
    pop si
; Original operand: pop     di
L_1103: ; original 0x24a23; IDA unnamed; logical 0x14a23; file 0x15d93
    pop di
; Original operand: pop     dx
L_1104: ; original 0x24a24; IDA unnamed; logical 0x14a24; file 0x15d94
    pop dx
; Original operand: pop     cx
L_1105: ; original 0x24a25; IDA unnamed; logical 0x14a25; file 0x15d95
    pop cx
; Original operand: pop     bx
L_1106: ; original 0x24a26; IDA unnamed; logical 0x14a26; file 0x15d96
    pop bx
; Original operand: pop     ax
L_1107: ; original 0x24a27; IDA unnamed; logical 0x14a27; file 0x15d97
    pop ax
; Original operand: pop     es
L_1108: ; original 0x24a28; IDA unnamed; logical 0x14a28; file 0x15d98
    pop es
; Original operand: pop     ds
L_1109: ; original 0x24a29; IDA unnamed; logical 0x14a29; file 0x15d99
    pop ds
; Original operand: jmp     dword ptr cs:byte_23929+9
L_110A: ; original 0x24a2a; IDA unnamed; logical 0x14a2a; file 0x15d9a
    jmp DWORD PTR cs:[012h]
; Original operand: push    ds
L_110F: ; original 0x24a2f; IDA sub_24A2F; logical 0x14a2f; file 0x15d9f
    push ds
; Original operand: push    es
L_1110: ; original 0x24a30; IDA unnamed; logical 0x14a30; file 0x15da0
    push es
; Original operand: push    ax
L_1111: ; original 0x24a31; IDA unnamed; logical 0x14a31; file 0x15da1
    push ax
; Original operand: push    bx
L_1112: ; original 0x24a32; IDA unnamed; logical 0x14a32; file 0x15da2
    push bx
; Original operand: push    cx
L_1113: ; original 0x24a33; IDA unnamed; logical 0x14a33; file 0x15da3
    push cx
; Original operand: push    dx
L_1114: ; original 0x24a34; IDA unnamed; logical 0x14a34; file 0x15da4
    push dx
; Original operand: push    di
L_1115: ; original 0x24a35; IDA unnamed; logical 0x14a35; file 0x15da5
    push di
; Original operand: push    si
L_1116: ; original 0x24a36; IDA unnamed; logical 0x14a36; file 0x15da6
    push si
; Original operand: push    bp
L_1117: ; original 0x24a37; IDA unnamed; logical 0x14a37; file 0x15da7
    push bp
; Original operand: mov     bp, sp
L_1118: ; original 0x24a38; IDA unnamed; logical 0x14a38; file 0x15da8
    mov bp, sp
; Original operand: mov     ax, cs
L_111A: ; original 0x24a3a; IDA unnamed; logical 0x14a3a; file 0x15daa
    mov ax, cs
; Original operand: mov     ds, ax
L_111C: ; original 0x24a3c; IDA unnamed; logical 0x14a3c; file 0x15dac
    mov ds, ax
; Original operand: mov     es, ax
L_111E: ; original 0x24a3e; IDA unnamed; logical 0x14a3e; file 0x15dae
    mov es, ax
; Original operand: mov     ax, [bp+var_sC]
L_1120: ; original 0x24a40; IDA unnamed; logical 0x14a40; file 0x15db0
    mov ax, WORD PTR [bp+0Ch]
; Original operand: cmp     byte_23B04, 0
L_1123: ; original 0x24a43; IDA unnamed; logical 0x14a43; file 0x15db3
    cmp BYTE PTR ds:[01E4h], 00h
; Original operand: jnz     short loc_24A7E
L_1128: ; original 0x24a48; IDA unnamed; logical 0x14a48; file 0x15db8
    jnz SHORT L_115E
; Original operand: mov     byte_23B04, 1
L_112A: ; original 0x24a4a; IDA unnamed; logical 0x14a4a; file 0x15dba
    mov BYTE PTR ds:[01E4h], 01h
; Original operand: sti
L_112F: ; original 0x24a4f; IDA unnamed; logical 0x14a4f; file 0x15dbf
    sti 
; Original operand: cld
L_1130: ; original 0x24a50; IDA unnamed; logical 0x14a50; file 0x15dc0
    cld 
; Original operand: mov     [bp+var_sC], 0FFFFh
L_1131: ; original 0x24a51; IDA unnamed; logical 0x14a51; file 0x15dc1
    mov WORD PTR [bp+0Ch], 0FFFFh
; Original operand: or      bx, bx
L_1136: ; original 0x24a56; IDA unnamed; logical 0x14a56; file 0x15dc6
    or bx, bx
; Original operand: js      short loc_24A67
L_1138: ; original 0x24a58; IDA unnamed; logical 0x14a58; file 0x15dc8
    js SHORT L_1147
; Original operand: cmp     bx, 0Fh
L_113A: ; original 0x24a5a; IDA unnamed; logical 0x14a5a; file 0x15dca
    cmp bx, 0Fh
; Original operand: jnb     short loc_24A77
L_113D: ; original 0x24a5d; IDA unnamed; logical 0x14a5d; file 0x15dcd
    jnb SHORT L_1157
; Original operand: shl     bx, 1
L_113F: ; original 0x24a5f; IDA unnamed; logical 0x14a5f; file 0x15dcf
    shl bx, 01h
; Original operand: call    funcs_24A61[bx]
L_1141: ; original 0x24a61; IDA unnamed; logical 0x14a61; file 0x15dd1
    call WORD PTR [bx+0261h]
; Original operand: jmp     short loc_24A74
L_1145: ; original 0x24a65; IDA unnamed; logical 0x14a65; file 0x15dd5
    jmp SHORT L_1154
; Original operand: not     bx
L_1147: ; original 0x24a67; IDA loc_24A67; logical 0x14a67; file 0x15dd7
    not bx
; Original operand: cmp     bx, 3
L_1149: ; original 0x24a69; IDA unnamed; logical 0x14a69; file 0x15dd9
    cmp bx, 03h
; Original operand: jnb     short loc_24A77
L_114C: ; original 0x24a6c; IDA unnamed; logical 0x14a6c; file 0x15ddc
    jnb SHORT L_1157
; Original operand: shl     bx, 1
L_114E: ; original 0x24a6e; IDA unnamed; logical 0x14a6e; file 0x15dde
    shl bx, 01h
; Original operand: call    funcs_24A70[bx]
L_1150: ; original 0x24a70; IDA unnamed; logical 0x14a70; file 0x15de0
    call WORD PTR [bx+027Fh]
; Original operand: mov     [bp+var_sC], ax
L_1154: ; original 0x24a74; IDA loc_24A74; logical 0x14a74; file 0x15de4
    mov WORD PTR [bp+0Ch], ax
; Original operand: mov     byte_23B04, 0
L_1157: ; original 0x24a77; IDA loc_24A77; logical 0x14a77; file 0x15de7
    mov BYTE PTR ds:[01E4h], 00h
; Original operand: jmp     short loc_24A83
L_115C: ; original 0x24a7c; IDA unnamed; logical 0x14a7c; file 0x15dec
    jmp SHORT L_1163
; Original operand: mov     [bp+var_sC], 0FFF8h
L_115E: ; original 0x24a7e; IDA loc_24A7E; logical 0x14a7e; file 0x15dee
    mov WORD PTR [bp+0Ch], 0FFF8h
; Original operand: pop     bp
L_1163: ; original 0x24a83; IDA loc_24A83; logical 0x14a83; file 0x15df3
    pop bp
; Original operand: pop     si
L_1164: ; original 0x24a84; IDA unnamed; logical 0x14a84; file 0x15df4
    pop si
; Original operand: pop     di
L_1165: ; original 0x24a85; IDA unnamed; logical 0x14a85; file 0x15df5
    pop di
; Original operand: pop     dx
L_1166: ; original 0x24a86; IDA unnamed; logical 0x14a86; file 0x15df6
    pop dx
; Original operand: pop     cx
L_1167: ; original 0x24a87; IDA unnamed; logical 0x14a87; file 0x15df7
    pop cx
; Original operand: pop     bx
L_1168: ; original 0x24a88; IDA unnamed; logical 0x14a88; file 0x15df8
    pop bx
; Original operand: pop     ax
L_1169: ; original 0x24a89; IDA unnamed; logical 0x14a89; file 0x15df9
    pop ax
; Original operand: pop     es
L_116A: ; original 0x24a8a; IDA unnamed; logical 0x14a8a; file 0x15dfa
    pop es
; Original operand: pop     ds
L_116B: ; original 0x24a8b; IDA unnamed; logical 0x14a8b; file 0x15dfb
    pop ds
; Original operand: retf
L_116C: ; original 0x24a8c; IDA unnamed; logical 0x14a8c; file 0x15dfc
    retf 
; Original operand: mov     ax, ds:0Ch
L_116D: ; original 0x24a8d; IDA sub_24A8D; logical 0x14a8d; file 0x15dfd
    mov ax, WORD PTR ds:[0Ch]
; Original operand: retn
L_1170: ; original 0x24a90; IDA unnamed; logical 0x14a90; file 0x15e00
    ret 
; Original operand: xchg    dx, ds:28h
L_1171: ; original 0x24a91; IDA sub_24A91; logical 0x14a91; file 0x15e01
    xchg WORD PTR ds:[028h], dx
; Original operand: xchg    ax, ds:26h
L_1175: ; original 0x24a95; IDA unnamed; logical 0x14a95; file 0x15e05
    xchg WORD PTR ds:[026h], ax
; Original operand: mov     [bp+6], dx
L_1179: ; original 0x24a99; IDA unnamed; logical 0x14a99; file 0x15e09
    mov WORD PTR [bp+06h], dx
; Original operand: push    ax
L_117C: ; original 0x24a9c; IDA unnamed; logical 0x14a9c; file 0x15e0c
    push ax
; Original operand: sub     ax, ax
L_117D: ; original 0x24a9d; IDA unnamed; logical 0x14a9d; file 0x15e0d
    sub ax, ax
; Original operand: call    sub_24333
L_117F: ; original 0x24a9f; IDA unnamed; logical 0x14a9f; file 0x15e0f
    call L_0A13
; Original operand: pop     ax
L_1182: ; original 0x24aa2; IDA unnamed; logical 0x14aa2; file 0x15e12
    pop ax
; Original operand: retn
L_1183: ; original 0x24aa3; IDA unnamed; logical 0x14aa3; file 0x15e13
    ret 
; Original operand: mov     ds:1E0h, cl
L_1184: ; original 0x24aa4; IDA sub_24AA4; logical 0x14aa4; file 0x15e14
    mov BYTE PTR ds:[01E0h], cl
; Original operand: mov     ds:18h, dx
L_1188: ; original 0x24aa8; IDA unnamed; logical 0x14aa8; file 0x15e18
    mov WORD PTR ds:[018h], dx
; Original operand: mov     ds:16h, ax
L_118C: ; original 0x24aac; IDA unnamed; logical 0x14aac; file 0x15e1c
    mov WORD PTR ds:[016h], ax
; Original operand: call    sub_2450E
L_118F: ; original 0x24aaf; IDA unnamed; logical 0x14aaf; file 0x15e1f
    call L_0BEE
; Original operand: call    sub_242CD
L_1192: ; original 0x24ab2; IDA unnamed; logical 0x14ab2; file 0x15e22
    call L_09AD
; Original operand: sub     ax, ax
L_1195: ; original 0x24ab5; IDA unnamed; logical 0x14ab5; file 0x15e25
    sub ax, ax
; Original operand: retn
L_1197: ; original 0x24ab7; IDA unnamed; logical 0x14ab7; file 0x15e27
    ret 
; Original operand: cmp     byte ptr ds:1E1h, 0
L_1198: ; original 0x24ab8; IDA sub_24AB8; logical 0x14ab8; file 0x15e28
    cmp BYTE PTR ds:[01E1h], 00h
; Original operand: jz      short locret_24AD5
L_119D: ; original 0x24abd; IDA unnamed; logical 0x14abd; file 0x15e2d
    jz SHORT L_11B5
; Original operand: pushf
L_119F: ; original 0x24abf; IDA unnamed; logical 0x14abf; file 0x15e2f
    pushf 
; Original operand: cli
L_11A0: ; original 0x24ac0; IDA unnamed; logical 0x14ac0; file 0x15e30
    cli 
; Original operand: sub     al, al
L_11A1: ; original 0x24ac1; IDA unnamed; logical 0x14ac1; file 0x15e31
    sub al, al
; Original operand: mov     ds:1E1h, al
L_11A3: ; original 0x24ac3; IDA unnamed; logical 0x14ac3; file 0x15e33
    mov BYTE PTR ds:[01E1h], al
; Original operand: mov     ax, ds:32h
L_11A6: ; original 0x24ac6; IDA unnamed; logical 0x14ac6; file 0x15e36
    mov ax, WORD PTR ds:[032h]
; Original operand: call    sub_2424D
L_11A9: ; original 0x24ac9; IDA unnamed; logical 0x14ac9; file 0x15e39
    call L_092D
; Original operand: call    sub_24284
L_11AC: ; original 0x24acc; IDA unnamed; logical 0x14acc; file 0x15e3c
    call L_0964
; Original operand: sub     ax, ax
L_11AF: ; original 0x24acf; IDA unnamed; logical 0x14acf; file 0x15e3f
    sub ax, ax
; Original operand: call    sub_24333
L_11B1: ; original 0x24ad1; IDA unnamed; logical 0x14ad1; file 0x15e41
    call L_0A13
; Original operand: popf
L_11B4: ; original 0x24ad4; IDA unnamed; logical 0x14ad4; file 0x15e44
    popf 
; Original operand: retn
L_11B5: ; original 0x24ad5; IDA locret_24AD5; logical 0x14ad5; file 0x15e45
    ret 
; Original operand: mov     ax, 0FFFDh
L_11B6: ; original 0x24ad6; IDA sub_24AD6; logical 0x14ad6; file 0x15e46
    mov ax, 0FFFDh
; Original operand: cmp     byte ptr ds:1E1h, 0
L_11B9: ; original 0x24ad9; IDA unnamed; logical 0x14ad9; file 0x15e49
    cmp BYTE PTR ds:[01E1h], 00h
; Original operand: jz      short locret_24AE5
L_11BE: ; original 0x24ade; IDA unnamed; logical 0x14ade; file 0x15e4e
    jz SHORT L_11C5
; Original operand: call    sub_24AB8
L_11C0: ; original 0x24ae0; IDA unnamed; logical 0x14ae0; file 0x15e50
    call L_1198
; Original operand: sub     ax, ax
L_11C3: ; original 0x24ae3; IDA unnamed; logical 0x14ae3; file 0x15e53
    sub ax, ax
; Original operand: retn
L_11C5: ; original 0x24ae5; IDA locret_24AE5; logical 0x14ae5; file 0x15e55
    ret 
; Original operand: mov     ds:32h, ax
L_11C6: ; original 0x24ae6; IDA sub_24AE6; logical 0x14ae6; file 0x15e56
    mov WORD PTR ds:[032h], ax
; Original operand: retn
L_11C9: ; original 0x24ae9; IDA unnamed; logical 0x14ae9; file 0x15e59
    ret 
; Original operand: mov     ds:34h, ax
L_11CA: ; original 0x24aea; IDA sub_24AEA; logical 0x14aea; file 0x15e5a
    mov WORD PTR ds:[034h], ax
; Original operand: call    sub_2424D
L_11CD: ; original 0x24aed; IDA unnamed; logical 0x14aed; file 0x15e5d
    call L_092D
; Original operand: sub     ax, ax
L_11D0: ; original 0x24af0; IDA unnamed; logical 0x14af0; file 0x15e60
    sub ax, ax
; Original operand: retn
L_11D2: ; original 0x24af2; IDA unnamed; logical 0x14af2; file 0x15e62
    ret 
; Original operand: mov     ds:42h, ax
L_11D3: ; original 0x24af3; IDA sub_24AF3; logical 0x14af3; file 0x15e63
    mov WORD PTR ds:[042h], ax
; Original operand: sub     ax, ax
L_11D6: ; original 0x24af6; IDA unnamed; logical 0x14af6; file 0x15e66
    sub ax, ax
; Original operand: retn
L_11D8: ; original 0x24af8; IDA unnamed; logical 0x14af8; file 0x15e68
    ret 
; Original operand: pushf
L_11D9: ; original 0x24af9; IDA sub_24AF9; logical 0x14af9; file 0x15e69
    pushf 
; Original operand: cli
L_11DA: ; original 0x24afa; IDA unnamed; logical 0x14afa; file 0x15e6a
    cli 
; Original operand: mov     ds:2Ch, dx
L_11DB: ; original 0x24afb; IDA unnamed; logical 0x14afb; file 0x15e6b
    mov WORD PTR ds:[02Ch], dx
; Original operand: mov     ds:2Ah, ax
L_11DF: ; original 0x24aff; IDA unnamed; logical 0x14aff; file 0x15e6f
    mov WORD PTR ds:[02Ah], ax
; Original operand: sub     ax, ax
L_11E2: ; original 0x24b02; IDA unnamed; logical 0x14b02; file 0x15e72
    sub ax, ax
; Original operand: popf
L_11E4: ; original 0x24b04; IDA unnamed; logical 0x14b04; file 0x15e74
    popf 
; Original operand: retn
L_11E5: ; original 0x24b05; IDA unnamed; logical 0x14b05; file 0x15e75
    ret 
; Original operand: mov     word ptr [bp+6], cs
L_11E6: ; original 0x24b06; IDA sub_24B06; logical 0x14b06; file 0x15e76
    mov WORD PTR [bp+06h], cs
; Original operand: mov     ax, 219h
L_11E9: ; original 0x24b09; IDA unnamed; logical 0x14b09; file 0x15e79
    mov ax, 0219h
; Original operand: retn
L_11EC: ; original 0x24b0c; IDA unnamed; logical 0x14b0c; file 0x15e7c
    ret 
; Original operand: pushf
L_11ED: ; original 0x24b0d; IDA sub_24B0D; logical 0x14b0d; file 0x15e7d
    pushf 
; Original operand: cli
L_11EE: ; original 0x24b0e; IDA unnamed; logical 0x14b0e; file 0x15e7e
    cli 
; Original operand: mov     dx, ds:1Ch
L_11EF: ; original 0x24b0f; IDA unnamed; logical 0x14b0f; file 0x15e7f
    mov dx, WORD PTR ds:[01Ch]
; Original operand: mov     ax, ds:1Ah
L_11F3: ; original 0x24b13; IDA unnamed; logical 0x14b13; file 0x15e83
    mov ax, WORD PTR ds:[01Ah]
; Original operand: mov     [bp+6], dx
L_11F6: ; original 0x24b16; IDA unnamed; logical 0x14b16; file 0x15e86
    mov WORD PTR [bp+06h], dx
; Original operand: popf
L_11F9: ; original 0x24b19; IDA unnamed; logical 0x14b19; file 0x15e89
    popf 
; Original operand: retn
L_11FA: ; original 0x24b1a; IDA unnamed; logical 0x14b1a; file 0x15e8a
    ret 
; Original operand: push    dx
L_11FB: ; original 0x24b1b; IDA sub_24B1B; logical 0x14b1b; file 0x15e8b
    push dx
; Original operand: push    ax
L_11FC: ; original 0x24b1c; IDA unnamed; logical 0x14b1c; file 0x15e8c
    push ax
; Original operand: mov     ds:1BAh, di
L_11FD: ; original 0x24b1d; IDA unnamed; logical 0x14b1d; file 0x15e8d
    mov WORD PTR ds:[01BAh], di
; Original operand: mov     ds:1B8h, di
L_1201: ; original 0x24b21; IDA unnamed; logical 0x14b21; file 0x15e91
    mov WORD PTR ds:[01B8h], di
; Original operand: mov     ax, cx
L_1205: ; original 0x24b25; IDA unnamed; logical 0x14b25; file 0x15e95
    mov ax, cx
; Original operand: call    sub_24B85
L_1207: ; original 0x24b27; IDA unnamed; logical 0x14b27; file 0x15e97
    call L_1265
; Original operand: mov     ah, 64h ; 'd'
L_120A: ; original 0x24b2a; IDA unnamed; logical 0x14b2a; file 0x15e9a
    mov ah, 064h
; Original operand: call    sub_24B96
L_120C: ; original 0x24b2c; IDA unnamed; logical 0x14b2c; file 0x15e9c
    call L_1276
; Original operand: mov     ds:1BCh, al
L_120F: ; original 0x24b2f; IDA unnamed; logical 0x14b2f; file 0x15e9f
    mov BYTE PTR ds:[01BCh], al
; Original operand: pop     ax
L_1212: ; original 0x24b32; IDA unnamed; logical 0x14b32; file 0x15ea2
    pop ax
; Original operand: cli
L_1213: ; original 0x24b33; IDA unnamed; logical 0x14b33; file 0x15ea3
    cli 
; Original operand: cmp     ax, 0FFFFh
L_1214: ; original 0x24b34; IDA unnamed; logical 0x14b34; file 0x15ea4
    cmp ax, IMM_FFFF
; Original operand: jnz     short loc_24B3E
L_1217: ; original 0x24b37; IDA unnamed; logical 0x14b37; file 0x15ea7
    jnz SHORT L_121E
; Original operand: mov     al, ds:1B6h
L_1219: ; original 0x24b39; IDA unnamed; logical 0x14b39; file 0x15ea9
    mov al, BYTE PTR ds:[01B6h]
; Original operand: jmp     short loc_24B46
L_121C: ; original 0x24b3c; IDA unnamed; logical 0x14b3c; file 0x15eac
    jmp SHORT L_1226
; Original operand: call    sub_24B85
L_121E: ; original 0x24b3e; IDA loc_24B3E; logical 0x14b3e; file 0x15eae
    call L_1265
; Original operand: mov     ah, 64h ; 'd'
L_1221: ; original 0x24b41; IDA unnamed; logical 0x14b41; file 0x15eb1
    mov ah, 064h
; Original operand: call    sub_24B96
L_1223: ; original 0x24b43; IDA unnamed; logical 0x14b43; file 0x15eb3
    call L_1276
; Original operand: mov     ds:1B6h, al
L_1226: ; original 0x24b46; IDA loc_24B46; logical 0x14b46; file 0x15eb6
    mov BYTE PTR ds:[01B6h], al
; Original operand: mov     ds:1B7h, al
L_1229: ; original 0x24b49; IDA unnamed; logical 0x14b49; file 0x15eb9
    mov BYTE PTR ds:[01B7h], al
; Original operand: sti
L_122C: ; original 0x24b4c; IDA unnamed; logical 0x14b4c; file 0x15ebc
    sti 
; Original operand: mov     cx, 0Bh
L_122D: ; original 0x24b4d; IDA unnamed; logical 0x14b4d; file 0x15ebd
    mov cx, 0Bh
; Original operand: mov     di, 174h
L_1230: ; original 0x24b50; IDA unnamed; logical 0x14b50; file 0x15ec0
    mov di, 0174h
; Original operand: mov     si, 17Fh
L_1233: ; original 0x24b53; IDA unnamed; logical 0x14b53; file 0x15ec3
    mov si, 017Fh
; Original operand: mov     dl, al
L_1236: ; original 0x24b56; IDA unnamed; logical 0x14b56; file 0x15ec6
    mov dl, al
; Original operand: lodsb
L_1238: ; original 0x24b58; IDA loc_24B58; logical 0x14b58; file 0x15ec8
    lodsb
; Original operand: mul     dl
L_1239: ; original 0x24b59; IDA unnamed; logical 0x14b59; file 0x15ec9
    mul dl
; Original operand: add     al, al
L_123B: ; original 0x24b5b; IDA unnamed; logical 0x14b5b; file 0x15ecb
    add al, al
; Original operand: adc     ah, 0
L_123D: ; original 0x24b5d; IDA unnamed; logical 0x14b5d; file 0x15ecd
    adc ah, 00h
; Original operand: mov     al, ah
L_1240: ; original 0x24b60; IDA unnamed; logical 0x14b60; file 0x15ed0
    mov al, ah
; Original operand: stosb
L_1242: ; original 0x24b62; IDA unnamed; logical 0x14b62; file 0x15ed2
    stosb
; Original operand: loop    loc_24B58
L_1243: ; original 0x24b63; IDA unnamed; logical 0x14b63; file 0x15ed3
    loop L_1238
; Original operand: pop     ax
L_1245: ; original 0x24b65; IDA unnamed; logical 0x14b65; file 0x15ed5
    pop ax
; Original operand: call    sub_24B85
L_1246: ; original 0x24b66; IDA unnamed; logical 0x14b66; file 0x15ed6
    call L_1265
; Original operand: mov     ah, 64h ; 'd'
L_1249: ; original 0x24b69; IDA unnamed; logical 0x14b69; file 0x15ed9
    mov ah, 064h
; Original operand: call    sub_24B96
L_124B: ; original 0x24b6b; IDA unnamed; logical 0x14b6b; file 0x15edb
    call L_1276
; Original operand: mov     ds:1B7h, al
L_124E: ; original 0x24b6e; IDA unnamed; logical 0x14b6e; file 0x15ede
    mov BYTE PTR ds:[01B7h], al
; Original operand: retn
L_1251: ; original 0x24b71; IDA unnamed; logical 0x14b71; file 0x15ee1
    ret 
; Original operand: cbw
L_1252: ; original 0x24b72; IDA unnamed; logical 0x14b72; file 0x15ee2
    cbw 
; Original operand: mov     bx, ax
L_1253: ; original 0x24b73; IDA unnamed; logical 0x14b73; file 0x15ee3
    mov bx, ax
; Original operand: shl     bx, 1
L_1255: ; original 0x24b75; IDA unnamed; logical 0x14b75; file 0x15ee5
    shl bx, 01h
; Original operand: shl     bx, 1
L_1257: ; original 0x24b77; IDA unnamed; logical 0x14b77; file 0x15ee7
    shl bx, 01h
; Original operand: mov     ax, [bx+0Eh]
L_1259: ; original 0x24b79; IDA unnamed; logical 0x14b79; file 0x15ee9
    mov ax, WORD PTR [bx+DISP_000E]
; Original operand: mov     dx, [bx+10h]
L_125D: ; original 0x24b7d; IDA unnamed; logical 0x14b7d; file 0x15eed
    mov dx, WORD PTR [bx+DISP_0010]
; Original operand: mov     [bp+6], dx
L_1261: ; original 0x24b81; IDA unnamed; logical 0x14b81; file 0x15ef1
    mov WORD PTR [bp+06h], dx
; Original operand: retn
L_1264: ; original 0x24b84; IDA unnamed; logical 0x14b84; file 0x15ef4
    ret 
; Original operand: or      ax, ax
L_1265: ; original 0x24b85; IDA sub_24B85; logical 0x14b85; file 0x15ef5
    or ax, ax
; Original operand: jns     short loc_24B8D
L_1267: ; original 0x24b87; IDA unnamed; logical 0x14b87; file 0x15ef7
    jns SHORT L_126D
; Original operand: sub     ax, ax
L_1269: ; original 0x24b89; IDA unnamed; logical 0x14b89; file 0x15ef9
    sub ax, ax
; Original operand: jmp     short locret_24B95
L_126B: ; original 0x24b8b; IDA unnamed; logical 0x14b8b; file 0x15efb
    jmp SHORT L_1275
; Original operand: cmp     ax, 64h ; 'd'
L_126D: ; original 0x24b8d; IDA loc_24B8D; logical 0x14b8d; file 0x15efd
    cmp ax, IMM_0064
; Original operand: jb      short locret_24B95
L_1270: ; original 0x24b90; IDA unnamed; logical 0x14b90; file 0x15f00
    jb SHORT L_1275
; Original operand: mov     ax, 63h ; 'c'
L_1272: ; original 0x24b92; IDA unnamed; logical 0x14b92; file 0x15f02
    mov ax, 063h
; Original operand: retn
L_1275: ; original 0x24b95; IDA locret_24B95; logical 0x14b95; file 0x15f05
    ret 
; Original operand: push    cx
L_1276: ; original 0x24b96; IDA sub_24B96; logical 0x14b96; file 0x15f06
    push cx
; Original operand: push    dx
L_1277: ; original 0x24b97; IDA unnamed; logical 0x14b97; file 0x15f07
    push dx
; Original operand: mov     dh, ah
L_1278: ; original 0x24b98; IDA unnamed; logical 0x14b98; file 0x15f08
    mov dh, ah
; Original operand: sub     ah, ah
L_127A: ; original 0x24b9a; IDA unnamed; logical 0x14b9a; file 0x15f0a
    sub ah, ah
; Original operand: div     dh
L_127C: ; original 0x24b9c; IDA unnamed; logical 0x14b9c; file 0x15f0c
    div dh
; Original operand: xchg    al, ah
L_127E: ; original 0x24b9e; IDA unnamed; logical 0x14b9e; file 0x15f0e
    xchg ah, al
; Original operand: sub     dl, dl
L_1280: ; original 0x24ba0; IDA unnamed; logical 0x14ba0; file 0x15f10
    sub dl, dl
; Original operand: mov     cx, 8
L_1282: ; original 0x24ba2; IDA unnamed; logical 0x14ba2; file 0x15f12
    mov cx, 08h
; Original operand: shl     dl, 1
L_1285: ; original 0x24ba5; IDA loc_24BA5; logical 0x14ba5; file 0x15f15
    shl dl, 01h
; Original operand: shl     al, 1
L_1287: ; original 0x24ba7; IDA unnamed; logical 0x14ba7; file 0x15f17
    shl al, 01h
; Original operand: jb      short loc_24BAF
L_1289: ; original 0x24ba9; IDA unnamed; logical 0x14ba9; file 0x15f19
    jb SHORT L_128F
; Original operand: cmp     al, dh
L_128B: ; original 0x24bab; IDA unnamed; logical 0x14bab; file 0x15f1b
    cmp al, dh
; Original operand: jb      short loc_24BB4
L_128D: ; original 0x24bad; IDA unnamed; logical 0x14bad; file 0x15f1d
    jb SHORT L_1294
; Original operand: or      dl, 1
L_128F: ; original 0x24baf; IDA loc_24BAF; logical 0x14baf; file 0x15f1f
    or dl, 01h
; Original operand: sub     al, dh
L_1292: ; original 0x24bb2; IDA unnamed; logical 0x14bb2; file 0x15f22
    sub al, dh
; Original operand: loop    loc_24BA5
L_1294: ; original 0x24bb4; IDA loc_24BB4; logical 0x14bb4; file 0x15f24
    loop L_1285
; Original operand: shl     al, 1
L_1296: ; original 0x24bb6; IDA unnamed; logical 0x14bb6; file 0x15f26
    shl al, 01h
; Original operand: adc     dl, 0
L_1298: ; original 0x24bb8; IDA unnamed; logical 0x14bb8; file 0x15f28
    adc dl, 00h
; Original operand: mov     al, dl
L_129B: ; original 0x24bbb; IDA unnamed; logical 0x14bbb; file 0x15f2b
    mov al, dl
; Original operand: pop     dx
L_129D: ; original 0x24bbd; IDA unnamed; logical 0x14bbd; file 0x15f2d
    pop dx
; Original operand: pop     cx
L_129E: ; original 0x24bbe; IDA unnamed; logical 0x14bbe; file 0x15f2e
    pop cx
; Original operand: retn
L_129F: ; original 0x24bbf; IDA unnamed; logical 0x14bbf; file 0x15f2f
    ret 
; Original operand: mov     cl, ds:1BCh
L_12A0: ; original 0x24bc0; IDA sub_24BC0; logical 0x14bc0; file 0x15f30
    mov cl, BYTE PTR ds:[01BCh]
; Original operand: mov     ax, ds:1B6h
L_12A4: ; original 0x24bc4; IDA unnamed; logical 0x14bc4; file 0x15f34
    mov ax, WORD PTR ds:[01B6h]
; Original operand: cmp     al, ah
L_12A7: ; original 0x24bc7; IDA unnamed; logical 0x14bc7; file 0x15f37
    cmp al, ah
; Original operand: jb      short loc_24BD1
L_12A9: ; original 0x24bc9; IDA unnamed; logical 0x14bc9; file 0x15f39
    jb SHORT L_12B1
; Original operand: sub     al, cl
L_12AB: ; original 0x24bcb; IDA unnamed; logical 0x14bcb; file 0x15f3b
    sub al, cl
; Original operand: jnb     short loc_24BD7
L_12AD: ; original 0x24bcd; IDA unnamed; logical 0x14bcd; file 0x15f3d
    jnb SHORT L_12B7
; Original operand: jmp     short loc_24BD5
L_12AF: ; original 0x24bcf; IDA unnamed; logical 0x14bcf; file 0x15f3f
    jmp SHORT L_12B5
; Original operand: add     al, cl
L_12B1: ; original 0x24bd1; IDA loc_24BD1; logical 0x14bd1; file 0x15f41
    add al, cl
; Original operand: jnb     short loc_24BD7
L_12B3: ; original 0x24bd3; IDA unnamed; logical 0x14bd3; file 0x15f43
    jnb SHORT L_12B7
; Original operand: mov     al, ah
L_12B5: ; original 0x24bd5; IDA loc_24BD5; logical 0x14bd5; file 0x15f45
    mov al, ah
; Original operand: mov     ds:1B6h, al
L_12B7: ; original 0x24bd7; IDA loc_24BD7; logical 0x14bd7; file 0x15f47
    mov BYTE PTR ds:[01B6h], al
; Original operand: mov     dl, al
L_12BA: ; original 0x24bda; IDA unnamed; logical 0x14bda; file 0x15f4a
    mov dl, al
; Original operand: mov     si, 17Fh
L_12BC: ; original 0x24bdc; IDA unnamed; logical 0x14bdc; file 0x15f4c
    mov si, 017Fh
; Original operand: mov     di, 174h
L_12BF: ; original 0x24bdf; IDA unnamed; logical 0x14bdf; file 0x15f4f
    mov di, 0174h
; Original operand: mov     cx, 0Bh
L_12C2: ; original 0x24be2; IDA unnamed; logical 0x14be2; file 0x15f52
    mov cx, 0Bh
; Original operand: lodsb
L_12C5: ; original 0x24be5; IDA loc_24BE5; logical 0x14be5; file 0x15f55
    lodsb
; Original operand: mul     dl
L_12C6: ; original 0x24be6; IDA unnamed; logical 0x14be6; file 0x15f56
    mul dl
; Original operand: add     al, al
L_12C8: ; original 0x24be8; IDA unnamed; logical 0x14be8; file 0x15f58
    add al, al
; Original operand: adc     ah, 0
L_12CA: ; original 0x24bea; IDA unnamed; logical 0x14bea; file 0x15f5a
    adc ah, 00h
; Original operand: mov     al, ah
L_12CD: ; original 0x24bed; IDA unnamed; logical 0x14bed; file 0x15f5d
    mov al, ah
; Original operand: stosb
L_12CF: ; original 0x24bef; IDA unnamed; logical 0x14bef; file 0x15f5f
    stosb
; Original operand: loop    loc_24BE5
L_12D0: ; original 0x24bf0; IDA unnamed; logical 0x14bf0; file 0x15f60
    loop L_12C5
; Original operand: retn
L_12D2: ; original 0x24bf2; IDA unnamed; logical 0x14bf2; file 0x15f62
    ret 
; Unknown data skipped: 0x24bf3..0x24c57
ORG 01337h
; Original operand: push    cx
L_1337: ; original 0x24c57; IDA sub_24C57; logical 0x14c57; file 0x15fc7
    push cx
; Original operand: push    dx
L_1338: ; original 0x24c58; IDA unnamed; logical 0x14c58; file 0x15fc8
    push dx
; Original operand: mov     cx, 40h ; '@'
L_1339: ; original 0x24c59; IDA unnamed; logical 0x14c59; file 0x15fc9
    mov cx, 040h
; Original operand: mov     ah, al
L_133C: ; original 0x24c5c; IDA unnamed; logical 0x14c5c; file 0x15fcc
    mov ah, al
; Original operand: and     ah, 0E0h
L_133E: ; original 0x24c5e; IDA unnamed; logical 0x14c5e; file 0x15fce
    and ah, 0E0h
; Original operand: mov     dx, ds:9
L_1341: ; original 0x24c61; IDA unnamed; logical 0x14c61; file 0x15fd1
    mov dx, WORD PTR ds:[09h]
; Original operand: in      al, dx
L_1345: ; original 0x24c65; IDA loc_24C65; logical 0x14c65; file 0x15fd5
    in al, dx
; Original operand: and     al, 0E0h
L_1346: ; original 0x24c66; IDA unnamed; logical 0x14c66; file 0x15fd6
    and al, 0E0h
; Original operand: cmp     ah, al
L_1348: ; original 0x24c68; IDA unnamed; logical 0x14c68; file 0x15fd8
    cmp ah, al
; Original operand: jz      short loc_24C71
L_134A: ; original 0x24c6a; IDA unnamed; logical 0x14c6a; file 0x15fda
    jz SHORT L_1351
; Original operand: loop    loc_24C65
L_134C: ; original 0x24c6c; IDA unnamed; logical 0x14c6c; file 0x15fdc
    loop L_1345
; Original operand: stc
L_134E: ; original 0x24c6e; IDA unnamed; logical 0x14c6e; file 0x15fde
    stc 
; Original operand: jmp     short loc_24C72
L_134F: ; original 0x24c6f; IDA unnamed; logical 0x14c6f; file 0x15fdf
    jmp SHORT L_1352
; Original operand: clc
L_1351: ; original 0x24c71; IDA loc_24C71; logical 0x14c71; file 0x15fe1
    clc 
; Original operand: pop     dx
L_1352: ; original 0x24c72; IDA loc_24C72; logical 0x14c72; file 0x15fe2
    pop dx
; Original operand: pop     cx
L_1353: ; original 0x24c73; IDA unnamed; logical 0x14c73; file 0x15fe3
    pop cx
; Original operand: retn
L_1354: ; original 0x24c74; IDA unnamed; logical 0x14c74; file 0x15fe4
    ret 
; Original operand: mov     ax, 100h
L_1355: ; original 0x24c75; IDA sub_24C75; logical 0x14c75; file 0x15fe5
    mov ax, 0100h
; Original operand: call    sub_2425E
L_1358: ; original 0x24c78; IDA unnamed; logical 0x14c78; file 0x15fe8
    call L_093E
; Original operand: mov     ax, 460h
L_135B: ; original 0x24c7b; IDA unnamed; logical 0x14c7b; file 0x15feb
    mov ax, 0460h
; Original operand: call    sub_2425E
L_135E: ; original 0x24c7e; IDA unnamed; logical 0x14c7e; file 0x15fee
    call L_093E
; Original operand: mov     ax, 480h
L_1361: ; original 0x24c81; IDA unnamed; logical 0x14c81; file 0x15ff1
    mov ax, 0480h
; Original operand: call    sub_2425E
L_1364: ; original 0x24c84; IDA unnamed; logical 0x14c84; file 0x15ff4
    call L_093E
; Original operand: mov     al, 0
L_1367: ; original 0x24c87; IDA unnamed; logical 0x14c87; file 0x15ff7
    mov al, 00h
; Original operand: call    sub_24C57
L_1369: ; original 0x24c89; IDA unnamed; logical 0x14c89; file 0x15ff9
    call L_1337
; Original operand: jb      short locret_24CAE
L_136C: ; original 0x24c8c; IDA unnamed; logical 0x14c8c; file 0x15ffc
    jb SHORT L_138E
; Original operand: mov     ax, 2FFh
L_136E: ; original 0x24c8e; IDA unnamed; logical 0x14c8e; file 0x15ffe
    mov ax, 02FFh
; Original operand: call    sub_2425E
L_1371: ; original 0x24c91; IDA unnamed; logical 0x14c91; file 0x16001
    call L_093E
; Original operand: mov     ax, 421h
L_1374: ; original 0x24c94; IDA unnamed; logical 0x14c94; file 0x16004
    mov ax, 0421h
; Original operand: call    sub_2425E
L_1377: ; original 0x24c97; IDA unnamed; logical 0x14c97; file 0x16007
    call L_093E
; Original operand: mov     al, 0C0h
L_137A: ; original 0x24c9a; IDA unnamed; logical 0x14c9a; file 0x1600a
    mov al, 0C0h
; Original operand: call    sub_24C57
L_137C: ; original 0x24c9c; IDA unnamed; logical 0x14c9c; file 0x1600c
    call L_1337
; Original operand: jb      short locret_24CAE
L_137F: ; original 0x24c9f; IDA unnamed; logical 0x14c9f; file 0x1600f
    jb SHORT L_138E
; Original operand: mov     ax, 460h
L_1381: ; original 0x24ca1; IDA unnamed; logical 0x14ca1; file 0x16011
    mov ax, 0460h
; Original operand: call    sub_2425E
L_1384: ; original 0x24ca4; IDA unnamed; logical 0x14ca4; file 0x16014
    call L_093E
; Original operand: mov     ax, 480h
L_1387: ; original 0x24ca7; IDA unnamed; logical 0x14ca7; file 0x16017
    mov ax, 0480h
; Original operand: call    sub_2425E
L_138A: ; original 0x24caa; IDA unnamed; logical 0x14caa; file 0x1601a
    call L_093E
; Original operand: clc
L_138D: ; original 0x24cad; IDA unnamed; logical 0x14cad; file 0x1601d
    clc 
; Original operand: retn
L_138E: ; original 0x24cae; IDA locret_24CAE; logical 0x14cae; file 0x1601e
    ret 
; Original operand: not     ax
L_138F: ; original 0x24caf; IDA unnamed; logical 0x14caf; file 0x1601f
    not ax
; Original operand: push    ax
L_1391: ; original 0x24cb1; IDA unnamed; logical 0x14cb1; file 0x16021
    push ax
; Original operand: mov     al, 20h ; ' '
L_1392: ; original 0x24cb2; IDA unnamed; logical 0x14cb2; file 0x16022
    mov al, 020h
; Original operand: out     20h, al; Interrupt controller, 8259A.
L_1394: ; original 0x24cb4; IDA unnamed; logical 0x14cb4; file 0x16024
    out 020h, al
; Original operand: pop     ax
L_1396: ; original 0x24cb6; IDA unnamed; logical 0x14cb6; file 0x16026
    pop ax
; Original operand: iret
L_1397: ; original 0x24cb7; IDA unnamed; logical 0x14cb7; file 0x16027
    iret 
; Original operand: mov     bx, 8
L_1398: ; original 0x24cb8; IDA sub_24CB8; logical 0x14cb8; file 0x16028
    mov bx, 08h
; Original operand: call    sub_2423A
L_139B: ; original 0x24cbb; IDA unnamed; logical 0x14cbb; file 0x1602b
    call L_091A
; Original operand: mov     ds:0Eh, ax
L_139E: ; original 0x24cbe; IDA unnamed; logical 0x14cbe; file 0x1602e
    mov WORD PTR ds:[0Eh], ax
; Original operand: mov     ds:10h, dx
L_13A1: ; original 0x24cc1; IDA unnamed; logical 0x14cc1; file 0x16031
    mov WORD PTR ds:[010h], dx
; Original operand: cli
L_13A5: ; original 0x24cc5; IDA unnamed; logical 0x14cc5; file 0x16035
    cli 
; Original operand: in      al, 21h; Interrupt controller, 8259A.
L_13A6: ; original 0x24cc6; IDA unnamed; logical 0x14cc6; file 0x16036
    in al, 021h
; Original operand: mov     ds:2Eh, ax
L_13A8: ; original 0x24cc8; IDA unnamed; logical 0x14cc8; file 0x16038
    mov WORD PTR ds:[02Eh], ax
; Original operand: mov     al, 0FEh
L_13AB: ; original 0x24ccb; IDA unnamed; logical 0x14ccb; file 0x1603b
    mov al, 0FEh
; Original operand: out     21h, al; Interrupt controller, 8259A.
L_13AD: ; original 0x24ccd; IDA unnamed; logical 0x14ccd; file 0x1603d
    out 021h, al
; Original operand: mov     ax, 1B58h
L_13AF: ; original 0x24ccf; IDA unnamed; logical 0x14ccf; file 0x1603f
    mov ax, 01B58h
; Original operand: call    sub_2424D
L_13B2: ; original 0x24cd2; IDA unnamed; logical 0x14cd2; file 0x16042
    call L_092D
; Original operand: mov     bx, 8
L_13B5: ; original 0x24cd5; IDA unnamed; logical 0x14cd5; file 0x16045
    mov bx, 08h
; Original operand: mov     dx, cs
L_13B8: ; original 0x24cd8; IDA unnamed; logical 0x14cd8; file 0x16048
    mov dx, cs
; Original operand: mov     ax, 138Fh
L_13BA: ; original 0x24cda; IDA unnamed; logical 0x14cda; file 0x1604a
    mov ax, 0138Fh
; Original operand: call    sub_24225
L_13BD: ; original 0x24cdd; IDA unnamed; logical 0x14cdd; file 0x1604d
    call L_0905
; Original operand: sub     ax, ax
L_13C0: ; original 0x24ce0; IDA unnamed; logical 0x14ce0; file 0x16050
    sub ax, ax
; Original operand: sub     cx, cx
L_13C2: ; original 0x24ce2; IDA unnamed; logical 0x14ce2; file 0x16052
    sub cx, cx
; Original operand: sti
L_13C4: ; original 0x24ce4; IDA unnamed; logical 0x14ce4; file 0x16054
    sti 
; Original operand: or      ax, ax
L_13C5: ; original 0x24ce5; IDA loc_24CE5; logical 0x14ce5; file 0x16055
    or ax, ax
; Original operand: jz      short loc_24CE5
L_13C7: ; original 0x24ce7; IDA unnamed; logical 0x14ce7; file 0x16057
    jz SHORT L_13C5
; Original operand: or      ax, ax
L_13C9: ; original 0x24ce9; IDA loc_24CE9; logical 0x14ce9; file 0x16059
    or ax, ax
; Original operand: jnz     short loc_24CE9
L_13CB: ; original 0x24ceb; IDA unnamed; logical 0x14ceb; file 0x1605b
    jnz SHORT L_13C9
; Original operand: nop
L_13CD: ; original 0x24ced; IDA loc_24CED; logical 0x14ced; file 0x1605d
    nop 
; Original operand: inc     cx
L_13CE: ; original 0x24cee; IDA unnamed; logical 0x14cee; file 0x1605e
    inc cx
; Original operand: or      ax, ax
L_13CF: ; original 0x24cef; IDA unnamed; logical 0x14cef; file 0x1605f
    or ax, ax
; Original operand: jz      short loc_24CED
L_13D1: ; original 0x24cf1; IDA unnamed; logical 0x14cf1; file 0x16061
    jz SHORT L_13CD
; Original operand: cli
L_13D3: ; original 0x24cf3; IDA unnamed; logical 0x14cf3; file 0x16063
    cli 
; Original operand: mov     ax, ds:2Eh
L_13D4: ; original 0x24cf4; IDA unnamed; logical 0x14cf4; file 0x16064
    mov ax, WORD PTR ds:[02Eh]
; Original operand: out     21h, al; Interrupt controller, 8259A.
L_13D7: ; original 0x24cf7; IDA unnamed; logical 0x14cf7; file 0x16067
    out 021h, al
; Original operand: mov     ax, ds:32h
L_13D9: ; original 0x24cf9; IDA unnamed; logical 0x14cf9; file 0x16069
    mov ax, WORD PTR ds:[032h]
; Original operand: call    sub_2424D
L_13DC: ; original 0x24cfc; IDA unnamed; logical 0x14cfc; file 0x1606c
    call L_092D
; Original operand: sti
L_13DF: ; original 0x24cff; IDA unnamed; logical 0x14cff; file 0x1606f
    sti 
; Original operand: mov     bx, 8
L_13E0: ; original 0x24d00; IDA unnamed; logical 0x14d00; file 0x16070
    mov bx, 08h
; Original operand: mov     dx, ds:10h
L_13E3: ; original 0x24d03; IDA unnamed; logical 0x14d03; file 0x16073
    mov dx, WORD PTR ds:[010h]
; Original operand: mov     ax, ds:0Eh
L_13E7: ; original 0x24d07; IDA unnamed; logical 0x14d07; file 0x16077
    mov ax, WORD PTR ds:[0Eh]
; Original operand: call    sub_24225
L_13EA: ; original 0x24d0a; IDA unnamed; logical 0x14d0a; file 0x1607a
    call L_0905
; Original operand: mov     ax, cx
L_13ED: ; original 0x24d0d; IDA unnamed; logical 0x14d0d; file 0x1607d
    mov ax, cx
; Original operand: shr     cx, 1
L_13EF: ; original 0x24d0f; IDA unnamed; logical 0x14d0f; file 0x1607f
    shr cx, 01h
; Original operand: shr     cx, 1
L_13F1: ; original 0x24d11; IDA unnamed; logical 0x14d11; file 0x16081
    shr cx, 01h
; Original operand: shr     cx, 1
L_13F3: ; original 0x24d13; IDA unnamed; logical 0x14d13; file 0x16083
    shr cx, 01h
; Original operand: add     ax, cx
L_13F5: ; original 0x24d15; IDA unnamed; logical 0x14d15; file 0x16085
    add ax, cx
; Original operand: mov     cl, 0Ah
L_13F7: ; original 0x24d17; IDA unnamed; logical 0x14d17; file 0x16087
    mov cl, 0Ah
; Original operand: shr     ax, cl
L_13F9: ; original 0x24d19; IDA unnamed; logical 0x14d19; file 0x16089
    shr ax, cl
; Original operand: mov     ds:2Eh, ax
L_13FB: ; original 0x24d1b; IDA unnamed; logical 0x14d1b; file 0x1608b
    mov WORD PTR ds:[02Eh], ax
; Original operand: mov     cx, ax
L_13FE: ; original 0x24d1e; IDA unnamed; logical 0x14d1e; file 0x1608e
    mov cx, ax
; Original operand: shl     cx, 1
L_1400: ; original 0x24d20; IDA unnamed; logical 0x14d20; file 0x16090
    shl cx, 01h
; Original operand: add     ax, cx
L_1402: ; original 0x24d22; IDA unnamed; logical 0x14d22; file 0x16092
    add ax, cx
; Original operand: shl     cx, 1
L_1404: ; original 0x24d24; IDA unnamed; logical 0x14d24; file 0x16094
    shl cx, 01h
; Original operand: add     ax, cx
L_1406: ; original 0x24d26; IDA unnamed; logical 0x14d26; file 0x16096
    add ax, cx
; Original operand: mov     ds:30h, ax
L_1408: ; original 0x24d28; IDA unnamed; logical 0x14d28; file 0x16098
    mov WORD PTR ds:[030h], ax
; Original operand: add     word ptr ds:9, 8
L_140B: ; original 0x24d2b; IDA unnamed; logical 0x14d2b; file 0x1609b
    add WORD PTR ds:[09h], 08h
; Original operand: retn
L_1410: ; original 0x24d30; IDA unnamed; logical 0x14d30; file 0x160a0
    ret 
; Original operand: mov     word ptr ds:32h, 0FFFFh
L_1411: ; original 0x24d31; IDA sub_24D31; logical 0x14d31; file 0x160a1
    mov WORD PTR ds:[032h], 0FFFFh
; Original operand: mov     word ptr ds:36h, 0FFFFh
L_1417: ; original 0x24d37; IDA unnamed; logical 0x14d37; file 0x160a7
    mov WORD PTR ds:[036h], 0FFFFh
; Original operand: mov     word ptr ds:34h, 48D3h
L_141D: ; original 0x24d3d; IDA unnamed; logical 0x14d3d; file 0x160ad
    mov WORD PTR ds:[034h], 048D3h
; Original operand: mov     word ptr ds:28h, cs
L_1423: ; original 0x24d43; IDA unnamed; logical 0x14d43; file 0x160b3
    mov WORD PTR ds:[028h], cs
; Original operand: mov     word ptr ds:26h, 1E5h
L_1427: ; original 0x24d47; IDA unnamed; logical 0x14d47; file 0x160b7
    mov WORD PTR ds:[026h], 01E5h
; Original operand: mov     ax, 120h
L_142D: ; original 0x24d4d; IDA unnamed; logical 0x14d4d; file 0x160bd
    mov ax, 0120h
; Original operand: call    sub_2425E
L_1430: ; original 0x24d50; IDA unnamed; logical 0x14d50; file 0x160c0
    call L_093E
; Original operand: mov     ax, 800h
L_1433: ; original 0x24d53; IDA unnamed; logical 0x14d53; file 0x160c3
    mov ax, 0800h
; Original operand: call    sub_2425E
L_1436: ; original 0x24d56; IDA unnamed; logical 0x14d56; file 0x160c6
    call L_093E
; Original operand: call    sub_24520
L_1439: ; original 0x24d59; IDA unnamed; logical 0x14d59; file 0x160c9
    call L_0C00
; Original operand: sub     ax, ax
L_143C: ; original 0x24d5c; IDA unnamed; logical 0x14d5c; file 0x160cc
    sub ax, ax
; Original operand: retn
L_143E: ; original 0x24d5e; IDA unnamed; logical 0x14d5e; file 0x160ce
    ret 
; Original operand: mov     bx, 8
L_143F: ; original 0x24d5f; IDA sub_24D5F; logical 0x14d5f; file 0x160cf
    mov bx, 08h
; Original operand: call    sub_2423A
L_1442: ; original 0x24d62; IDA unnamed; logical 0x14d62; file 0x160d2
    call L_091A
; Original operand: mov     ds:10h, dx
L_1445: ; original 0x24d65; IDA unnamed; logical 0x14d65; file 0x160d5
    mov WORD PTR ds:[010h], dx
; Original operand: mov     ds:0Eh, ax
L_1449: ; original 0x24d69; IDA unnamed; logical 0x14d69; file 0x160d9
    mov WORD PTR ds:[0Eh], ax
; Original operand: mov     bx, 8
L_144C: ; original 0x24d6c; IDA unnamed; logical 0x14d6c; file 0x160dc
    mov bx, 08h
; Original operand: mov     dx, cs
L_144F: ; original 0x24d6f; IDA unnamed; logical 0x14d6f; file 0x160df
    mov dx, cs
; Original operand: mov     ax, 1052h
L_1451: ; original 0x24d71; IDA unnamed; logical 0x14d71; file 0x160e1
    mov ax, 01052h
; Original operand: call    sub_24225
L_1454: ; original 0x24d74; IDA unnamed; logical 0x14d74; file 0x160e4
    call L_0905
; Original operand: mov     bx, 9
L_1457: ; original 0x24d77; IDA unnamed; logical 0x14d77; file 0x160e7
    mov bx, 09h
; Original operand: call    sub_2423A
L_145A: ; original 0x24d7a; IDA unnamed; logical 0x14d7a; file 0x160ea
    call L_091A
; Original operand: mov     ds:14h, dx
L_145D: ; original 0x24d7d; IDA unnamed; logical 0x14d7d; file 0x160ed
    mov WORD PTR ds:[014h], dx
; Original operand: mov     ds:12h, ax
L_1461: ; original 0x24d81; IDA unnamed; logical 0x14d81; file 0x160f1
    mov WORD PTR ds:[012h], ax
; Original operand: mov     bx, 9
L_1464: ; original 0x24d84; IDA unnamed; logical 0x14d84; file 0x160f4
    mov bx, 09h
; Original operand: mov     dx, cs
L_1467: ; original 0x24d87; IDA unnamed; logical 0x14d87; file 0x160f7
    mov dx, cs
; Original operand: mov     ax, 10DAh
L_1469: ; original 0x24d89; IDA unnamed; logical 0x14d89; file 0x160f9
    mov ax, 010DAh
; Original operand: call    sub_24225
L_146C: ; original 0x24d8c; IDA unnamed; logical 0x14d8c; file 0x160fc
    call L_0905
; Original operand: retn
L_146F: ; original 0x24d8f; IDA unnamed; logical 0x14d8f; file 0x160ff
    ret 
; Original operand: mov     bx, 8
L_1470: ; original 0x24d90; IDA sub_24D90; logical 0x14d90; file 0x16100
    mov bx, 08h
; Original operand: mov     dx, ds:10h
L_1473: ; original 0x24d93; IDA unnamed; logical 0x14d93; file 0x16103
    mov dx, WORD PTR ds:[010h]
; Original operand: mov     ax, ds:0Eh
L_1477: ; original 0x24d97; IDA unnamed; logical 0x14d97; file 0x16107
    mov ax, WORD PTR ds:[0Eh]
; Original operand: call    sub_24225
L_147A: ; original 0x24d9a; IDA unnamed; logical 0x14d9a; file 0x1610a
    call L_0905
; Original operand: mov     bx, 9
L_147D: ; original 0x24d9d; IDA unnamed; logical 0x14d9d; file 0x1610d
    mov bx, 09h
; Original operand: mov     dx, ds:14h
L_1480: ; original 0x24da0; IDA unnamed; logical 0x14da0; file 0x16110
    mov dx, WORD PTR ds:[014h]
; Original operand: mov     ax, ds:12h
L_1484: ; original 0x24da4; IDA unnamed; logical 0x14da4; file 0x16114
    mov ax, WORD PTR ds:[012h]
; Original operand: call    sub_24225
L_1487: ; original 0x24da7; IDA unnamed; logical 0x14da7; file 0x16117
    call L_0905
; Original operand: retn
L_148A: ; original 0x24daa; IDA unnamed; logical 0x14daa; file 0x1611a
    ret 
; Original operand: call    sub_24CB8
L_148B: ; original 0x24dab; IDA sub_24DAB; logical 0x14dab; file 0x1611b
    call L_1398
; Original operand: call    sub_24C75
L_148E: ; original 0x24dae; IDA unnamed; logical 0x14dae; file 0x1611e
    call L_1355
; Original operand: jb      short loc_24DBE
L_1491: ; original 0x24db1; IDA unnamed; logical 0x14db1; file 0x16121
    jb SHORT L_149E
; Original operand: call    sub_24D31
L_1493: ; original 0x24db3; IDA unnamed; logical 0x14db3; file 0x16123
    call L_1411
; Original operand: call    sub_24D5F
L_1496: ; original 0x24db6; IDA unnamed; logical 0x14db6; file 0x16126
    call L_143F
; Original operand: mov     ax, 1
L_1499: ; original 0x24db9; IDA unnamed; logical 0x14db9; file 0x16129
    mov ax, 01h
; Original operand: jmp     short locret_24DC0
L_149C: ; original 0x24dbc; IDA unnamed; logical 0x14dbc; file 0x1612c
    jmp SHORT L_14A0
; Original operand: sub     ax, ax
L_149E: ; original 0x24dbe; IDA loc_24DBE; logical 0x14dbe; file 0x1612e
    sub ax, ax
; Original operand: retn
L_14A0: ; original 0x24dc0; IDA locret_24DC0; logical 0x14dc0; file 0x16130
    ret 
; Original operand: call    sub_24AB8
L_14A1: ; original 0x24dc1; IDA sub_24DC1; logical 0x14dc1; file 0x16131
    call L_1198
; Original operand: call    sub_24D90
L_14A4: ; original 0x24dc4; IDA unnamed; logical 0x14dc4; file 0x16134
    call L_1470
; Original operand: sub     ax, ax
L_14A7: ; original 0x24dc7; IDA unnamed; logical 0x14dc7; file 0x16137
    sub ax, ax
; Original operand: retn
L_14A9: ; original 0x24dc9; IDA unnamed; logical 0x14dc9; file 0x16139
    ret 
; Original operand: mov     ds:9, ax
L_14AA: ; original 0x24dca; IDA sub_24DCA; logical 0x14dca; file 0x1613a
    mov WORD PTR ds:[09h], ax
; Original operand: sub     ax, ax
L_14AD: ; original 0x24dcd; IDA unnamed; logical 0x14dcd; file 0x1613d
    sub ax, ax
; Original operand: retn
L_14AF: ; original 0x24dcf; IDA unnamed; logical 0x14dcf; file 0x1613f
    ret 
_TEXT ENDS
END module_entry
