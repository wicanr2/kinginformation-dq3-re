/* DQ3.EXE, 115282 bytes, SHA256
 * 5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c.
 * IDA Pro 9.4 original sub_1CFE8, linear 1CFE8..1CFF2;
 * logical CFE8..CFF2, file E358..E362; exclusive ends, 10 bytes.
 * Confirmed static contract: write raw DS:0743 word zero without changing
 * AX, near CALL original sub_1D2AB with incoming AX intact, then near RET.
 * Callee writes AL, preserves AH on the controlled mask-zero return path.
 * Original C prototype/field type/full callee ABI remain unknown.
 * Evidence: docs/25-match-progress.md, writer-abi-ida-r1.json and the
 * explicitly injected, direct-entry writer-abi-oracle-r1 receipt.
 */
extern volatile unsigned unknown_DS_0743;
unsigned sub_d2ab(unsigned incoming_ax);
void sub_cfe8(unsigned incoming_ax);
#pragma aux sub_d2ab "_*" parm [ax] value [ax] modify exact [ax bx cx dx si di bp es];
#pragma aux sub_cfe8 "_*" parm [ax] modify exact [ax bx cx dx si di bp es];

void sub_cfe8(unsigned incoming_ax)
{
    unknown_DS_0743 = 0;
    sub_d2ab(incoming_ax);
}
