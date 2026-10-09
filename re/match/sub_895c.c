/* DQ3.EXE, 115282 bytes, SHA256
 * 5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c.
 * IDA Pro 9.4 original sub_1895C, linear 1895C..18966;
 * logical 895C..8966, file 9CCC..9CD6; exclusive ends, 10 bytes.
 * Confirmed static operation: sign-extend AX into DX:AX, add to the
 * DS:4F37/4F39 word pair modulo 2^32, then near return. AX is unchanged.
 * WCC16 candidate int/unsigned long widths reproduce that operation;
 * original field type, compiler and full runtime DS context remain unknown.
 * Evidence: docs/25-match-progress.md, primary-accumulator-ida-r1.json.
 */
extern volatile unsigned long unknown_DS_4f37;
void sub_895c(int delta);
#pragma aux sub_895c "_*" parm [ax] modify exact [dx];

void sub_895c(int delta)
{
    unknown_DS_4f37 += (long)delta;
}
