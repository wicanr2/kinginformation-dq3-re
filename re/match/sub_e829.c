/* DQ3.EXE, 115282 bytes, SHA256
 * 5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c.
 * IDA Pro 9.4 original sub_1E829, linear 1E829..1E82D;
 * logical E829..E82D, file FB99..FB9D; exclusive ends, 4 bytes.
 * Confirmed: near CALL original sub_1E82D, then far return; no BP frame.
 * Original callee prototype and full ABI remain unknown; clobbers conservative.
 * Profile: cdecl-size-calls-unframed. No pragma contains instructions.
 * Evidence: docs/25-match-progress.md, primary-accumulator-ida-r1.json.
 */
void sub_e82d(void);
void sub_e829(void);
#pragma aux sub_e82d "_*" modify exact [ax bx cx dx si di bp es];
#pragma aux sub_e829 "_*" far modify exact [ax bx cx dx si di bp es];

void sub_e829(void)
{
    sub_e82d();
}
