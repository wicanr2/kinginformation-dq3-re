/* DQ3.EXE, 115282 bytes, SHA256
 * 5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c.
 * IDA Pro 9.4 original sub_1EE19, linear 1EE19..1EE23;
 * logical EE19..EE23, file 10189..10193; exclusive ends, 10 bytes.
 * Confirmed: DS:25D1 word goes to DX, then far CALL 109C:007A,
 * original IDA callee sub_20A3A, then near return.
 * Original C types/prototypes remain unknown; clobbers are conservative.
 * No pragma contains instructions. Profile: cdecl-size-calls.
 * Evidence: docs/25-match-progress.md, primary-call-ida-r1 export.
 */
extern volatile unsigned unknown_DS_25d1;
void sub_20a3a(unsigned value);
#pragma aux sub_20a3a "_*" far parm [dx] modify exact [ax bx cx dx si di bp es];
void sub_ee19(void);
#pragma aux sub_ee19 "_*" modify exact [ax bx cx dx si di bp es];

void sub_ee19(void)
{
    sub_20a3a(unknown_DS_25d1);
}
