/* DQ3.EXE, 115282 bytes, SHA256
 * 5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c.
 * IDA Pro 9.4 original sub_1EEFF, linear 1EEFF..1EF03;
 * logical EEFF..EF03, file 1026F..10273; exclusive ends, 4 bytes.
 * Confirmed: near CALL original sub_1EF03, then far return; no BP frame.
 * Original callee prototype and full ABI remain unknown; clobbers conservative.
 * Profile: cdecl-size-calls-unframed. No pragma contains instructions.
 * Evidence: docs/25-match-progress.md, primary-accumulator-ida-r1.json.
 */
void sub_ef03(void);
void sub_eeff(void);
#pragma aux sub_ef03 "_*" modify exact [ax bx cx dx si di bp es];
#pragma aux sub_eeff "_*" far modify exact [ax bx cx dx si di bp es];

void sub_eeff(void)
{
    sub_ef03();
}
