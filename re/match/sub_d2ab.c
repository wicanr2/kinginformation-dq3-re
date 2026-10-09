/* DQ3.EXE, 115282 bytes, SHA256
 * 5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c.
 * IDA Pro 9.4 original sub_1D2AB, linear 1D2AB..1D2D3;
 * logical D2AB..D2D3, file E61B..E643; exclusive ends, 40 bytes.
 * Confirmed static contract: two volatile DS:24B5 word reads produce
 * the indexed byte at DS:37C5; replace AL, preserve incoming AH, mask
 * AL with 18h and dispatch 08/18/10 to the original near call targets.
 * AX is passed to and returned from each callee at the machine boundary.
 * The union is an authored AX representation, not a recovered original type.
 * Original record type/extent, valid indices, runtime DS/full ABI unknown.
 * Evidence: docs/25-match-progress.md, writer-abi-ida-r1.json.
 */
extern volatile unsigned unknown_DS_24b5;
extern volatile unsigned char unknown_DS_37c5[];
unsigned sub_d2d3(unsigned value);
unsigned sub_d338(unsigned value);
unsigned sub_d3b7(unsigned value);
unsigned sub_d2ab(unsigned incoming_ax);
#pragma aux sub_d2d3 "_*" parm [ax] value [ax] modify exact [ax bx cx dx si di bp es];
#pragma aux sub_d338 "_*" parm [ax] value [ax] modify exact [ax bx cx dx si di bp es];
#pragma aux sub_d3b7 "_*" parm [ax] value [ax] modify exact [ax bx cx dx si di bp es];
#pragma aux sub_d2ab "_*" parm [ax] value [ax] modify exact [ax bx cx dx si di bp es];

union word_parts {
    unsigned word;
    struct { unsigned char low, high; } bytes;
};
typedef char word_parts_must_be_word[(sizeof(union word_parts) == 2) ? 1 : -1];

unsigned sub_d2ab(unsigned incoming_ax)
{
    union word_parts value;
    value.word = incoming_ax;
    value.bytes.low = unknown_DS_37c5[unknown_DS_24b5 * 2u + unknown_DS_24b5];
    value.bytes.low &= 0x18;
    if (value.bytes.low == 8)
        return sub_d2d3(value.word);
    if (value.bytes.low == 0x18)
        return sub_d338(value.word);
    if (value.bytes.low == 0x10)
        return sub_d3b7(value.word);
    return value.word;
}
