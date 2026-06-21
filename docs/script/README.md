# 解碼後的遊戲腳本文案（render 圖）

精訊版 DQ3 的對話／劇情腳本 `D3TXT00~09.TXT` 解碼後 render 出的可讀畫面。
每張 `txtNN_full.png` 是該腳本檔的完整內容，逐筆記錄、逐碼取 `D3TXT00.FON` 字模拼成。

解碼方式見 [`../03-text-format.md`](../03-text-format.md)；可追溯的結構化資料（glyph index 序列）見 [`../data/`](../data/)。

| 檔案 | 內容 |
|---|---|
| `txt00_full.png` | 道具／武器／防具名、道具說明、系統訊息 |
| `txt01_full.png` ~ `txt07_full.png` | 主線各城鎮 NPC 對話（阿里阿罕、羅馬利亞、諾阿尼魯、波魯多加、達瑪轉職神殿、商人建城支線…）|
| `txt08_full.png` ~ `txt09_full.png` | 結局 + 下層 DQ1 世界（愛列夫加特／拉達多姆、精靈魯比斯、洛特封號）|
| `atlas_00-8f.png` | `D3TXT00.FON` 前 0x90 個字模對照表（碼 → 字）|

> 目前為 render 圖；待建立 glyph index → Unicode 全表後，可進一步 dump 成純文字（UTF-8）。
> 文案動態插值處（主角名／道具／金額）在原始資料中為控制碼占位，render 時呈空白缺口。
