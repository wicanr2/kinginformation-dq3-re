/* DQ3.EXE, 115282 bytes, SHA256
 * 5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c.
 * IDA Pro 9.4 original sub_1EE15, linear 1EE15..1EE19;
 * logical EE15..EE19, file 10185..10189; exclusive ends, 4 bytes.
 * Confirmed: near CALL original sub_1EE76, then near return.
 * Original callee prototype and full ABI remain unknown; clobbers conservative.
 * Evidence: docs/25-match-progress.md, primary-accumulator-ida-r1.json.
 */
void sub_ee76(void);
void sub_ee15(void);
#pragma aux sub_ee76 "_*" modify exact [ax bx cx dx si di bp es];
#pragma aux sub_ee15 "_*" modify exact [ax bx cx dx si di bp es];

void sub_ee15(void)
{
    sub_ee76();
}
