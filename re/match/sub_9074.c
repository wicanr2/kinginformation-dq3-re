/* DQ3.EXE, 115282 bytes, SHA256
 * 5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c.
 * IDA Pro 9.4 original sub_19074, linear 19074..19081;
 * logical 9074..9081, file A3E4..A3F1; exclusive ends, 13 bytes.
 * Confirmed: near CALL original sub_1BF35, write word 00FF to raw
 * DS:26FE, near CALL original sub_19090, then near return.
 * Original field type/prototypes/full ABI unknown; clobbers conservative.
 * Evidence: docs/25-match-progress.md, register-calls-ida-r2.json.
 */
extern volatile unsigned unknown_DS_26fe;
void sub_bf35(void);
void sub_9090(void);
void sub_9074(void);
#pragma aux sub_bf35 "_*" modify exact [ax bx cx dx si di bp es];
#pragma aux sub_9090 "_*" modify exact [ax bx cx dx si di bp es];
#pragma aux sub_9074 "_*" modify exact [ax bx cx dx si di bp es];

void sub_9074(void)
{
    sub_bf35();
    unknown_DS_26fe = 0xff;
    sub_9090();
}
