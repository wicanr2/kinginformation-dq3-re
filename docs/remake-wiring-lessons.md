# remake 接線(wiring)經驗教訓

> 把反組譯出的事件/道具/boss 接成「照杜勝利攻略玩得通」的 remake 時,反覆踩到 / 反覆有用的模式。
> 與 `docs/00-re-methodology.md`(EXE 靜態 RE 技巧)互補:00 是「怎麼從 EXE 解出事實」,
> 本檔是「怎麼把事實接成可玩、可驗證的 remake 行為」。對應 session:2026-06-26 B 道具鏈串接(B-1~B-9)。

## 1. 驗證失敗 ≠ 資料缺失 —— 先換取法,再下結論

最常踩。一個取得鏈「驗不出來」時,第一反應不該是「原版沒這段 / 資料缺 / 要 DOSBox」,
而是「我的觸發方式錯了」。本 session 兩個實例:

- **寶箱 examine 要面對它**:歐里空金屬 CTY84 寶箱在 (23,35),直接 `warp` 把玩家放到 (23,35)
  再 examine **不觸發**;正解是站 (23,36) **朝上** examine(寶箱事件靠「面前那格是事件 tile」判定)。
  之前判「位置校正失敗」是錯的——資料一直都在,錯的是站位。
- **section table off-by-one**(前 session):自寫 parser 多加 count 前綴,整批座標靜默錯位,
  誤判 byte4=52 NPC「不存在」。跟遊戲 loader 對帳才發現。

對策:撞「驗不出來」時,窮舉合法觸發方式(四向 examine、相鄰格、各 section、加前置 flag)
再宣告缺失。呼應記憶 `feedback-dont-retreat-to-overlay-dosbox`、`re-content-exists-not-build-incomplete`。

## 2. 三種事件 gate 接法(按事件落點選)

接「達成條件才放行」的事件時,gate 接在不同落點:

| gate 種類 | 落點 | 範例 |
|---|---|---|
| **寶箱 examine gate** | chest handler 的 is_item 分支前判 flag | 甘達特金皇冠(flag 0x210 未設→空寶箱) |
| **scripted give gate** | sub2 dispatch 對特定 byte4 判 flag,未滿足→before_rec | 古布達黑胡椒(flag 0x211=達妮亞救出) |
| **runner region** | `[0x722]` 座標 hit-test / 顯式 setter | boss 區(甘達特=runner boss,無 sub2 NPC) |

辨識要點:**有 sub2 NPC 的對話事件** → scripted give gate;**寶箱/examine tile** → chest gate;
**踩進某區觸發**(boss、過場) → runner region(`[0x722]` 狀態機,docs/re-log-722)。

## 3. 前置劇情用 flag 表達,原子事件用 token 組合演示

攻略裡「先做 A 才能做 B」的前置(救人、打 boss、解詛咒),在 remake 統一用一個 flag 表達:
事件 A 完成 → 設 flag → 事件 B 的 gate 檢查該 flag。

- 古布達救人:巴哈拉達洞窟雙 boss(甘達特手下怪27 → 甘達特怪26)→ flag 0x211 → 古布達給黑胡椒。
- 在洞窟地圖 CTY 尚未定位時,該原子事件先用既有 debug token **組合**演示 + game_tester 驗 gate 行為:
  `boss:27;boss:26;flag:0x211`。地圖內真正觸發點(洞窟 NPC)待定位後補,但 gate 邏輯已可驗、已正確。

好處:gate 邏輯與「在哪觸發」解耦——邏輯先對、先可驗,地圖觸發點是獨立的後續定位工作。

## 4. game_tester 一定要在 docker 內跑(binary 在 volume)

`tools/game_tester.sh` 預設 `BIN=/build/dq3_remake`,binary 在 docker named volume `dq3build` 裡,
**host 直接 `bash tools/game_tester.sh` 會全 FAIL**(host 看不到 volume 內的 binary)。
症狀:PASS=4 FAIL=60 這種「幾乎全掛」。正解:
```
docker run --rm -v "$PWD":/repo -v dq3build:/build dq3-remake bash -lc \
  'export SDL_VIDEODRIVER=dummy SDL_AUDIODRIVER=dummy; bash /repo/tools/game_tester.sh'
```
靜態檢查類斷言(grep 原始碼)要用 `/repo/...` 路徑(docker 內),不是 host cwd。

## 5. 加 gate 會打到既有斷言 —— 同步改依賴舊行為的測試

對 byte4=25 黑胡椒加救人 gate 後,原本「直接給黑胡椒」的 game_tester 斷言(`sub2 給物 25`)立刻 FAIL,
因為現在救人前不給。改 gate **必同步更新依賴舊行為的測試**:把單條「給物」拆成「gate 前不給 + 救人後給」
兩條,既修紅又把新行為納入回歸。改行為前先 grep 測試找出所有依賴點。

## 6. 杜勝利攻略當 ground truth,座標/取法逐章查證不臆造

取法、座標、前置順序一律查攻略原文(`work/walkthrough/`),不憑記憶或經典 DQ3 慣例臆造。
攻略寫「養羊圍欄綠草地中央調查」「賣他 22500 之後就有賣王者之劍 35000」這種具體值,直接照接。
攻略當 oracle 但不照抄字串——遊戲內官方中文字串以 RE 解出的 record 為準(docs/re-ui-menus-as-text-records)。

## 7. 每步記 journal,文件對齊是獨立工序

接線是高頻小步(一條鏈一個 commit),`docs/re-log-722-state-machine.md` 逐 Step 記。
但**狀態類文件(README 計數、WORKLIST 快照、flow-audit 已接/未接表)會快速 drift**:
做完一批接線後要專門過一輪對齊,否則 README 還停在 27/27、audit 還把已接的列在「❌ 未接」。
對策:把「文件對齊」當接線批次的收尾固定工序,不要混在每條鏈裡隨手改(容易漏)。


## 8. 工具輸出可能是幻覺 —— ground truth 排序與 git 驗證

某些環境下工具回顯會夾帶幻覺內容:假的 PASS 數字、假的 commit 成功訊息、假的 grep 結果、
甚至 Read 回傳檔案裡根本沒有的行。後果嚴重——會誤以為「接好了、測過了、提交了」,實際 git log 全無。

**可靠來源排序(高→低)**:
1. `git log --oneline` / `git diff` / `git status -sb`(版控是硬事實,最可靠)
2. python 寫檔 → 再讀回數字 / 布林(`open().write` 後 `PASS=N` 這種純數字輸出)
3. `grep -c` 純數字計數
4. bash stdout 直接回顯敘述性文字(最不可靠,易被幻覺替換)

**紀律**:
- 宣告「完成 X」前,先 `git diff` 看真實改了什麼(不是憑記憶說改了什麼)。
- 驗證測試用「檔案中介 + 抓 PASS=N 數字」(`game_tester > f; grep PASS f`),不信「我看到全綠」的敘述。
- commit 後用 `git log -1` 確認真的進去了;push 後用 `git status` 確認 not ahead。
- Edit 宣稱成功但行為沒變時,用 `grep -c` 數字確認字串真的在檔案裡,別重試假動作。
- 呼應使用者提醒:記憶/前文不一定同步,用第一性原理 + ground truth(reference 攻略、ITEM.DAT、git)重新查證。
