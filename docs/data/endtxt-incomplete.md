# ENDTXT.TXT(結局文本)在未發售版為不完整佔位

`tools` 解碼 + 遊戲內渲染交叉確認:`ENDTXT.TXT`(534 bytes,56 筆指標)在精訊未發售版**結局文字未定稿**:

- 指標表前 16 筆(rec 0–15)指向 2-byte 的 `0xffff`(空記錄)。
- 後段(rec 16+)有 glyph 資料,但 text-mode 渲染出**羅馬字母**序列(如 rec24 = `MNOPQRSTUV…`),
  非繁中劇情字 —— 像 staff credits 佔位或未填實字模。多筆以 `0x0000` 起(對話解析視為終止 → 空框)。

**結論**:這是未發售版「結局/credits 尚未完成」的一手證據,非 remake 缺陷。remake 的結局捲動
**機制**完整(run_finale → 逐段 ENDTXT → THE END,stderr 驗證 55 段全播);呈現的內容忠實反映
原始(不完整)資料。日後若取得完成版 ENDTXT 可直接替換 assets,機制不需改。
