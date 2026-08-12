# 127 — 日邦格八頭大蛇正式 trace 的 V2 畫面比對

> 日期：2026-08-12。狀態：玩家流程 E3 已通過；本文件只確認八頭大蛇戰鬥畫面為 near-state
> V2，**不是**同狀態 V3。所有原版素材、影片與暫存 PNG 均未加入 Git。

## 問題與範圍

舊 `docs/img/jipang_orochi_*_battle.png` 是單勇者 fixture，無法回答「正式四人隊的戰鬥呈現
是否 match 原版」。本輪只使用既有 `TestOpeningProductionInputTrace` 的正式 `InputState` 路徑，
在兩場八頭大蛇已經開始後取樣；沒有直接呼叫事件 handler、注入座標、替換隊伍或改寫 RNG。

為避免這條高效率 trace 只執行 `Update` 而凍結 renderer 的受擊閃光，取樣 hook 在寫 PNG 前
以同一 renderer 補足暫態 `flashCol` 的可見幀。這只消耗正常 Draw 會消耗的視覺計數，不改寫
命令、HP、RNG 或輸入；輸出代表玩家不按下一鍵時的 settled phase，不代表逐動作閃光時序。

## 輸入與可追溯性

| 類別 | 路徑／條件 | SHA-256／尺寸 |
|---|---|---|
| 原版流程畫格 | `dq3_real_video/frames/f000745.jpg`（第一戰勝利訊息） | `5f1eca39247b3f8fc6a4b253fd976af946ba36b5c4641f7cc162a02eb5b7dd61`／480×360 |
| 原版流程畫格 | `f000746.jpg`、`f000748.jpg`（第一戰後掉落、轉場） | `0a0749fb11748d6dd77ee6a705f163610930319e7e3c2ae6dd38a54e58bed16d`、`e0b995f9171323541ebe0889cb4b9d487bfb961ff13d839bd036ab555a6b1fa4` |
| 原版流程畫格 | `f000765.jpg`、`f000768.jpg`、`f000774.jpg`（選項、第二戰遭遇、戰後寶箱） | `8e9cd860860a02b36cf566d8c6e487fa540ebba8caca345262ca985b408d2931`、`f723ac737340e0722b83c6f079dd2138e72fdaa6cb9f86fd7dbab322999b7334`、`ef562484e1698b57a891d490cf56f3dcf9389a514c676e1f90b22c394c1cd7f1` |
| 原版靜態參考 | `dosbox/orochi_boss.png` | `746d7a95c2456fa6338438e6b61ff4fe120670304ebb0f98fdb98c36ef0cad55`／640×350 |
| remake 正式 trace | `orochi_first_command.png`、`orochi_first_message.png`、`orochi_first_end.png` | `67e4a45037abb985ca60991031ee46a8fd8b9823446d3fe334a299419dfd3e5b`、`8d25363e7773624884d621bc68b042028320d833093d63c21c45b623c5ac9519`、`d7951e097447d7b46d1ad2381f72df31199fac6c70e26af1bebe2f71accda071`／各 640×350 |
| remake 正式 trace | `orochi_second_command.png`、`orochi_second_message.png`、`orochi_second_end.png` | `5fb9627e2a0afa4c69a4d2dba92e9469d50f4d044389330703a888ff18af5351`、`197ba2b9e4814a8124b65e5fd8c287161e8a9a3d6bccca74e98a9a040f286624`、`f1b4699d04f97f89e0aa5ac3604c0ac838de840efb0a0f3dac2c671ca813eb5b`／各 640×350 |

原版流程畫格出自專案既有的同源影片：[PC 精訊版《勇者鬥惡龍3》快速通關攻略](https://www.youtube.com/watch?v=J_fozjiKTB8)。影片用於玩家可見 phase／背景 oracle；數值、flag 與 raw record 仍由原始 EXE／DAT ledger 仲裁。

Docker＋Xvfb 一次性重播 `TestOpeningProductionInputTrace` 通過，從標題以正式輸入抵達
`THE END`；最終有界 hook 重播的六張 PNG hash 與本表一致。輸出在 `/tmp` 暫存，比對後清除。

## phase 對照與結果

| phase | 原版玩家可見畫面 | remake 正式 trace | 判定 |
|---|---|---|---|
| 第一戰結束 | `f000745`：八頭大蛇被打倒後仍為洞窟背景 | `orochi_first_end`：天空／草地背景 | 不 match；背景已證實不同 |
| 第一戰後 | `f000746` 掉落、`f000748` 自動轉場與 record 70 | trace 已正常走到後續宮殿 | 流程 E3 已閉合；本項不是同一靜態 phase 對拍 |
| 第二次交談 | `f000765`：選項後 No 才開戰 | trace 已由正式 NPC／選項到第二戰 | 流程 E3 已閉合 |
| 第二戰遭遇 | `f000768`：沙漠背景、八頭大蛇出現 | `orochi_second_command`：綠色背景、四人 HUD | 不 match；背景已證實不同 |
| 第二戰後 | `f000774`：回到宮殿並可調查寶箱 | trace 已正常取得紫寶珠 | 流程 E3 已閉合 |

靜態 `dosbox/orochi_boss.png` 與 settled first-command 都有四人 HUD、命令框、敵名框及橘金色
八頭大蛇；它可排除「怪物沒有載入」和舊單人 fixture 的誤判。兩張不是同一隊伍／數值／動態
phase，且場景背景不同，直接全圖 AE 或 RMSE 不可作 V3 指標。肉眼可見的相同素材形狀只能標為
強推論（strong），不可替代同狀態 oracle。

## 原因與推論等級

| 斷言 | 等級 | 證據 |
|---|---|---|
| 第一、第二戰的玩家可見背景分別是洞窟、沙漠 | 已證實（confirmed） | 同源影片的 `f000745`、`f000768`，位於連續事件序列 |
| remake 正式路徑在兩戰分別顯示天空草地、綠色背景 | 已證實（confirmed） | 本輪六張正式 trace PNG |
| 原版 staged-boss raw 背景值為第一戰 35、第二戰 26 | 已證實（confirmed） | `docs/95` 的 IDA Pro 9.4 DGROUP ledger，保留 raw bytes／位址 |
| remake 將 formation byte1 的 `BackgroundRaw` 直接作為 `PACKBG.SCR` page | 已證實（confirmed） | `game/staged_boss.go` 的 `startPackFormation` 呼叫 `DecodePackBG(g.battle.scr, f.BackgroundRaw)`；本輪 trace 正好產生天空草地／綠色背景 |
| formation byte2 的 `PageRaw=5` 目前沒有 runtime consumer | 已證實（confirmed） | `game/` 對 `PageRaw` 零引用；它只在 pack schema／raw parity validation 保存 |
| raw 35／26、page5 應如何共同對應 archive page／palette／轉場細節 | 未知（unknown） | 尚未閉合原版 selector 的 archive loader／consumer；不得用影片顏色猜 JSON 值 |

因此真正缺口是 **兩個 formation raw byte 到 battle backdrop selector 的對應**，不是 SHP 解碼、
四人 HUD 或怪物透明度。最初的整片紅怪物是 test trace 未 Draw 時凍結的受擊 flash；在 normal
renderer settled 後，正式輸出顯示完整怪物色盤，不能把那個中間狀態寫成產品素材 defect。

## 下一個最小 gate（未執行）

1. 以 IDA Pro 9.4 非破壞性閉合「DGROUP bytes 35／26、5 → archive／page／palette source →
   renderer consumer」；每一段附輸入 hash、位址基準、raw bytes 與推論等級。
2. 若值達 D2／D3，將 selector 和引用放入 `dq3_cht` game pack，並讓共用 renderer 只消費
   pack 契約；缺值 fail-closed，不在 Go 寫 DQ3 特例、直接 page mapping 或 page22 fallback。
3. 重播相同正式 trace，取第一戰結束與第二戰遭遇的 settled PNG；先以影片作 V2 phase 對照。
   只有取得相同隊伍／數值／輸入／動態 phase 的原版 frame，才可重新評估 V3。
