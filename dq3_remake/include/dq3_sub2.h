/* dq3_sub2.h — 子型2 scripted-event NPC 主對話 rec 表(docs/42)。 */
#ifndef DQ3_SUB2_H
#define DQ3_SUB2_H
extern const unsigned char dq3_sub2_rec[];
extern const int dq3_sub2_rec_len;
/* 子型2 NPC(byte4)的主對話 rec(section bank 內;-1=純特殊事件無單純對話)。 */
int dq3_sub2_dialogue(int byte4);
#endif
