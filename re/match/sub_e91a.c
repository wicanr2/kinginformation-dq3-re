/* DQ3.EXE, 115282 bytes, SHA256
 * 5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c.
 * IDA Pro 9.4 original sub_1E91A, linear 1E91A..1E91E;
 * logical E91A..E91E, file FC8A..FC8E; exclusive ends, 4 bytes.
 * Confirmed: near CALL original sub_1EC53, then far return; no BP frame.
 * Original callee prototype and full ABI remain unknown; clobbers conservative.
 * Profile: cdecl-size-calls-unframed. No pragma contains instructions.
 * Evidence: docs/25-match-progress.md, primary-accumulator-ida-r1.json.
 */
void sub_ec53(void);
void sub_e91a(void);
#pragma aux sub_ec53 "_*" modify exact [ax bx cx dx si di bp es];
#pragma aux sub_e91a "_*" far modify exact [ax bx cx dx si di bp es];

void sub_e91a(void)
{
    sub_ec53();
}
