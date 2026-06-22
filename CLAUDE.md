# 精訊版本的 DQ3 反組譯

- 目的還原當年未發售的經典從 exe 檔透過 RE 變成 **C** 的原始碼 (可重編譯), 並且把遊戲素材都拆解出來
  - 註:`DQ3.EXE` 經反組譯偵察 (docs/05-exe-recon.md) 確認為 **16-bit real-mode、large model 分段 DOS 程式**, 鏈結手寫組語低階驅動;**非 Borland Pascal、非保護模式** (先前依檔頭 `MZP` 的 'P' 判為 Pascal 保護模式有誤, 那只是 e_cblp=80)。經使用者確認, RE 目標訂為還原成可用 DOS C 編譯器重建的 **C 原始碼**, 以 dosbox 比對原版驗證。

驗證方式 RE的版本可以重新編譯執行, 與原版一模一樣 使用dosbox驗證

dosbox 可以用 docker 的也可以用 host 的只要不污染host環境都ok

re 工具需要在 docker 內執行

希望解析出流程與素材, 還有挖掘在中文版未發售的 dq3 中的相關技術

# 其他資訊

青衫先生整理的攻略 @./reference

# 專案的 github repo

https://github.com/wicanr2/kinginformation-dq3-re.git (紀錄整個反組譯）
