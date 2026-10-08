# Matching decompilation 來源

這裡保存提交到GitHub的C與組語來源。下表只列完整函式bytes已精確匹配的來源，
完整EXE還原仍未完成。主程式須用C；組語的完整Goal資格另依已確認底層用途審查。

| 來源 | 原版IDA9.4定位 | 已驗證範圍 | 限制 |
|---|---|---|---|
| [sub_5d49.c](sub_5d49.c) | sub_15D49，linear15D49..15D50 | 完整C PROC，7 bytes | MZ入口code segment；產品角色未知，compiler ENDP後NOP配置未解 |
| [sub_136f5.c](sub_136f5.c) | sub_236F5，linear236F5..236F9 | 完整C PROC，4 bytes | 原版CTVMEM.ASM driver；欄位、型別與DS脈絡未知 |
| [sub_14a8d.c](sub_14a8d.c) | sub_24A8D，linear24A8D..24A91 | 完整C PROC，4 bytes | 原版CMFDRV.ASM driver；欄位、型別與DS脈絡未知 |
| [sub_14ae6.c](sub_14ae6.c) | sub_24AE6，linear24AE6..24AEA | 完整C module，4 bytes | 原版CMFDRV.ASM driver；AX入參word store，欄位、型別與DS脈絡未知 |
| [sub_e6b9.asm](sub_e6b9.asm) | sub_1E6B9，linear1E6B9..1E6C9 | 七條語意指令，16 bytes | 局部RNG核心；不證明原版來源語言或完整campaign |
| [sub_e6c9.asm](sub_e6c9.asm) | sub_1E6C9，linear1E6C9..1E6E7 | 完整有界RNG，30 bytes | BX入參及AX／DX結果已局部驗證；完整Goal底層資格仍需分類 |

其他C檔與rng_adapter.asm是研究候選。它們有DIFF、REFUSED或僅register/state等價的結果，
不計精確匹配。原版OBJ、EXE、SDK、IDA database與生成封包不提交。

原始輸入為 `assets_raw/DQ3.EXE`，115282 bytes，SHA-256
`5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
每份來源保留原始logical／file定位；linear基準0x10000，file=`linear-0x10000+0x1370`。

C來源清單與真正FIXUPP重定位由[MSC manifest](../../tools/matching_c_manifest.json)、
[Watcom manifest](../../tools/watcom_matching_manifest.json)及其runner核對。
重建一律使用[隔離工具鏈入口](../../tools/build/README.md)；可用MSC、Watcom16及NASM，
不以原版machine-code拼接代替來源。
原版SDK只用於[module身分證據](../../docs/25-match-progress.md)，不作最終原碼重建輸入。

[研究與驗證範圍](../../docs/25-match-progress.md)、
[完整Goal契約](../../tools/matching_goal_contract.json)與
[目前狀態](../../CONTEXT.md)保存所有未完成事項。
