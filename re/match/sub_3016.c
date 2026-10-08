/* DQ3.EXE, 115282 bytes, SHA256
 * 5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c.
 * IDA Pro 9.4 original sub_13016, linear 13016..13023;
 * logical 3016..3023, file 4386..4393; exclusive ends, 13 bytes.
 * Confirmed: four near calls in this order, then near return.
 * C symbols retain logical-address identity; their original IDA callees are
 * sub_13023, sub_130CF, sub_131BD and sub_1333C, respectively.
 * Clobber declarations are conservative, not recovered original prototypes.
 * No pragma contains instructions. Profile: cdecl-size-calls.
 * Evidence: docs/25-match-progress.md, primary-call-ida-r1 export.
 */
void sub_3023(void);
void sub_30cf(void);
void sub_31bd(void);
void sub_333c(void);
void sub_3016(void);
#pragma aux sub_3023 "_*" modify exact [ax bx cx dx si di bp es];
#pragma aux sub_30cf "_*" modify exact [ax bx cx dx si di bp es];
#pragma aux sub_31bd "_*" modify exact [ax bx cx dx si di bp es];
#pragma aux sub_333c "_*" modify exact [ax bx cx dx si di bp es];
#pragma aux sub_3016 "_*" modify exact [ax bx cx dx si di bp es];

void sub_3016(void)
{
    sub_3023();
    sub_30cf();
    sub_31bd();
    sub_333c();
}
