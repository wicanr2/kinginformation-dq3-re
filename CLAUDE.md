# 精訊版本的 DQ3 反組譯

- 目的還原當年未發售的經典從 exe 檔透過 RE 變成 **Pascal** 的原始碼, 並且把遊戲素材都拆解出來
  - 註:`DQ3.EXE` 檔頭為 `MZP` = Borland Pascal 7.0 DPMI 保護模式執行檔, 原始語言即 Pascal, 故 RE 目標訂為還原 Pascal 原始碼 (可同語言重編譯比對, 比翻成 C 正確)。

驗證方式 RE的版本可以重新編譯執行, 與原版一模一樣 使用dosbox驗證

dosbox 可以用 docker 的也可以用 host 的只要不污染host環境都ok

re 工具需要在 docker 內執行

希望解析出流程與素材, 還有挖掘在中文版未發售的 dq3 中的相關技術

# 其他資訊

青衫先生整理的攻略 @./reference

# 專案的 github repo

https://github.com/wicanr2/kinginformation-dq3-re.git (紀錄整個反組譯）
