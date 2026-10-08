/* DQ3.EXE, 115282 bytes, SHA256
 * 5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c.
 * IDA Pro 9.4 original sub_19834, linear 19834..19842;
 * logical 9834..9842, file ABA4..ABB2; exclusive ends.
 * Confirmed: DS:259C minus one, doubled, indexes words at DS:4F15;
 * SI receives the word; BX is preserved. No instructions in the pragma.
 * Scope: valid caller indices into the original near-pointer table.
 * Original declaration, table extent and compiler remain unknown.
 * Evidence: docs/25-match-progress.md, primary-table-ida-r1 export.
 */
extern volatile unsigned unknown_DS_259c;
extern volatile unsigned unknown_DS_4f15[];

unsigned sub_9834(void);
#pragma aux sub_9834 "_*" value [si] modify exact [si];

unsigned sub_9834(void)
{
    unsigned index = unknown_DS_259c - 1;
    return unknown_DS_4f15[index];
}
