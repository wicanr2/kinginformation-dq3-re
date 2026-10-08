/* @match entry=0x14a8d model=/AS
 * DQ3.EXE, 115282 bytes, SHA256
 * 5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c.
 * IDA9.4 original sub_24A8D, linear 24A8D..24A91;
 * load-image logical 14A8D..14A91, file 15DFD..15E01.
 * DS context and the field's product meaning remain unknown.
 * Observed operation: load DS-relative word 000C into AX, near return.
 * This source does not establish the original source type/compiler.
 */
extern volatile unsigned unknown_DS_000c;

unsigned sub_14a8d(void)
{
    return unknown_DS_000c;
}
