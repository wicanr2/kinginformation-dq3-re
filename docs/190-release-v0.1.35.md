# DQ3 Go／Ebitengine v0.1.35 發行紀錄

日期：2026-08-24。source commit：`b26bb4864afb77dab1ae2b7246a6a840d829a5a6`；schema
`0.1.51`；content `0.1.57`。

GitHub release：<https://github.com/wicanr2/kinginformation-dq3-re/releases/tag/v0.1.35>

## 公開 patch

公開 release 只包含下列四個 patch，不含原版遊戲資料、OGG、MT-32 ROM 或其他私有素材：

| 平台 | 檔案 | bytes | SHA-256 |
|---|---|---:|---|
| Linux x86_64 | `dq3-remake-v0.1.35-patch-linux-amd64.AppImage` | 7,465,464 | `29160b5740d227ea04e026064450119841ddbb96f742b24af9dff5d24502089d` |
| Windows x86_64 | `dq3-remake-v0.1.35-patch-windows-x86_64.zip` | 6,636,004 | `94bfdc707cdb7c6c887b451303c23d117d00afd17de6e60e5ef49eeb7c14c729` |
| macOS Intel | `dq3-remake-v0.1.35-patch-macos-x86_64.zip` | 6,767,200 | `4b6769d6964311fd3e17d96891e354be938e2f674faa6ce801ea0d19e5a4ab02` |
| macOS Apple Silicon | `dq3-remake-v0.1.35-patch-macos-arm64.zip` | 6,300,738 | `2e57d563756aeffe9c381da060bb276ae5ea1dcc72888486507292ebd36393f3` |

Linux patch 使用旁置 `assets_raw/` 或 `DQ3_ASSETS`；Windows／macOS ZIP 內含 launcher 與說明。
macOS 只完成 osxcross 靜態架構／ZIP 驗證，未經真機驗收。

## 本機完整版

`dist-all/v0.1.35/full/` 保存相同四平台的本機完整版，包含使用者合法持有的原版資料、
修正版 `DQ3MNS.SHP`、`track_00..17.ogg` 與兩個 MT-32 ROM。這些檔案不加入 Git、不上傳
GitHub。四包 checksum 見本機 `dist-all/v0.1.35/SHA256SUMS.txt`。

## 驗收

- opening escort、王座勇者像素與 game-pack fail-closed validator：PASS。
- 三個 patch ZIP：CRC PASS，私有素材命中數 0。
- 三個 full ZIP：18 OGG、2 ROM、修正版 SHP hash 全部符合。
- Linux patch/full AppImage：實際解包通過；patch 私有素材 0，full 素材與 BUILD metadata
  通過；Docker＋Xvfb＋ALSA null 各啟動 8 秒，均以外層 timeout `124` 正常結束，未見
  `NewGame 失敗`、panic 或 fatal。
- PE 為 x86_64；Mach-O 分別為 X86_64／ARM64。
- 依使用者要求不跑完整長時間回歸；不把這次封包 smoke 當成新一輪 THE END 證據。

## 完成度與已知差異

功能層維持 campaign E3，資料／規則主要為 E2；整體視覺與聲音仍是 V1／V2，不是 V3。
本輪原版抽樣確認 rec80 背景顯著不符，rec78 雖角色關係相符，但 viewport、palette、對話框與
分頁仍不同；詳見 [`docs/189`](189-opening-sampled-parity-20260824.md)。本 release 是可玩功能
checkpoint，不宣稱逐像素、逐幀或逐聲音完全還原。

> 2026-08-25 勘誤：rec80 背景差異後來證實是 BLK 四段位平面順序反轉；現行工作樹已修正，
> 但 v0.1.35 產物仍保留該缺陷。現行 opening 另閉合 rec80→rec79，見 `docs/192`。
