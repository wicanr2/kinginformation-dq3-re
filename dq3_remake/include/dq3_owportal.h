/* dq3_owportal.h — overworld 旗標條件 portal(同一地表點依進度載不同城,城鎮變體)。
 * RE: 函式 0x396e(docs/44 §7、docs/45 §3.2)。dispatcher 0x39cb: [0x256c]=目的CTY; load_cty。
 * 非下降(目的多為地表城);remake 旗標未驅動時取 default 分支(早期變體)。 */
#ifndef DQ3_OWPORTAL_H
#define DQ3_OWPORTAL_H
#include "dq3_inventory.h"   /* dq3_storyflags */

/* 玩家在 overworld (x,y) → 依旗標解出該載入的 CTY;非 portal 位置回 -1。 */
int dq3_owportal_resolve(int x, int y, const dq3_storyflags *flags);

#endif
