/* DQ3.EXE, 115282 bytes, SHA256
 * 5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c.
 * IDA Pro 9.4 original sub_137F9, linear 137F9..13829;
 * logical 37F9..3829, file 4B69..4B99; exclusive ends, 48 bytes.
 * Confirmed: SI points to writable words, BX supplies two word stores;
 * AX/CX are saved and restored. The original record type remains unknown.
 * The void function's non-AX value register suppresses Watcom's default AX
 * clobber. This pragma contains no instructions; SI itself is unchanged.
 * Build profile: watcall-speed-no-reorder in the Watcom manifest.
 * Evidence: docs/25-match-progress.md, primary-record-ida-r1 export.
 */
extern volatile unsigned unknown_DS_072f;
extern volatile unsigned unknown_DS_0731;
extern volatile unsigned unknown_DS_0733;

void sub_37f9(volatile unsigned *record, unsigned incoming_bx);
#pragma aux sub_37f9 "_*" parm [si] [bx] value [si] modify exact [];

void sub_37f9(volatile unsigned *record, unsigned incoming_bx)
{
    unsigned value;
    value = unknown_DS_072f;
    record[1] = value;
    value += 2;
    record[12] = value;
    value = unknown_DS_0731;
    record[2] = value;
    value += 16;
    record[13] = value;
    value = unknown_DS_0733;
    record[6] = incoming_bx;
    record[10] = incoming_bx;
    value += 2;
    record[4] = value << 4;
}
