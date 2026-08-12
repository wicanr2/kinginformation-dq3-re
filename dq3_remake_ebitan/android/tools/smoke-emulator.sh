#!/usr/bin/env bash
set -euo pipefail

APK_PATH="${1:?usage: dq3-android-smoke APK_PATH [OUTPUT_DIR]}"
OUTPUT_DIR="${2:-/out}"
AVD_NAME="dq3_api34_x86_64"
PACKAGE_NAME="com.wicanr2.dq3"
ACTIVITY_NAME="${PACKAGE_NAME}/.MainActivity"
ADB="/opt/android-sdk/platform-tools/adb"
EMULATOR="/opt/android-sdk/emulator/emulator"
AVDMANAGER="/opt/android-sdk/cmdline-tools/latest/bin/avdmanager"

test -f "$APK_PATH"
if [[ ! -d /.android || ! -w /.android ]]; then
    printf '%s\n' '必須以可寫 tmpfs 掛載 /.android，供 emulator JWKS 與 ADB 金鑰使用' >&2
    exit 2
fi
WORK_ROOT="/tmp/dq3-android-emulator-${UID}"
mkdir -p "$OUTPUT_DIR" \
    "$WORK_ROOT/android-user" \
    "$WORK_ROOT/avd" \
    "$WORK_ROOT/tmp"
rm -f "$OUTPUT_DIR"/{activity-first.txt,activity-resumed.txt,activity-restarted.txt,after-input.png,after-restart.png,boot-status.log,crash-matches.txt,emulator.log,install.log,logcat.txt,pid.txt,restart.log,resume.log,start.log}
export ANDROID_USER_HOME="$WORK_ROOT/android-user"
export ANDROID_AVD_HOME="$WORK_ROOT/avd"
export TMPDIR="$WORK_ROOT/tmp"

ACCEL_MODE="off"
if [[ -r /dev/kvm && -w /dev/kvm ]]; then
    ACCEL_MODE="on"
fi
printf 'emulator_accel=%s\n' "$ACCEL_MODE" >"$OUTPUT_DIR/environment.txt"

cleanup() {
    "$ADB" emu kill >/dev/null 2>&1 || true
    if [[ -n "${EMULATOR_PID:-}" ]]; then
        kill "$EMULATOR_PID" >/dev/null 2>&1 || true
        wait "$EMULATOR_PID" >/dev/null 2>&1 || true
    fi
    "$ADB" kill-server >/dev/null 2>&1 || true
}
trap cleanup EXIT INT TERM

printf 'no\n' | "$AVDMANAGER" create avd \
    --force \
    --name "$AVD_NAME" \
    --package 'system-images;android-34;google_apis;x86_64'

# 縮小 headless smoke 的顯示與記憶體形狀，降低純 TCG 開機成本。
cat >>"$ANDROID_AVD_HOME/${AVD_NAME}.avd/config.ini" <<'EOF'
hw.lcd.width=640
hw.lcd.height=360
hw.lcd.density=160
hw.ramSize=1536
EOF

"$EMULATOR" \
    -avd "$AVD_NAME" \
    -accel "$ACCEL_MODE" \
    -gpu swiftshader_indirect \
    -no-audio \
    -no-window \
    -no-snapshot \
    -no-boot-anim \
    -no-metrics \
    -wipe-data \
    -memory 1536 \
    -cores 2 \
    >"$OUTPUT_DIR/emulator.log" 2>&1 &
EMULATOR_PID=$!

boot_deadline=$((SECONDS + 480))
boot_complete=""
last_status_at=0
while (( SECONDS < boot_deadline )); do
    if ! kill -0 "$EMULATOR_PID" 2>/dev/null; then
        printf '%s\n' 'Android emulator 在啟動完成前退出' >&2
        exit 1
    fi
    boot_complete=$("$ADB" shell getprop sys.boot_completed 2>/dev/null | tr -d '\r' || true)
    if (( SECONDS - last_status_at >= 30 )); then
        {
            printf 'seconds=%s\n' "$SECONDS"
            "$ADB" devices -l || true
            printf 'sys.boot_completed=%s\n' "$boot_complete"
            "$ADB" shell getprop init.svc.bootanim 2>/dev/null || true
            "$ADB" shell getprop ro.build.version.release 2>/dev/null || true
        } >>"$OUTPUT_DIR/boot-status.log"
        last_status_at=$SECONDS
    fi
    [[ "$boot_complete" == "1" ]] && break
    sleep 2
done
[[ "$boot_complete" == "1" ]]

# boot_completed 可能早於 package/activity/window service 可穩定回應；三者都通過後再安裝。
framework_deadline=$((SECONDS + 90))
framework_ready=""
while (( SECONDS < framework_deadline )); do
    if timeout 10s "$ADB" shell pm path android 2>/dev/null | grep -q '^package:' \
        && timeout 10s "$ADB" shell service check activity 2>/dev/null | grep -q 'found' \
        && timeout 10s "$ADB" shell service check window 2>/dev/null | grep -q 'found'; then
        framework_ready="1"
        break
    fi
    sleep 3
done
[[ "$framework_ready" == "1" ]]
sleep 15

"$ADB" shell settings put system accelerometer_rotation 0
"$ADB" shell settings put system user_rotation 1
"$ADB" install -r "$APK_PATH" | tee "$OUTPUT_DIR/install.log"
"$ADB" logcat -c
"$ADB" shell am start -W -n "$ACTIVITY_NAME" | tee "$OUTPUT_DIR/start.log"
sleep 8
"$ADB" shell dumpsys activity activities >"$OUTPUT_DIR/activity-first.txt"
"$ADB" logcat -d -v threadtime >"$OUTPUT_DIR/logcat-first.txt"
grep -E '(mResumedActivity|topResumedActivity|Resumed:|ResumedActivity:).*com\.wicanr2\.dq3' \
    "$OUTPUT_DIR/activity-first.txt"

# 不假設遊戲內部場景；只驗證 Android 觸控輸入通道可送達前景 Activity。
"$ADB" shell input tap 550 300
"$ADB" shell input swipe 90 290 180 290 500
sleep 3
"$ADB" exec-out screencap -p >"$OUTPUT_DIR/after-input.png"

# 進入 Android HOME 再回復，覆蓋 onPause/onResume。
"$ADB" shell input keyevent KEYCODE_HOME
sleep 3
"$ADB" shell am start -W -n "$ACTIVITY_NAME" | tee "$OUTPUT_DIR/resume.log"
sleep 5
"$ADB" shell dumpsys activity activities >"$OUTPUT_DIR/activity-resumed.txt"
grep -E '(mResumedActivity|topResumedActivity).*com\.wicanr2\.dq3' \
    "$OUTPUT_DIR/activity-resumed.txt"

# force-stop 後重新啟動，覆蓋全新 Activity／EbitenView 建立。
"$ADB" shell am force-stop "$PACKAGE_NAME"
"$ADB" shell am start -W -n "$ACTIVITY_NAME" | tee "$OUTPUT_DIR/restart.log"
sleep 8
"$ADB" shell dumpsys activity activities >"$OUTPUT_DIR/activity-restarted.txt"
grep -E '(mResumedActivity|topResumedActivity).*com\.wicanr2\.dq3' \
    "$OUTPUT_DIR/activity-restarted.txt"
"$ADB" exec-out screencap -p >"$OUTPUT_DIR/after-restart.png"

"$ADB" logcat -d -v threadtime >"$OUTPUT_DIR/logcat.txt"
if grep -E 'FATAL EXCEPTION|ANR in com\.wicanr2\.dq3|Process: com\.wicanr2\.dq3.*has died' \
    "$OUTPUT_DIR/logcat.txt" >"$OUTPUT_DIR/crash-matches.txt"; then
    printf '%s\n' 'Android smoke 偵測到 crash/ANR' >&2
    exit 1
fi

"$ADB" shell pidof "$PACKAGE_NAME" | tee "$OUTPUT_DIR/pid.txt"
printf '%s\n' 'Android emulator smoke 通過'
