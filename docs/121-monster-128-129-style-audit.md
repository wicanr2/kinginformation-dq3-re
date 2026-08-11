# 128／129 怪物 sprite 風格與格式稽核

本文件只描述外部回補圖的可追溯產物，不把回補圖冒充成原版美術，也不改寫
`assets_raw/DQ3MNS.SHP`。原始素材與 IDA 資料庫仍依專案規則留在本機。

## Git 歷史證據

| 提交 | 可追溯結論 |
| --- | --- |
| `7296241` | 先前圖鑑以 128 格展示。 |
| `01f023c` | 歷史提交記錄 `id 128` 歐里狄加與 `id 129` 五頭龍大王為未完成槽位，依參考圖重繪成 16 色 4-plane planar，注入 `DQ3MNS.SHP` 以避免缺圖當機。 |
| `b8c6331` | C/SDL 歷史實作曾以空槽回退圖確認「不當機」到「真顯示」；這是舊實作線索，不是現行 Go runtime 的規格。 |
| `8a538a5` | 圖鑑更新為 130 格，並將 `id 129` 調成與實機索瑪戰相符的綠色 King Hydra。 |

以上是提交說明所聲稱的歷史；本輪沒有用它取代原版檔案的證據。原始檔
`assets_raw/DQ3MNS.SHP`（SHA-256 `bad40e552343141b75191a5a9576adddb7a16ddf1378fe139eba4c4e15ba8bfd`）
仍保持不變；這是原始基線，不直接當作執行版。回補／執行版 SHP 另存於 ignored 的
`work/DQ3MNS_fixed.SHP`，並同步到 ignored 的 `dq3_remake_ebitan/mobile/assets/DQ3MNS.SHP`，
不加入 Git 或公開 release。

## 本輪產物與風格判定

`tools/make_sprites.py` 的目前規則已經把兩張外部參考圖量化到 `MNSBK.PAL` 前 16 色，並
以 `plane 3 → plane 0` 的順序寫回 SHP；`id 129` 的紫／藍色索引另外重映成實機可見的綠色。
本輪再以 `tools/style_sprites.py` 做局部風格整理：保留原 opaque mask、輪廓、邏輯尺寸與
角色姿態，只把 128 的白色 alias 8→15、藍色 alias 4→13 統一到原生常用索引，並在已有
透明外框內側加入固定相位的低密度 checker cluster／深色邊界像素；129 只加入相同規則的綠色
明暗 cluster，不新增部位、不平滑、不使用 AI 生成。兩支工具的規則同步，重新執行後得到：

| 槽位 | 原始邏輯尺寸 | 展示 PNG（3× nearest） | SHA-256 |
| ---: | ---: | ---: | --- |
| 128 | 96×88 | [`spr_128.png`](monsters/spr_128.png)（288×264） | `dbf98b0b4b625a3ccb25e24bfbb8530ba0949e856bbe9f0dfff88ef604c7b1bc` |
| 129 | 128×96 | [`spr_129.png`](monsters/spr_129.png)（384×288） | `e5a86a9de52fdaa4ed321f87fcb00fcaf059c8b2e3afdafc54b32a4a827496f7` |

| 產物 | SHA-256 |
| --- | --- |
| `work/DQ3MNS_fixed.SHP`（本機、不入 Git；含 128／129 RLE mask） | `eb846763ee5d5b1582d0e54678f238539b02e997f359f618d833ce8bac2b8adb` |
| `dq3_remake_ebitan/mobile/assets/DQ3MNS.SHP`（embedded 執行版副本） | `eb846763ee5d5b1582d0e54678f238539b02e997f359f618d833ce8bac2b8adb` |
| `dq3_remake/src/dq3_restored_sprites.c`（歷史 C fallback） | `4d483a2cea9e4e653054fde27c57a2529828b93b9b30834a55b468d6b4140f6e` |
| `dq3_remake_ebitan/internal/dq3data/restored_sprites.go` | `8fb0f764ecc0b1e639be142e9792e45be23293ea86bdf68e6837abf5a2cbcd93` |
| `spr_128.png`（288×264） | `dbf98b0b4b625a3ccb25e24bfbb8530ba0949e856bbe9f0dfff88ef604c7b1bc` |
| `spr_129.png`（384×288） | `e5a86a9de52fdaa4ed321f87fcb00fcaf059c8b2e3afdafc54b32a4a827496f7` |
| `monster_sheet.png`（1170×1300） | `4b09429be5e51bcb1a3b237318315e0d4f5be6ce2c708ffd4f48a2e6c87bdde6` |
| `restored_128_129.png`（800×340） | `0a60a2747ac7d073b122c8d5209a41e0eaf13b5e78341d28c2ebcf19f3c58cb0` |

檢視結果：兩者均為透明背景、整數倍 nearest 像素、同一套 16 色 palette 與相同 SHP
plane 封裝；局部整理後仍維持 96×88／128×96 與原 silhouette，且完整圖鑑與 128／129
聯絡圖已重新產生。這是「風格接近原生」的 `strong` 工具性推論，不是原始美術 bytes 的
`confirmed` 證據；若日後取得原版 128／129 素材，仍應以新證據重新比對。

Docker 內以風格整理前的保留基線做遮罩對拍：128 為 `8448/8448` opaque mask 相同、
263 個像素只改 palette／群組；129 為 `12288/12288` mask 相同、105 個像素只改群組。
兩者邏輯尺寸、透明像素數與 SHP plane 順序均未變；工具第二次重播的全部八個產物 SHA-256
完全相同，確認產生流程冪等。現行 Go package 的 `TestRestoredSprites` 已在
`DQ3_ASSETS=dq3_remake_ebitan/mobile/assets` 下通過，並以
逐 byte 對拍確認 `internal/dq3data/restored_sprites.go` 的 128／129 內容等於本機 SHP
回補資料；`mobile/assets/DQ3MNS.SHP` 也與工作 SHP 完全相同。128／129 的 mask 分別以
逐列 run-length 解碼為 2695／4109 個 opaque pixels，Go 可走 SHP direct decode，不必回退；
C fallback 亦以 `gcc -fsyntax-only -Wall -Wextra -Werror` 通過。

## 發佈界線

公開 patch 包不含遊戲檔，也不含 `DQ3MNS_fixed.SHP`。本機完整版包會在複製使用者合法持有
的 `assets_raw/` 後，僅以回補 SHP 替換同名檔，再做雜湊與啟動檢查；原始目錄與工作檔不會
被覆寫。若未提供合法素材，完整版建置應失敗即關閉（fail-closed），不能用空檔或合理預設
代替。
