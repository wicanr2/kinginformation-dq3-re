# dos.zczc.cz 遊戲下載技巧 (js-dos bundle)

`dos.zczc.cz`(同源站 `dos.lol`)的遊戲頁面**沒有直接下載連結**。它是 SPA + js-dos 線上模擬器,遊戲檔藏在後端 API。以下是把任一款遊戲的 bundle zip 抓出來的方法。

## 架構

| 角色 | 主機 |
|---|---|
| 前端頁面 (SPA) | `dos.zczc.cz/games/<中文名>/` |
| 遊戲清單 / metadata API | `api.dos.lol`(備援 `api-workers.dos.lol`) |
| 遊戲 bundle / 圖片 | `binary.dos.lol` / `images.dos.lol` |

頁面骨架只有標題,遊戲資料由 JS 向 `api.dos.lol` 取得後,再去 `binary.dos.lol` 串流。

## 下載步驟

### 1. 用遊戲 slug 查 metadata

slug = 頁面網址裡的遊戲名(**簡體中文**,需 URL-encode)。直查單款:

```bash
# 例:勇者斗恶龙3
curl -sL "https://api.dos.lol/v1/games/%E5%8B%87%E8%80%85%E6%96%97%E6%81%B6%E9%BE%993" \
  | python3 -m json.tool
```

不確定 slug 時可用搜尋:

```bash
curl -sL "https://api.dos.lol/v1/games?search=<URL-encoded 關鍵字>&pageSize=10&pageNumber=1"
```

### 2. 取 `binary_jsdos` 欄位

回傳 JSON 的關鍵欄位:

| 欄位 | 意義 |
|---|---|
| `identifier` | 簡體 slug |
| `name` | `zh-Hans` / `zh-Hant` / `en` 三語名 |
| `executable` | DOS 啟動指令(例 `dq3`) |
| `image` | 封面圖檔名(放 `images.dos.lol`) |
| **`binary_jsdos`** | **bundle zip 的檔名雜湊** ← 下載目標 |

### 3. 組出真正的下載連結

```
https://binary.dos.lol/<binary_jsdos 的值>
```

```bash
# 例:勇者斗恶龙3
curl -L "https://binary.dos.lol/55f91582f5d5459273d61e52ceea12a2c60f8811f55da29f668a4b6790e58546.zip" \
  -o dq3.zip
```

下載前可先 `curl -sIL <url>` 確認 `content-type: application/zip` 與 `content-length`。

## bundle 內容

zip = 整個 DOS 遊戲目錄打包(js-dos bundle 格式),含執行檔、資源檔(`.DAT` / `.SCR` / `.PIC` / `.FON` 等)。要本機跑:用 DOSBox 解壓後執行 `executable` 欄位的指令,或直接餵給 js-dos 載入。

## 實例:勇者鬥惡龍3 (Dragon Quest III)

- 頁面:`https://dos.zczc.cz/games/勇者斗恶龙3/`
- slug:`勇者斗恶龙3`
- `executable`:`dq3`
- `binary_jsdos`:`55f91582f5d5459273d61e52ceea12a2c60f8811f55da29f668a4b6790e58546.zip`
- 下載:`https://binary.dos.lol/55f91582f5d5459273d61e52ceea12a2c60f8811f55da29f668a4b6790e58546.zip`(2,400,450 bytes,精訊 1993 中文版)

## 一鍵腳本

```bash
#!/usr/bin/env bash
# 用法: ./fetch.sh "勇者斗恶龙3" dq3.zip
slug_enc=$(python3 -c "import urllib.parse,sys;print(urllib.parse.quote(sys.argv[1]))" "$1")
bin=$(curl -sL "https://api.dos.lol/v1/games/${slug_enc}" \
      | python3 -c "import json,sys;print(json.load(sys.stdin)['binary_jsdos'])")
curl -L "https://binary.dos.lol/${bin}" -o "${2:-game.zip}"
```
