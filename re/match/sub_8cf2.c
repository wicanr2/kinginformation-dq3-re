/* DQ3.EXE, 115282 bytes, SHA256
 * 5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c.
 * IDA Pro 9.4 original sub_18CF2, linear 18CF2..18CFC;
 * logical 8CF2..8CFC, file A062..A06C; exclusive ends, 10 bytes.
 * Confirmed: CX=4, near CALL original sub_1EF95, near CALL original
 * sub_1EF03, then near return. Original prototypes/full ABI unknown.
 * Clobbers conservative; no pragma contains instructions.
 * Evidence: docs/25-match-progress.md, register-calls-ida-r1.json.
 */
void sub_ef95(unsigned value);
void sub_ef03(void);
void sub_8cf2(void);
#pragma aux sub_ef95 "_*" parm [cx] modify exact [ax bx cx dx si di bp es];
#pragma aux sub_ef03 "_*" modify exact [ax bx cx dx si di bp es];
#pragma aux sub_8cf2 "_*" modify exact [ax bx cx dx si di bp es];

void sub_8cf2(void)
{
    sub_ef95(4);
    sub_ef03();
}
