# Matching decompilation 來源

這裡保存提交到GitHub的C與組語來源。下表分開標示完整函式與已驗證的指令區段，
完整EXE還原仍未完成。主程式須用C；組語只用於已確認底層用途。

| 來源 | 原版IDA9.4定位 | 已驗證範圍 | 限制 |
|---|---|---|---|
| [sub_5d49.c](sub_5d49.c) | sub_15D49，linear15D49..15D50 | 完整C PROC，7 bytes | MZ入口code segment；產品角色未知，compiler ENDP後NOP配置未解 |
| [sub_136f5.c](sub_136f5.c) | sub_236F5，linear236F5..236F9 | 完整C PROC，4 bytes | 原版CTVMEM.ASM driver；欄位、型別與DS脈絡未知 |
| [sub_14a8d.c](sub_14a8d.c) | sub_24A8D，linear24A8D..24A91 | 完整C PROC，4 bytes | 原版CMFDRV.ASM driver；欄位、型別與DS脈絡未知 |
| [sub_14ae6.c](sub_14ae6.c) | sub_24AE6，linear24AE6..24AEA | 完整C module，4 bytes | 原版CMFDRV.ASM driver；AX入參word store，欄位、型別與DS脈絡未知 |
| [sub_e6b9.asm](sub_e6b9.asm) | sub_1E6B9，linear1E6B9..1E6C9 | 七條語意指令，16 bytes | 局部RNG核心；不證明原版來源語言或完整campaign |
| [sub_e6c9.asm](sub_e6c9.asm) | sub_1E6C9，linear1E6C9..1E6E7 | 完整有界RNG，30 bytes | BX入參及AX／DX結果已局部驗證；完整Goal底層資格仍需分類 |
| [ctvmem_code.asm](ctvmem_code.asm)、[encoding常數](ctvmem_constants.asm)與[typed data](ctvmem_data.json) | CTVMEM driver，linear22F60..2391D | 完整byte-layout，2493 bytes | literal6B06、11-byte unknown payload保持，原始semantics／hardware runtime未知 |
| [cmfdrv_code.asm](cmfdrv_code.asm)、[encoding常數](cmfdrv_constants.asm)與[typed data](cmfdrv_data.json) | CMFDRV driver，linear23920..24DD0 | 完整byte-layout，5296 bytes | code與2406 data從來源重建相同；原始field semantics及完整runtime硬體parity未知 |

其他C檔與rng_adapter.asm是研究候選。它們有DIFF、REFUSED或僅register/state等價的結果，
不計精確匹配。原版OBJ、EXE、SDK、IDA database與生成封包不提交。

兩個driver的ASM只保存語意指令與EQU常數，沒有原版code array或DB拼接。
ORG所保留的空間沒有還原資料，不能直接當成可執行的完整driver。
兩driver完整recipe以[common verifier](../../tools/verify_sdk_source_module.py)把已審typed data放回data區段，
從source完整重建CTV2493及CMF5296 bytes，不使用原版objects。
instruction-only profile仍省略data；完整recipe與data semantics聲明分開。
兩次source／OBJ／MZ／FIXUPP重編一致，逐原始指令位置匹配5148 bytes。
其中12 bytes與上表三個C helper重疊；新增唯一指令來源為5136 bytes，不能重複計數。

原始輸入為 `assets_raw/DQ3.EXE`，115282 bytes，SHA-256
`5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
每份來源保留原始logical／file定位；linear基準0x10000，file=`linear-0x10000+0x1370`。

C來源清單與真正FIXUPP重定位由[MSC manifest](../../tools/matching_c_manifest.json)、
[Watcom manifest](../../tools/watcom_matching_manifest.json)及其runner核對。
SDK instruction來源以[範圍manifest](../../tools/sdk_instruction_source_manifest.json)及
[source verifier](../../tools/verify_sdk_instruction_sources.py)核對實際OMF written mask，拒絕資料／code array匯入。
重建一律使用[隔離工具鏈入口](../../tools/build/README.md)；可用MSC、Watcom16及NASM，
不以原版machine-code拼接代替來源。
原版SDK只用於[module身分證據](../../docs/25-match-progress.md)，不作最終原碼重建輸入。

[研究與驗證範圍](../../docs/25-match-progress.md)、
[完整Goal契約](../../tools/matching_goal_contract.json)與
[目前狀態](../../CONTEXT.md)保存所有未完成事項。
