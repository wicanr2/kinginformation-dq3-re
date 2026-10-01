"""僅在 Docker 內使用；建立明示工具修正的 dosgolem 原版收據。"""
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def run(args, cwd, log, expected=0):
    with log.open('w') as stream:
        result = subprocess.run(args, cwd=cwd, stdout=stream,
                                stderr=subprocess.STDOUT, timeout=100)
    if (expected == 0 and result.returncode != 0) or (expected != 0 and result.returncode == 0):
        raise RuntimeError(f'工具結果不符預期：{args}，請讀 {log}')


def main():
    source, original = Path('/dosgolem'), Path('/repo/assets_raw')
    out = Path('/repo/work/dosgolem-opening')
    if out.exists() and out.stat().st_uid != os.getuid():
        raise RuntimeError('輸出目錄擁有權錯誤')
    out.mkdir(exist_ok=True)
    # 先撤銷上次收據，失敗不能留下看似有效的舊 PASS。
    (out / 'receipt.json').unlink(missing_ok=True)
    (out / 'comparison.json').unlink(missing_ok=True)
    source_file = source / 'internal/dos/files.go'
    expected_source = '463a8a83315a3054af75d6220a8d5bcc1464021ea28bad15659e6b9d5a0ce69e'
    if digest(source_file) != expected_source:
        raise RuntimeError('dosgolem 檔名服務版本已變，需重新審查工具修正')
    exe = original / 'DQ3.EXE'
    if exe.stat().st_size != 115282 or digest(exe) != '5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c':
        raise RuntimeError('原版 EXE 與已審查版本不符')
    # 上游始終唯讀；複製明確需要的程式碼，不複製語料、原版或資料庫。
    with tempfile.TemporaryDirectory(prefix='dq3-dosgolem-') as temp:
        work = Path(temp)
        shutil.copytree(source / 'internal', work / 'internal')
        shutil.copytree(source / 'cmd/probe', work / 'cmd/probe')
        shutil.copy(source / 'go.mod', work / 'go.mod')
        shutil.copy('/repo/tools/dosgolem_filename_contract_test.go',
                    work / 'internal/dos/padded_filename_test.go')
        run(['go', 'test', '-p', '2', './internal/dos', '-run', '^TestPaddedASCIIZFilename$'],
            work, out / 'filename-before.log', expected=1)
        if '--- FAIL: TestPaddedASCIIZFilename' not in (out / 'filename-before.log').read_text():
            raise RuntimeError('修正前失敗不是預期的檔名契約，不能當成回歸證據')
        target = work / 'internal/dos/files.go'
        text = target.read_text()
        old = 'func (d *DOS) resolve(name string) string {\n\tbase := baseName(name)\n'
        if text.count(old) != 1:
            raise RuntimeError('工具修正定位不唯一')
        target.write_text(text.replace(old,
            'func (d *DOS) resolve(name string) string {\n\tbase := strings.ReplaceAll(baseName(name), " ", "")\n'))
        patched_hash = digest(target)
        probe_source = work / 'cmd/probe/main.go'
        original_probe_hash = digest(probe_source)
        marker = '\tfor m.Steps < *steps && !m.CPU.Halted && !d.Exited {\n'
        observer = '''
        // 僅讀取已自然完成的翻頁；定位由固定 EXE 的 IDA bytes/xref 審查。
        if m.CPU.Seg[cpu.CS] == 0x1319 && m.CPU.IP == 0x0445 {
            dsbase := cpu.Linear(m.CPU.Seg[cpu.DS], 0)
            letter := m.Read8(dsbase + 0x5af9)
            if letter >= 'A' && letter <= 'F' {
                animation := m.Read16(cpu.Linear(m.CPU.Seg[cpu.SS], m.CPU.R[cpu.SP])) == 0x016b
                y := m.Read16(dsbase + 0x5bdc)
                label := fmt.Sprintf("card-%c", letter)
                if animation { label = fmt.Sprintf("logo-y%03d", y) }
                path := os.Getenv("DQ3_OBSERVATION_OUT") + "/" + label + ".png"
                if err := writeEGA(path, m); err != nil { die(err) }
                fmt.Printf("DQ3_OBSERVATION %c %t %d %d %d %d %s\\n", letter, animation, y, m.Steps, m.Ticks, m.PITDivisor(), label)
            }
        }
'''
        probe_text = probe_source.read_text()
        if probe_text.count(marker) != 1:
            raise RuntimeError('唯讀翻頁觀測掛鉤定位不唯一')
        probe_source.write_text(probe_text.replace(marker, marker + observer))
        os.environ['DQ3_OBSERVATION_OUT'] = str(out)
        run(['go', 'test', '-p', '2', './internal/dos'], work, out / 'filename-after.log')
        binary = work / 'probe'
        run(['go', 'build', '-p', '2', '-o', str(binary), './cmd/probe'], work, out / 'build.log')
        # 無輸入、無狀態注入；步數在執行前固定，取各卡片已繪出的畫格。
        steps = [20000000, 100000000, 200000000, 300000000, 400000000, 500000000]
        names = ['TITA', 'TITB', 'TITC', 'TITD', 'TITE', 'TITF']
        shots = ';'.join(f'{step}:{out / (name + ".png")}' for step, name in zip(steps, names))
        args = [str(binary), '-exe', str(exe), '-root', str(original),
                '-steps', '600000001', '-trace', '12', '-dump-at', shots]
        run(args, work, out / 'original.log')
        log = (out / 'original.log').read_text()
        if '沒實作的服務（0 種）' not in log or '找不到的檔' in log:
            raise RuntimeError('原版探測有工具服務或缺檔疑點')
        opened = next(line for line in log.splitlines() if line.startswith('開過的檔（'))
        actual = [name for name in opened.split('：', 1)[1].split() if name in [n + '.P' for n in names]]
        if actual != [n + '.P' for n in names]:
            raise RuntimeError(f'原版開場讀檔順序不符：{actual}')
        frames = []
        for step, name in zip(steps, names):
            pixels, png, asset = out / (name + '.bin'), out / (name + '.png'), original / (name + '.P')
            raw = pixels.read_bytes()
            if len(raw) != 640 * 350 or max(raw) > 15 or not any(raw):
                raise RuntimeError(f'{name} 色號收據錯誤')
            if not png.read_bytes().startswith(b'\x89PNG\r\n\x1a\n'):
                raise RuntimeError(f'{name} 不是 PNG')
            frames.append({'asset': asset.name, 'asset_size': asset.stat().st_size,
                           'asset_sha256': digest(asset), 'step': step, 'width': 640, 'height': 350,
                           'pixels_file': pixels.name, 'pixels_sha256': digest(pixels),
                           'png_file': png.name, 'png_sha256': digest(png)})
        observed = []
        for line in log.splitlines():
            if not line.startswith('DQ3_OBSERVATION '):
                continue
            _, letter, animation, y, step, ticks, divisor, label = line.split()
            pixels = out / (label + '.bin')
            observed.append({'asset': 'TIT' + letter + '.P', 'animation': animation == 'true',
                             'y': int(y), 'step': int(step), 'irq0_ticks': int(ticks),
                             'pit_divisor': int(divisor), 'pixels_file': pixels.name,
                             'pixels_sha256': digest(pixels), 'png_file': label + '.png',
                             'png_sha256': digest(out / (label + '.png'))})
        logo = [f for f in observed if f['animation']]
        if [f['asset'] for f in observed if not f['animation']] != [n + '.P' for n in names]:
            raise RuntimeError('原版背景翻頁觀測不完整或重複')
        if [f['y'] for f in logo] != list(range(348, 90, -2)):
            raise RuntimeError('原版 129 次標誌翻頁尚未完整觀測')
        receipt = {'method': 'dosgolem-natural-boot-planar', 'dosgolem_revision': os.environ['DOSGOLEM_REVISION'],
                   'tool_patch': 'DOS ASCII 空白解析；只修改容器內的可丟棄副本',
                   'observer_patch': '唯讀觀測 IDA file 0x13845 的自然翻頁完成點；1319:0445，動畫返回位置1319:016b；不更改遊戲、輸入、記憶體或時鐘',
                   'source_files_sha256': expected_source, 'patched_files_sha256': patched_hash,
                   'original_probe_source_sha256': original_probe_hash,
                   'observed_probe_source_sha256': digest(probe_source),
                   'probe_sha256': digest(binary), 'original_path': 'assets_raw/DQ3.EXE',
                   'original_size': exe.stat().st_size, 'original_sha256': digest(exe),
                   'go_version': subprocess.check_output(['go', 'version'], text=True).strip(),
                   'input': [], 'state_injection': False, 'rng_comparison': False,
                   'scope': '六張開場卡片順序與原始色號；不含 RGB 淡入、時序、音訊或 campaign',
                   'original_log_sha256': digest(out / 'original.log'), 'frames': frames,
                   'observations': observed}
        (out / 'receipt.json').write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + '\n')
        print('dosgolem 原版六張開場收據已產生；尚須 remake 對拍。', flush=True)


if __name__ == '__main__':
    main()
