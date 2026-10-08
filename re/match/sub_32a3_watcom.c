/* DQ3 primary-code candidate, docs/25. Original sub_132A3.
 * Input DQ3.EXE: 115282 bytes, SHA256
 * 5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c.
 * IDA9.4 linear 132A3..132B4; logical 32A3..32B4;
 * file 4613..4624. Original operation: 13 byte stores at offsets
 * 8,10,...,32 from DS:265D, leaving BX=34 and CX=0.
 * Product field meaning and source types remain unknown.
 */
extern volatile unsigned char unknown_DS_265d[];
unsigned sub_32a3(void);
#pragma aux sub_32a3 "_*" value [cx] modify exact [bx cx];

unsigned sub_32a3(void)
{
    unsigned index = 8;
    unsigned count = 13;
    do {
        unknown_DS_265d[index] = 255;
        index += 2;
    } while (--count);
    return count;
}
