/* DQ3.EXE, 115282 bytes, SHA256
 * 5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c.
 * IDA Pro 9.4 original sub_14AE5, linear 14AE5..14AF6;
 * logical 4AE5..4AF6, file 5E55..5E66; exclusive ends, 17 bytes.
 * Confirmed: near CALL original sub_14AF6, read DS:0726 byte;
 * if not 1, near CALL sub_14AB5 then sub_14C2B; near return.
 * Original state meaning/prototypes/full ABI unknown; clobbers conservative.
 * Evidence: docs/25-match-progress.md, conditional-ida-r1.json.
 */
extern volatile unsigned char unknown_DS_0726;
void sub_4af6(void);
void sub_4ab5(void);
void sub_4c2b(void);
void sub_4ae5(void);
#pragma aux sub_4af6 "_*" modify exact [ax bx cx dx si di bp es];
#pragma aux sub_4ab5 "_*" modify exact [ax bx cx dx si di bp es];
#pragma aux sub_4c2b "_*" modify exact [ax bx cx dx si di bp es];
#pragma aux sub_4ae5 "_*" modify exact [ax bx cx dx si di bp es];

void sub_4ae5(void)
{
    sub_4af6();
    if (unknown_DS_0726 != 1) {
        sub_4ab5();
        sub_4c2b();
    }
}
