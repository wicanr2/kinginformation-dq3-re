# DQ3 Go／Ebitengine v0.1.36 發行紀錄

日期：2026-08-26。binary source commit：`d6184f426ab0cfb15d598fbd73fb03afe16ac1dd`；
schema `0.1.53`；content `0.1.59`。

GitHub release：<https://github.com/wicanr2/kinginformation-dq3-re/releases/tag/v0.1.36>

## 公開 patch

公開 release 只有下列四個 patch，不含原版遊戲資料、OGG、MT-32 ROM、IDA database 或
其他私人素材：

| 平台 | 檔案 | bytes | SHA-256 |
|---|---|---:|---|
| Linux x86_64 | `dq3-remake-v0.1.36-patch-linux-amd64.AppImage` | 4,184,568 | `3e6c6fd1c8551c3059f7ced5f31023cbe5052cda04666e78d27d96d2a624c0b0` |
| Windows x86_64 | `dq3-remake-v0.1.36-patch-windows-x86_64.zip` | 3,513,010 | `9d3ceee49d609a1b6566569064dd8313b56e44d3d5ff0e429e2d840e6fab64a9` |
| macOS Intel | `dq3-remake-v0.1.36-patch-macos-x86_64.zip` | 3,360,596 | `b1b851170138c1801edcfaddaf5f53faa14f8e4ee71e4eba0c28dcc2fe6d1005` |
| macOS Apple Silicon | `dq3-remake-v0.1.36-patch-macos-arm64.zip` | 3,127,033 | `ed81841ff76cb89bbf38f270a680273f582eaaad8a932c388ca344e5cd891535` |

## 本機完整版

`dist-all/v0.1.36/full/` 保存相同四個平台組合的本機完整版，包含目前 `assets_raw`、
修正版 `DQ3MNS.SHP`、`track_00..17.ogg` 及兩個 MT-32 ROM。四包只留本機，不加入 Git、
不附加到 GitHub Release；checksum 在同目錄 `SHA256SUMS.txt`。

刷新素材時曾發現 `assets_raw` 內的 IDA sidecar；封裝已明確排除 `.id0/.id1/.nam/.til`，
四個完整版的 RE database 命中數均為 0。現行 `DQ3MNS.SHP` SHA-256 為
`bad40e552343141b75191a5a9576adddb7a16ddf1378fe139eba4c4e15ba8bfd`。

## 驗收

- Windows 為 PE32+ x86-64；macOS 分別為 Mach-O x86_64／arm64；Linux 為 ELF x86-64。
- 六個 ZIP CRC 通過；三個 patch ZIP 私人素材為 0；三個 full ZIP 均為 18 OGG、2 ROM、
  RE database 0，且使用現行修正版 SHP。
- patch AppImage 私人素材為 0；full AppImage 的 OGG、ROM、RE database 與 SHP 同樣通過。
- Linux patch/full 在 Docker＋Xvfb＋ALSA null 各啟動 8 秒，外層 timeout `124`，log 無
  `NewGame` 失敗、panic、fatal 或 audio error。
- Windows full staging 的七個正式場景 OGG 通過 [`docs/193`](193-audio-sampling-polish-20260825.md)
  的完整解碼、格式、音量、峰值與長靜音閘門。
- macOS 僅完成 osxcross 靜態架構與 ZIP 驗證，尚未經真機驗收。

本輪依使用者要求不跑完整長時間 campaign 回歸；封包 smoke 不冒稱新的 THE END 或全畫面／
全音訊 V3 證據。

