# v0.1.34 三平台正式發布

> GitHub release：<https://github.com/wicanr2/kinginformation-dq3-re/releases/tag/v0.1.34>。
> tag／附件均已核對；tag 指向 `9d639d01b46ce327c476e4a470dffa57ccbd069d`。

以 checkpoint `9d639d01b46ce327c476e4a470dffa57ccbd069d` 在 Docker 內封裝。公開 patch 不含
`assets_raw`；full 僅供本機合法持有素材者驗收，不上傳 GitHub。四個 binary 已核對為 Linux
x86-64 ELF、Windows x86-64 PE32+、macOS x86_64／arm64 Mach-O；六個 ZIP 均通過 CRC、
`BUILD.txt` commit 與素材邊界檢查。macOS 只有靜態驗證，未經真機執行。

| 類型 | 平台 | bytes | SHA-256 |
|---|---|---:|---|
| patch | Linux amd64 AppImage | 7,354,872 | `0df3310db10a1ab655e51278b9df16eb36241098a2ba513b6dd144fd7d94ecd4` |
| patch | Windows x86_64 ZIP | 6,513,074 | `a1bdf952533d8e2250d8f3b1beb3c15cc03f52d9315b04af1545f490bcda38d0` |
| patch | macOS Intel ZIP | 6,650,283 | `647b85b6ef733bd44e9976a907d4c10aaa5a792b20a35041e190aa0307411e7d` |
| patch | macOS Apple Silicon ZIP | 6,189,083 | `e1014b4abc8456c5c7bb6c928a7c0fb0e60fb749b20e92109be45ebe49aac4a8` |
| full | Linux amd64 AppImage | 9,464,312 | `5ea1a7157d0b5fea84836243f96df79dfeab5acf68646442f1419e0eeb4c5645` |
| full | Windows x86_64 ZIP | 9,053,850 | `c097bea6232b45f0cff794151bd2e88cb3c7b74c86a753a6612a30e5b3b257be` |
| full | macOS Intel ZIP | 9,205,918 | `9a3ec8e99a882a59e0e30ae83331da96d365932a23166d7d7a0caae32c7efb99` |
| full | macOS Apple Silicon ZIP | 8,744,328 | `459b252534541b2325a83124705ccadce7e6d6b6de4fa3c523c65f756cd8b77e` |

Linux patch 以唯讀外部素材、full 以內嵌素材各執行 Docker＋Xvfb 8 秒，皆由 timeout 正常
終止（status 124），無 panic、fatal 或載入失敗。AppImage 封裝固定使用 repo 既有
`runtime-x86_64`，避免離線環境誤嘗試下載。
