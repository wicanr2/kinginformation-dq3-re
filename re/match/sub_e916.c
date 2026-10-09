/* DQ3.EXE, 115282 bytes, SHA256
 * 5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c.
 * IDA Pro 9.4 original sub_1E916, linear 1E916..1E91A;
 * logical E916..E91A, file FC86..FC8A; exclusive ends, 4 bytes.
 * Confirmed: near CALL original sub_1EA8C, then far return; no BP frame.
 * Original callee prototype and full ABI remain unknown; clobbers conservative.
 * Profile: cdecl-size-calls-unframed. No pragma contains instructions.
 * Evidence: docs/25-match-progress.md, primary-accumulator-ida-r1.json.
 */
void sub_ea8c(void);
void sub_e916(void);
#pragma aux sub_ea8c "_*" modify exact [ax bx cx dx si di bp es];
#pragma aux sub_e916 "_*" far modify exact [ax bx cx dx si di bp es];

void sub_e916(void)
{
    sub_ea8c();
}
