; DQ3.EXE 115282 bytes, SHA256
; 5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c
; IDA9.4 seg005 region linear209CE..20B60, logical109CE..10B60,
; file11D3E..11ED0; MZ frame109C, first entry offset000E.
; Confirmed low-level BIOS/video-port/memory service instructions.
; Original object identity, complete ABI and runtime hardware parity unknown.
; Scope: docs/25-match-progress.md READY source spec; no instruction bytes.
; Alignment byte is supplied separately by typed JSON at region+401.
.8086
VIDEO_CODE SEGMENT BYTE PUBLIC 'CODE'
PUBLIC region_entry
region_entry:
sub_209CE:
    mov     ah, 0 ; IDA linear 0x209ce, file 0x11d3e
original_209D0:
    mov     al, 10h ; IDA linear 0x209d0, file 0x11d40
original_209D2:
    int     10h ; IDA linear 0x209d2, file 0x11d42
original_209D4:
    mov     ah, 5 ; IDA linear 0x209d4, file 0x11d44
original_209D6:
    mov     al, 0 ; IDA linear 0x209d6, file 0x11d46
original_209D8:
    int     10h ; IDA linear 0x209d8, file 0x11d48
original_209DA:
    retf ; IDA linear 0x209da, file 0x11d4a
original_209DB:
    mov     ah, 0 ; IDA linear 0x209db, file 0x11d4b
original_209DD:
    mov     al, 10h ; IDA linear 0x209dd, file 0x11d4d
original_209DF:
    int     10h ; IDA linear 0x209df, file 0x11d4f
original_209E1:
    mov     ah, 5 ; IDA linear 0x209e1, file 0x11d51
original_209E3:
    mov     al, 1 ; IDA linear 0x209e3, file 0x11d53
original_209E5:
    int     10h ; IDA linear 0x209e5, file 0x11d55
original_209E7:
    retf ; IDA linear 0x209e7, file 0x11d57
sub_209E8:
    mov     ah, 0 ; IDA linear 0x209e8, file 0x11d58
original_209EA:
    mov     al, 3 ; IDA linear 0x209ea, file 0x11d5a
original_209EC:
    int     10h ; IDA linear 0x209ec, file 0x11d5c
original_209EE:
    mov     ah, 5 ; IDA linear 0x209ee, file 0x11d5e
original_209F0:
    mov     al, 0 ; IDA linear 0x209f0, file 0x11d60
original_209F2:
    int     10h ; IDA linear 0x209f2, file 0x11d62
original_209F4:
    retf ; IDA linear 0x209f4, file 0x11d64
sub_209F5:
    push    es ; IDA linear 0x209f5, file 0x11d65
original_209F6:
    mov     ax, 0B000h ; IDA linear 0x209f6, file 0x11d66
original_209F9:
    mov     es, ax ; IDA linear 0x209f9, file 0x11d69
original_209FB:
    mov     cx, 7D0h ; IDA linear 0x209fb, file 0x11d6b
original_209FE:
    xor     di, di ; IDA linear 0x209fe, file 0x11d6e
original_20A00:
    mov     ax, 720h ; IDA linear 0x20a00, file 0x11d70
original_20A03:
    rep stosw ; IDA linear 0x20a03, file 0x11d73
original_20A05:
    pop     es ; IDA linear 0x20a05, file 0x11d75
original_20A06:
    retf ; IDA linear 0x20a06, file 0x11d76
sub_20A07:
    mov     dx, 3CEh ; IDA linear 0x20a07, file 0x11d77
original_20A0A:
    mov     ax, 0FF08h ; IDA linear 0x20a0a, file 0x11d7a
original_20A0D:
    out     dx, ax ; IDA linear 0x20a0d, file 0x11d7d
original_20A0E:
    mov     ax, 5 ; IDA linear 0x20a0e, file 0x11d7e
original_20A11:
    out     dx, ax ; IDA linear 0x20a11, file 0x11d81
original_20A12:
    mov     ax, 3 ; IDA linear 0x20a12, file 0x11d82
original_20A15:
    out     dx, ax ; IDA linear 0x20a15, file 0x11d85
original_20A16:
    push    ax ; IDA linear 0x20a16, file 0x11d86
original_20A17:
    push    cx ; IDA linear 0x20a17, file 0x11d87
original_20A18:
    push    dx ; IDA linear 0x20a18, file 0x11d88
original_20A19:
    push    di ; IDA linear 0x20a19, file 0x11d89
original_20A1A:
    push    es ; IDA linear 0x20a1a, file 0x11d8a
original_20A1B:
    xor     ax, ax ; IDA linear 0x20a1b, file 0x11d8b
original_20A1D:
    mov     es, ax ; IDA linear 0x20a1d, file 0x11d8d
original_20A1F:
    mov     cx, es:[44Ch] ; IDA linear 0x20a1f, file 0x11d8f
original_20A24:
    pop     es ; IDA linear 0x20a24, file 0x11d94
original_20A25:
    mov     cx, 0FFFFh ; IDA linear 0x20a25, file 0x11d95
original_20A28:
    mov     dx, 3C4h ; IDA linear 0x20a28, file 0x11d98
original_20A2B:
    mov     ax, 0F02h ; IDA linear 0x20a2b, file 0x11d9b
original_20A2E:
    out     dx, ax ; IDA linear 0x20a2e, file 0x11d9e
original_20A2F:
    xor     ax, ax ; IDA linear 0x20a2f, file 0x11d9f
original_20A31:
    xor     di, di ; IDA linear 0x20a31, file 0x11da1
original_20A33:
    rep stosw ; IDA linear 0x20a33, file 0x11da3
original_20A35:
    pop     di ; IDA linear 0x20a35, file 0x11da5
original_20A36:
    pop     dx ; IDA linear 0x20a36, file 0x11da6
original_20A37:
    pop     cx ; IDA linear 0x20a37, file 0x11da7
original_20A38:
    pop     ax ; IDA linear 0x20a38, file 0x11da8
original_20A39:
    retf ; IDA linear 0x20a39, file 0x11da9
sub_20A3A:
    push    es ; IDA linear 0x20a3a, file 0x11daa
original_20A3B:
    mov     ax, ds ; IDA linear 0x20a3b, file 0x11dab
original_20A3D:
    mov     es, ax ; IDA linear 0x20a3d, file 0x11dad
original_20A3F:
    mov     ah, 10h ; IDA linear 0x20a3f, file 0x11daf
original_20A41:
    mov     al, 12h ; IDA linear 0x20a41, file 0x11db1
original_20A43:
    mov     bx, 0 ; IDA linear 0x20a43, file 0x11db3
original_20A46:
    mov     cx, 10h ; IDA linear 0x20a46, file 0x11db6
original_20A49:
    int     10h ; IDA linear 0x20a49, file 0x11db9
original_20A4B:
    pop     es ; IDA linear 0x20a4b, file 0x11dbb
original_20A4C:
    retf ; IDA linear 0x20a4c, file 0x11dbc
sub_20A4D:
    push    es ; IDA linear 0x20a4d, file 0x11dbd
original_20A4E:
    mov     ax, ds ; IDA linear 0x20a4e, file 0x11dbe
original_20A50:
    mov     es, ax ; IDA linear 0x20a50, file 0x11dc0
original_20A52:
    mov     ah, 10h ; IDA linear 0x20a52, file 0x11dc2
original_20A54:
    mov     al, 12h ; IDA linear 0x20a54, file 0x11dc4
original_20A56:
    mov     bx, 0 ; IDA linear 0x20a56, file 0x11dc6
original_20A59:
    mov     cx, 100h ; IDA linear 0x20a59, file 0x11dc9
original_20A5C:
    int     10h ; IDA linear 0x20a5c, file 0x11dcc
original_20A5E:
    pop     es ; IDA linear 0x20a5e, file 0x11dce
original_20A5F:
    retf ; IDA linear 0x20a5f, file 0x11dcf
original_20A60:
    mov     ah, 10h ; IDA linear 0x20a60, file 0x11dd0
original_20A62:
    mov     al, 10h ; IDA linear 0x20a62, file 0x11dd2
original_20A64:
    int     10h ; IDA linear 0x20a64, file 0x11dd4
original_20A66:
    retf ; IDA linear 0x20a66, file 0x11dd6
sub_20A67:
    call    sub_20A70 ; IDA linear 0x20a67, file 0x11dd7
original_20A6A:
    mov     bh, 0 ; IDA linear 0x20a6a, file 0x11dda
original_20A6C:
    call    sub_20A7B ; IDA linear 0x20a6c, file 0x11ddc
original_20A6F:
    retf ; IDA linear 0x20a6f, file 0x11ddf
sub_20A70:
    mov     ah, 10h ; IDA linear 0x20a70, file 0x11de0
original_20A72:
    mov     al, 13h ; IDA linear 0x20a72, file 0x11de2
original_20A74:
    mov     bl, 0 ; IDA linear 0x20a74, file 0x11de4
original_20A76:
    mov     bh, 1 ; IDA linear 0x20a76, file 0x11de6
original_20A78:
    int     10h ; IDA linear 0x20a78, file 0x11de8
original_20A7A:
    retn ; IDA linear 0x20a7a, file 0x11dea
sub_20A7B:
    mov     ah, 10h ; IDA linear 0x20a7b, file 0x11deb
original_20A7D:
    mov     al, 13h ; IDA linear 0x20a7d, file 0x11ded
original_20A7F:
    mov     bl, 1 ; IDA linear 0x20a7f, file 0x11def
original_20A81:
    int     10h ; IDA linear 0x20a81, file 0x11df1
original_20A83:
    retn ; IDA linear 0x20a83, file 0x11df3
sub_20A84:
    push    es ; IDA linear 0x20a84, file 0x11df4
original_20A85:
    push    ds ; IDA linear 0x20a85, file 0x11df5
original_20A86:
    pop     es ; IDA linear 0x20a86, file 0x11df6
original_20A87:
    lea     dx, ds:[25B4h] ; IDA linear 0x20a87, file 0x11df7
original_20A8B:
    mov     ah, 10h ; IDA linear 0x20a8b, file 0x11dfb
original_20A8D:
    mov     al, 2 ; IDA linear 0x20a8d, file 0x11dfd
original_20A8F:
    int     10h ; IDA linear 0x20a8f, file 0x11dff
original_20A91:
    pop     es ; IDA linear 0x20a91, file 0x11e01
original_20A92:
    retf ; IDA linear 0x20a92, file 0x11e02
sub_20A93:
    push    es ; IDA linear 0x20a93, file 0x11e03
original_20A94:
    mov     dx, 40h ; IDA linear 0x20a94, file 0x11e04
original_20A97:
    mov     es, dx ; IDA linear 0x20a97, file 0x11e07
original_20A99:
    mov     dx, es:[63h] ; IDA linear 0x20a99, file 0x11e09
original_20A9E:
    add     dl, 6 ; IDA linear 0x20a9e, file 0x11e0e
loc_20AA1:
    in      al, dx ; IDA linear 0x20aa1, file 0x11e11
original_20AA2:
    test    al, 8 ; IDA linear 0x20aa2, file 0x11e12
original_20AA4:
    jz      short loc_20AA1 ; IDA linear 0x20aa4, file 0x11e14
loc_20AA6:
    in      al, dx ; IDA linear 0x20aa6, file 0x11e16
original_20AA7:
    test    al, 8 ; IDA linear 0x20aa7, file 0x11e17
original_20AA9:
    jnz     short loc_20AA6 ; IDA linear 0x20aa9, file 0x11e19
original_20AAB:
    cli ; IDA linear 0x20aab, file 0x11e1b
original_20AAC:
    sub     dl, 6 ; IDA linear 0x20aac, file 0x11e1c
original_20AAF:
    mov     ah, bh ; IDA linear 0x20aaf, file 0x11e1f
original_20AB1:
    mov     al, 0Ch ; IDA linear 0x20ab1, file 0x11e21
original_20AB3:
    out     dx, ax ; IDA linear 0x20ab3, file 0x11e23
original_20AB4:
    mov     ah, bl ; IDA linear 0x20ab4, file 0x11e24
original_20AB6:
    inc     al ; IDA linear 0x20ab6, file 0x11e26
original_20AB8:
    out     dx, ax ; IDA linear 0x20ab8, file 0x11e28
original_20AB9:
    sti ; IDA linear 0x20ab9, file 0x11e29
original_20ABA:
    add     dl, 6 ; IDA linear 0x20aba, file 0x11e2a
loc_20ABD:
    in      al, dx ; IDA linear 0x20abd, file 0x11e2d
original_20ABE:
    test    al, 8 ; IDA linear 0x20abe, file 0x11e2e
original_20AC0:
    jz      short loc_20ABD ; IDA linear 0x20ac0, file 0x11e30
original_20AC2:
    sti ; IDA linear 0x20ac2, file 0x11e32
original_20AC3:
    pop     es ; IDA linear 0x20ac3, file 0x11e33
original_20AC4:
    retf ; IDA linear 0x20ac4, file 0x11e34
sub_20AC5:
    mov     cl, ds:[25F1h] ; IDA linear 0x20ac5, file 0x11e35
original_20AC9:
    xor     ch, ch ; IDA linear 0x20ac9, file 0x11e39
original_20ACB:
    push    es ; IDA linear 0x20acb, file 0x11e3b
original_20ACC:
    mov     dx, 40h ; IDA linear 0x20acc, file 0x11e3c
original_20ACF:
    mov     es, dx ; IDA linear 0x20acf, file 0x11e3f
original_20AD1:
    mov     dx, es:[63h] ; IDA linear 0x20ad1, file 0x11e41
original_20AD6:
    add     dl, 6 ; IDA linear 0x20ad6, file 0x11e46
original_20AD9:
    in      al, dx ; IDA linear 0x20ad9, file 0x11e49
original_20ADA:
    in      al, dx ; IDA linear 0x20ada, file 0x11e4a
original_20ADB:
    cli ; IDA linear 0x20adb, file 0x11e4b
original_20ADC:
    sub     dl, 6 ; IDA linear 0x20adc, file 0x11e4c
original_20ADF:
    mov     ah, bh ; IDA linear 0x20adf, file 0x11e4f
original_20AE1:
    mov     al, 0Ch ; IDA linear 0x20ae1, file 0x11e51
original_20AE3:
    out     dx, ax ; IDA linear 0x20ae3, file 0x11e53
original_20AE4:
    mov     ah, bl ; IDA linear 0x20ae4, file 0x11e54
original_20AE6:
    inc     al ; IDA linear 0x20ae6, file 0x11e56
original_20AE8:
    out     dx, ax ; IDA linear 0x20ae8, file 0x11e58
original_20AE9:
    sti ; IDA linear 0x20ae9, file 0x11e59
original_20AEA:
    add     dl, 6 ; IDA linear 0x20aea, file 0x11e5a
original_20AED:
    in      al, dx ; IDA linear 0x20aed, file 0x11e5d
original_20AEE:
    cli ; IDA linear 0x20aee, file 0x11e5e
original_20AEF:
    sub     dl, 6 ; IDA linear 0x20aef, file 0x11e5f
original_20AF2:
    mov     ah, ch ; IDA linear 0x20af2, file 0x11e62
original_20AF4:
    mov     al, 8 ; IDA linear 0x20af4, file 0x11e64
original_20AF6:
    out     dx, ax ; IDA linear 0x20af6, file 0x11e66
original_20AF7:
    mov     dl, 0C0h ; IDA linear 0x20af7, file 0x11e67
original_20AF9:
    mov     al, 33h ; IDA linear 0x20af9, file 0x11e69
original_20AFB:
    out     dx, al ; IDA linear 0x20afb, file 0x11e6b
original_20AFC:
    mov     al, cl ; IDA linear 0x20afc, file 0x11e6c
original_20AFE:
    and     al, 7 ; IDA linear 0x20afe, file 0x11e6e
original_20B00:
    out     dx, al ; IDA linear 0x20b00, file 0x11e70
original_20B01:
    sti ; IDA linear 0x20b01, file 0x11e71
original_20B02:
    pop     es ; IDA linear 0x20b02, file 0x11e72
original_20B03:
    retf ; IDA linear 0x20b03, file 0x11e73
sub_20B04:
    mov     dx, 54h ; IDA linear 0x20b04, file 0x11e74
original_20B07:
    mov     cx, 140h ; IDA linear 0x20b07, file 0x11e77
loc_20B0A:
    push    es ; IDA linear 0x20b0a, file 0x11e7a
original_20B0B:
    mov     ax, 40h ; IDA linear 0x20b0b, file 0x11e7b
original_20B0E:
    mov     es, ax ; IDA linear 0x20b0e, file 0x11e7e
original_20B10:
    mov     es:[4Ah], dl ; IDA linear 0x20b10, file 0x11e80
original_20B15:
    mov     bh, dl ; IDA linear 0x20b15, file 0x11e85
original_20B17:
    mov     ax, cx ; IDA linear 0x20b17, file 0x11e87
original_20B19:
    dec     al ; IDA linear 0x20b19, file 0x11e89
original_20B1B:
    mov     es:[84h], al ; IDA linear 0x20b1b, file 0x11e8b
original_20B1F:
    inc     al ; IDA linear 0x20b1f, file 0x11e8f
original_20B21:
    mul     bh ; IDA linear 0x20b21, file 0x11e91
original_20B23:
    mov     es:[4Ch], ax ; IDA linear 0x20b23, file 0x11e93
original_20B27:
    mov     ah, bh ; IDA linear 0x20b27, file 0x11e97
original_20B29:
    shr     ah, 1 ; IDA linear 0x20b29, file 0x11e99
original_20B2B:
    mov     al, 13h ; IDA linear 0x20b2b, file 0x11e9b
original_20B2D:
    mov     dx, es:[63h] ; IDA linear 0x20b2d, file 0x11e9d
original_20B32:
    out     dx, ax ; IDA linear 0x20b32, file 0x11ea2
original_20B33:
    pop     es ; IDA linear 0x20b33, file 0x11ea3
original_20B34:
    mov     dx, 3CEh ; IDA linear 0x20b34, file 0x11ea4
original_20B37:
    mov     ax, 3 ; IDA linear 0x20b37, file 0x11ea7
original_20B3A:
    out     dx, ax ; IDA linear 0x20b3a, file 0x11eaa
original_20B3B:
    mov     ax, 805h ; IDA linear 0x20b3b, file 0x11eab
original_20B3E:
    out     dx, ax ; IDA linear 0x20b3e, file 0x11eae
original_20B3F:
    mov     ax, 7 ; IDA linear 0x20b3f, file 0x11eaf
original_20B42:
    out     dx, ax ; IDA linear 0x20b42, file 0x11eb2
original_20B43:
    mov     ax, 0FF08h ; IDA linear 0x20b43, file 0x11eb3
original_20B46:
    out     dx, ax ; IDA linear 0x20b46, file 0x11eb6
original_20B47:
    mov     dl, 0C4h ; IDA linear 0x20b47, file 0x11eb7
original_20B49:
    retf ; IDA linear 0x20b49, file 0x11eb9
sub_20B4A:
    jmp     short loc_20B0A ; IDA linear 0x20b4a, file 0x11eba
original_20B4C:
    retf ; IDA linear 0x20b4c, file 0x11ebc
sub_20B4D:
    mov     ah, 12h ; IDA linear 0x20b4d, file 0x11ebd
original_20B4F:
    mov     bl, 36h ; IDA linear 0x20b4f, file 0x11ebf
original_20B51:
    mov     al, 0 ; IDA linear 0x20b51, file 0x11ec1
original_20B53:
    int     10h ; IDA linear 0x20b53, file 0x11ec3
original_20B55:
    retf ; IDA linear 0x20b55, file 0x11ec5
sub_20B56:
    mov     ah, 12h ; IDA linear 0x20b56, file 0x11ec6
original_20B58:
    mov     bl, 36h ; IDA linear 0x20b58, file 0x11ec8
original_20B5A:
    mov     al, 1 ; IDA linear 0x20b5a, file 0x11eca
original_20B5C:
    int     10h ; IDA linear 0x20b5c, file 0x11ecc
original_20B5E:
    retf ; IDA linear 0x20b5e, file 0x11ece
VIDEO_CODE ENDS
END
