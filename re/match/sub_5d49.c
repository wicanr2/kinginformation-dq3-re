/* @match entry=0x5d49 model=/AS
 * DQ3.EXE, 115282 bytes, SHA256
 * 5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c.
 * IDA9.4 original sub_15D49, linear 15D49..15D50;
 * logical 5D49..5D50, file 70B9..70C0.
 * This source describes the observed word store and near return only.
 * The data symbol remains unknown until its caller/writer/consumer closes.
 * Exact codegen does not identify the original compiler/source language.
 */
extern volatile unsigned unknown_DGROUP_0b60;

void sub_5d49(void)
{
    unknown_DGROUP_0b60 = 1;
}
