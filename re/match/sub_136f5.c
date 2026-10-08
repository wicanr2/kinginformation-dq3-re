/* @match entry=0x136f5 model=/AS
 * DQ3.EXE, 115282 bytes, SHA256
 * 5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c.
 * IDA9.4 original sub_236F5, linear 236F5..236F9;
 * load-image logical 136F5..136F9, file 14A65..14A69.
 * DS context and the field's product meaning remain unknown.
 * Observed operation: load DS-relative word 0033 into AX, near return.
 * This source does not establish the original source type/compiler.
 */
extern volatile unsigned unknown_DS_0033;

unsigned sub_136f5(void)
{
    return unknown_DS_0033;
}
