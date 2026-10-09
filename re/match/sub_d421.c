/* DQ3.EXE, 115282 bytes, SHA256
 * 5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c.
 * IDA Pro 9.4 original sub_1D421, linear 1D421..1D433;
 * logical D421..D433, file E791..E7A3; exclusive ends, 18 bytes.
 * Confirmed static contract: load raw DS:0743 word once, scale by two,
 * compare DS:[BX+38F4] word to zero, and if nonzero read it again for
 * a near indirect CALL; near return. Table reads remain volatile.
 * Original table extent, valid indices, runtime DS and callback ABI unknown.
 * Evidence: docs/25-match-progress.md, conditional-ida-r1.json.
 */
typedef void (__near *unknown_near_callback)(void);
typedef char near_callback_must_be_word[(sizeof(unknown_near_callback) == 2) ? 1 : -1];
extern unknown_near_callback volatile unknown_DS_38f4[];
extern volatile unsigned unknown_DS_0743;
void sub_d421(void);
#pragma aux sub_d421 "_*" modify exact [ax bx cx dx si di bp es];

void sub_d421(void)
{
    unsigned index = unknown_DS_0743;
    if (unknown_DS_38f4[index])
        unknown_DS_38f4[index]();
}
