// Package mobile 是 Android / iOS 的 ebitenmobile 綁定進入點。
//
// 綁定產物(.aar / .xcframework)由 ebitenmobile 產生:
//
//	ebitenmobile bind -target android -javapkg com.wicanr2.dq3 -o dq3.aar ./mobile
//
// 實際綁定邏輯在 mobile.go(帶 android||ios build tag);此檔(無 tag)只確保
// 套件在桌面 `go build ./...` 時仍是合法空套件(綁定碼在桌面平台被 build tag 排除)。
//
// 資產:原版素材(版權,gitignore 不入庫)於綁定前放進 mobile/assets/(見該目錄說明),
// 由 //go:embed 打包進 .aar → APK。
package mobile
