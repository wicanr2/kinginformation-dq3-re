# 108 — 戰鬥訊息 pack 化與原始 glyph stream

本批把 dq3_remake_ebitan/game/battle.go 內的玩家可見戰鬥句子遷出 Go，
改由 interface.json.battle_texts 的穩定角色 ID 指向 texts.json。
renderer 不再把 Unicode 字元猜成 glyph，而是直接消費 D3TXT00 的原始
字模／換行／插值控制碼；缺少角色或 text ID 時，NewGameWithPack 會拒絕
啟動（fail closed）。

本輪也把訊息外框接回 `interface.json.battle_message`：DQ3.EXE 的 DGROUP `0x3e6e`
共用 `win_rect` 證實為 `(152,238,360,112)`、inset `(16,16)`、20 欄／4 行。訊息階段
現在隱藏指令／敵名框，只繪製這個 pack-owned rect；更新後的
`dq3_remake_ebitan/docs/battle_message_queue.png` 可與 `references/game3.png` 的下方
訊息區域直接核對，未再把文字塞進左下 150×108 指令框。

## 證據與欄位

| 角色 | D3TXT00 record | 推論等級 | consumer |
|---|---:|---|---|
| 睡眠／醒來 | 349／354 | confirmed | 玩家／怪物行動前狀態 consumer |
| 不會咒文／MP 不足 | 262／366 | confirmed | 指令與咒文 MP gate |
| 逃跑、受傷、攻擊、沒打中 | 348／332／330／335 | confirmed | 敵我回合訊息 |
| 勝利／經驗／金錢 | 353／372／373 | confirmed | 戰鬥結算與 reward consumer |
| 封咒、睡眠、力量／守備、混亂 | 351／340／338／319／326／396 | strong | 敵我狀態與 buff consumer；逐 action 停頓仍待影片 V3 |
| 解毒／解麻痺、無效、空道具 | 364／399／346／271 | confirmed | 咒文／藥草失敗結果 |
| 帕魯朋特吹走／吸 MP | 317／318 | confirmed | execPalpunte 五個有效 slot |
| 我方全滅 | 361 | confirmed | wipedOut 與原版 di=0x169 |

以上 record 的 raw words、offset 與輸入檔雜湊可回查
docs/data/d3txt_codes.json 與 docs/script/txt00.txt；原始位址基準是
D3TXT00.TXT record，不把它冒充 IDA linear address。EXE 的 victory／敗北
consumer 見 docs/13-exe-battle.md。

## Runtime 契約

battleMessage 保存 pack value 供 log／測試，但畫面只讀 glyph_codes。
每一個原始 VAR control 消耗一個 runtime glyph slice；未設定的變數保留一格
空白，換行／換頁按原始 control words 處理。玩家／怪物名稱由目前戰鬥的
角色 glyph 或 D3TXT00 monster-name record 提供，數字由共用十進位 glyph
轉換器提供。

本批只閉合文字資料、插值 renderer 與訊息外框；戰鬥的逐動作停頓、動畫／音效 cue、完整
抗性矩陣與所有 formation 的原版同狀態 V3 仍列在
docs/74-ebiten-remake-completion-plan.md 長尾工作，不因 pack 化而宣稱 remake 已完成。
