/* DQ3 matching source candidate, docs/25. Original sub_24AE6.
 * Input DQ3.EXE: 115282 bytes, SHA256
 * 5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c.
 * IDA9.4 linear 24AE6..24AEA; load-image logical 14AE6..14AEA;
 * file 15E56..15E5A. DS context/product field meaning remain unknown.
 * Register annotations contain no assembly instructions or raw bytes.
 */
extern volatile unsigned unknown_DS_0032;
void sub_14ae6(unsigned value);
#pragma aux sub_14ae6 "_*" parm [ax] modify exact [];

void sub_14ae6(unsigned value)
{
    unknown_DS_0032 = value;
}
