/* DQ3 primary-code candidate, docs/25. Original sub_16FCF.
 * Input DQ3.EXE: 115282 bytes, SHA256
 * 5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c.
 * IDA9.4 linear 16FCF..16FDE; logical 6FCF..6FDE;
 * file 833F..834E. AH supplies a target byte, SI a six-byte source.
 * BL reports the first matching 1-based index, or 7 when none matches.
 * Original also leaves AL=current byte, SI advanced, CX=remaining count.
 * Actual byte matching must validate these implicit register effects.
 */
unsigned char sub_6fcf(unsigned incoming_ax, const unsigned char __near *source);
#pragma aux sub_6fcf "_*" parm [ax] [si] value [bl] modify exact [ax bx cx si];

unsigned char sub_6fcf(unsigned incoming_ax, const unsigned char __near *source)
{
    unsigned char target = (unsigned char)(incoming_ax >> 8);
    unsigned char index = 1;
    unsigned count = 6;
    do {
        unsigned char current = *source++;
        if (current == target)
            return index;
        ++index;
    } while (--count);
    return index;
}
