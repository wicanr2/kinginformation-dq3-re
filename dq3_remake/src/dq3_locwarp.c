/* dq3_locwarp.c — location scripted warp 資料(tools/gen_locwarp.py 產;勿手改)。docs/44 §5b。 */
#include "dq3_locwarp.h"

/* {src_cty, dest_cty, x, y, gate_type(0未明/1flag/2runner), gate_val} */
const dq3_locwarp dq3_locwarps[8] = {
  {  31,  6,  89,  1, 0, 0x00 },  /* call@0x56ea 源CTY31 → 6 阿莎拉慕 */
  {  24,  1,  26,  1, 1, 0x2e },  /* call@0x680b 源CTY24 → 1 雷貝(flag 0x2e)*/
  {  26,  5,  75,  1, 2, 0x01 },  /* call@0x6d27 源CTY26 → 5 精靈村(runner ev1)*/
  {  35,  5,  75,  1, 1, 0x44 },  /* call@0x7035 源CTY35 → 5 精靈村(flag 0x44)*/
  {  12,  5,  67,  1, 0, 0x00 },  /* call@0x71bf 源CTY12 → 5 精靈村 */
  {  24,  1,  57,  4, 2, 0x01 },  /* call@0x7285 源CTY24 → 1 雷貝(runner ev1)*/
  {  24,  1,  56,  1, 0, 0x00 },  /* call@0x737e 源CTY24 → 1 雷貝 */
  {  39,  6, 121,  1, 0, 0x00 },  /* call@0x75a7 源CTY39 → 6 阿莎拉慕 */
};
