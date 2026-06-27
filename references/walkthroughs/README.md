# 攻略原文收錄(台灣本土電玩史保存)

精訊版 DQ3 從未正式發售,沒有官方攻略;能把這個未發售中文版的流程、地名、商品、進度兜起來,
靠的是一整代台灣玩家在 BBS / 個人網站上留下的攻略。這些是 1990s–2010s **台灣本土的電玩文化史料**——
本目錄將原文收錄、**完整註明原始作者與出處**,作為歷史保存與本反組譯專案的 ground-truth 交叉佐證之用。
著作權仍歸原作者所有;若原作者不希望被收錄,請來信告知即移除。

## 收錄清單

| 檔案 | 作者 | 原始出處 | 說明 |
|---|---|---|---|
| `dusheng_dq3.txt` / `.html` | **杜勝利**(喵喵的家 / 喵喵笨兔)| `http://vvv.myweb.hinet.net`、`vv0817.neocities.org/gametxt/09_dq3` | 56 步詳細全流程攻略(2016 完成):從阿里阿罕開場到索瑪終戰、各城進度、關鍵商品。本專案 CTY→地名、商店商品、進度串接的主參照。 |
| `qingshan-kongfangxiong-bbs-1994.html` | **孔方兄**(Dollar Cheng / 陳煒元,94/07)、**青衫詩客**(小邱 / Chi'u I-Nan,1995.09)等 | 1994–95 台灣 FidoNet / 90Net / Gamenet BBS 討論串 | 條列式攻略 + 一手 bug 修改法(hex editor patch)+ 當機點/存檔修改。當年玩家摸出的 binary patch,正好對得上今天從 `DQ3.EXE` 反組譯出的同一批 bug(見 `docs/18`–`22`)。 |

## 用途

- **ground truth**:remake 接線以攻略為原版行為基準(主線里程碑、道具取得鏈、boss 觸發、過場事件)。
  逐章對照盤點見 [`../../docs/data/walkthrough-flow-audit.md`](../../docs/data/walkthrough-flow-audit.md)。
- **bug 交叉佐證**:青衫/孔方兄當年的 patch 與本專案反組譯結果互證(見 [`../../docs/history/dq3-bbs-1994.md`](../../docs/history/dq3-bbs-1994.md))。
- **致敬**:這些作者在沒有官方支援、靠社群口耳相傳的年代,把一個半成品遊戲玩到破關並留下紀錄。為這份心血留個見證。
