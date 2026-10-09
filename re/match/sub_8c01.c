/* DQ3.EXE, 115282 bytes, SHA256
 * 5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c.
 * IDA Pro 9.4 original sub_18C01, linear 18C01..18C0F;
 * logical 8C01..8C0F, file 9F71..9F7F; exclusive ends, 14 bytes.
 * Confirmed: DS:24FC word to BX then near CALL original sub_16EF4;
 * DI=01E4 then near CALL original sub_15023; near return.
 * Original field type/prototypes/full ABI unknown; clobbers conservative.
 * Evidence: docs/25-match-progress.md, register-calls-ida-r2.json.
 */
extern volatile unsigned unknown_DS_24fc;
void sub_6ef4(unsigned value);
void sub_5023(unsigned value);
void sub_8c01(void);
#pragma aux sub_6ef4 "_*" parm [bx] modify exact [ax bx cx dx si di bp es];
#pragma aux sub_5023 "_*" parm [di] modify exact [ax bx cx dx si di bp es];
#pragma aux sub_8c01 "_*" modify exact [ax bx cx dx si di bp es];

void sub_8c01(void)
{
    sub_6ef4(unknown_DS_24fc);
    sub_5023(0x1e4);
}
