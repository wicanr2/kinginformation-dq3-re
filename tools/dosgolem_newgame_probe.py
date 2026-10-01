"""原版冷啟動到主選單／初始命名的收據；不包含能力擲骰或母親開場。

入口與證據分級見 docs/113-newgame-geometry-re.md、GitHub Issue #4。
只在一次性 Docker 執行；原版、上游來源唯讀，工作產物不入 Git。
"""
from pathlib import Path
import hashlib, json, os, shutil, subprocess, tempfile

repo = Path('/repo')
out = Path('/work/dosgolem-opening')
assert out.is_dir() and out.stat().st_uid == os.getuid()
exe = repo / 'assets_raw/DQ3.EXE'
assert len(exe.read_bytes()) == 115282
assert hashlib.sha256(exe.read_bytes()).hexdigest() == '5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c'
for p in out.glob('issue4-*'):
    assert p.stat().st_uid == os.getuid(), str(p)
# 重生前以內容 hash 保留上一批，避免下輪覆寫唯一的正式收據與圖像。
# 沿用既有工作目錄，不另建交付或研究目錄。
for previous in sorted(out.glob('issue4-keylog-*')):
    assert previous.is_file(), str(previous)
    data = previous.read_bytes()
    archive = out / ('issue4-archive-' + hashlib.sha256(data).hexdigest() + previous.suffix)
    if archive.exists():
        assert archive.is_file() and archive.read_bytes() == data
    else:
        archive.write_bytes(data)
with tempfile.TemporaryDirectory(prefix='dq3-issue4-') as temp:
    src = Path(temp)
    for name in ('internal', 'cmd/probe'):
        shutil.copytree(Path('/dosgolem') / name, src / name)
    shutil.copyfile('/dosgolem/go.mod', src / 'go.mod')
    files = src / 'internal/dos/files.go'
    raw = files.read_bytes()
    assert hashlib.sha256(raw).hexdigest() == '463a8a83315a3054af75d6220a8d5bcc1464021ea28bad15659e6b9d5a0ce69e'
    old = 'func (d *DOS) resolve(name string) string {\n\tbase := baseName(name)\n'
    new = 'func (d *DOS) resolve(name string) string {\n\tbase := strings.ReplaceAll(baseName(name), " ", "")\n'
    text = raw.decode()
    assert text.count(old) == 1
    files.write_text(text.replace(old, new))
    shutil.copyfile(repo / 'tools/dosgolem_filename_contract_test.go', src / 'internal/dos/dq3_filename_contract_test.go')
    shutil.copyfile(repo / 'tools/dosgolem_palette_contract_test.go', src / 'internal/dos/dq3_palette_contract_test.go')
    before = subprocess.run(['go','test','-p','2','./internal/dos','-run','TestDACColorPageContract','-count=1'], cwd=src, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True)
    assert before.returncode != 0 and 'DAC=0x3b want=0x5b' in before.stdout, before.stdout
    (out / 'issue4-palette-before.log').write_text(before.stdout)
    bios = src / 'internal/dos/bios.go'
    biosraw = bios.read_bytes()
    marker = '\t\tcase 0x15: // 讀單一 DAC'
    assert bios.read_text().count(marker) == 1
    bios.write_text(bios.read_text().replace(marker, '\t\tcase 0x13: // BIOS 色盤頁面選擇，依公開平台契約\n\t\t\tif bl(c) <= 1 {\n\t\t\t\td.M.VGA.SelectDACPage(bl(c), uint8(c.R[cpu.BX] >> 8))\n\t\t\t} else {\n\t\t\t\td.note(0x10, 0x10, al(c))\n\t\t\t}\n'+marker))
    vga = src / 'internal/machine/vga.go'
    vgaraw = vga.read_bytes()
    vga.write_text(vga.read_text()+'''\n// SelectDACPage 依 BIOS AX=1013h 選擇標準色盤頁面。\nfunc (v *VGA) SelectDACPage(function, mode uint8) {\n if function == 0 {\n  v.ac[0x10] &= 0x7f\n  if mode != 0 { v.ac[0x10] |= 0x80 }\n } else {\n  if v.ac[0x10]&0x80 == 0 { mode <<= 2 }\n  v.ac[0x14] = mode & 0x0f\n }\n}\n''')
    subprocess.run(['gofmt','-w',str(bios),str(vga),str(src / 'internal/dos/dq3_palette_contract_test.go')],cwd=src,check=True)
    with (out / 'issue4-palette-after.log').open('w') as log:
        subprocess.run(['go','test','-p','2','./internal/dos','./internal/machine','-count=1'], cwd=src, stdout=log, stderr=subprocess.STDOUT, check=True)
    probe = src / 'cmd/probe/main.go'
    probetext = probe.read_text()
    marker = '\tfor m.Steps < *steps && !m.CPU.Halted && !d.Exited {\n'
    assert probetext.count(marker) == 1
    probetext = probetext.replace(marker, '''\tm.KeyEvery = 500000\n\tvar previousKeyIRQs uint64\n'''+marker+'''\t\tif m.Steps == 710000000 || m.Steps == 731000000 {\n\t\t\tfmt.Printf("DQ3_KEY_QUEUED step=%d scan=1c\\n", m.Steps)\n\t\t\tm.QueueKey(0x1c)\n\t\t}\n\t\tif m.KeyIRQs != previousKeyIRQs {\n\t\t\tfmt.Printf("DQ3_KEY_DELIVERED step=%d count=%d port60=%02x CSIP=%04x:%04x\\n", m.Steps, m.KeyIRQs, m.In8(0x60), m.CPU.Seg[cpu.CS], m.CPU.IP)\n\t\t\tpreviousKeyIRQs = m.KeyIRQs\n\t\t}\n''')
    probe.write_text(probetext)
    subprocess.run(['gofmt','-w',str(probe)],cwd=src,check=True)
    binary = out / 'issue4-probe-keylog'
    subprocess.run(['go','build','-trimpath','-p','2','-o',str(binary),'./cmd/probe'], cwd=src, check=True)
    args = [str(binary),'-exe',str(exe),'-root',str(exe.parent),'-steps','750000001','-trace','16','-log-calls',
            '-dump-at',f'729000000:{out}/issue4-keylog-menu.png;740000000:{out}/issue4-keylog-create.png;750000000:{out}/issue4-keylog-final.png',
            '-save-state',f'750000000:{out}/issue4-keylog-final.state']
    meta = {'kind':'原版自然啟動能力探測；尚未完成新遊戲對拍',
            'original_path':str(exe),'original_size':exe.stat().st_size,
            'original_sha256':hashlib.sha256(exe.read_bytes()).hexdigest(),
            'upstream_revision_observed':'2f44a68ebfc54b28fb15dd4a34510b0b04a5415d',
            'docker_image':'dq3-ebiten-test:20260822-r1',
            'go_version':subprocess.check_output(['go','version'],text=True).strip(),
            'generation_script_sha256':hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
            'build_flags':['-trimpath','-p','2'],
            'upstream_files_sha256':hashlib.sha256(raw).hexdigest(),'patched_files_sha256':hashlib.sha256(files.read_bytes()).hexdigest(),
            'original_bios_sha256':hashlib.sha256(biosraw).hexdigest(),'patched_bios_sha256':hashlib.sha256(bios.read_bytes()).hexdigest(),
            'original_vga_sha256':hashlib.sha256(vgaraw).hexdigest(),'patched_vga_sha256':hashlib.sha256(vga.read_bytes()).hexdigest(),
            'probe_source_sha256':hashlib.sha256(probe.read_bytes()).hexdigest(),
            'probe_sha256':hashlib.sha256(binary.read_bytes()).hexdigest(),'args':args,'player_input':['IRQ1 Enter queued@710000000','IRQ1 Enter queued@731000000'], 'minimum_scan_interval':500000,'game_state_injection':False}
    (out / 'issue4-keylog-meta.json').write_text(json.dumps(meta,ensure_ascii=False,indent=2)+'\n')
    print('原版自然探測開始',flush=True)
    with (out / 'issue4-keylog.log').open('w') as log:
        result = subprocess.run(args,stdout=log,stderr=subprocess.STDOUT,timeout=150)
    print('原版探測結束',result.returncode,flush=True)
    lines = (out / 'issue4-keylog.log').read_text().splitlines()
    assert result.returncode == 0
    assert '沒實作的服務（0 種）：' in lines
    assert not any('找不到的檔（' in line for line in lines)
    key_events = [line for line in lines if line.startswith('DQ3_KEY_DELIVERED ')]
    expected_steps = [710000001, 710500001, 731000001, 731500001]
    assert len(key_events) == len(expected_steps), key_events
    for line, step in zip(key_events, expected_steps):
        assert f'step={step} ' in line, line
    meta['actual_irq1_events'] = key_events
    meta['rng_comparison'] = False
    meta['scope'] = '只到主選單及初始注音命名；沒有創角能力、出生點或母親開場 parity'
    meta['artifacts'] = []
    for name in ['issue4-keylog-menu.png','issue4-keylog-menu.bin','issue4-keylog-create.png','issue4-keylog-create.bin','issue4-keylog-final.png','issue4-keylog-final.bin','issue4-keylog.log']:
        artifact = out / name
        assert artifact.is_file() and artifact.stat().st_size > 0
        assert artifact.stat().st_uid == os.getuid()
        meta['artifacts'].append({'path':name,'size':artifact.stat().st_size,'sha256':hashlib.sha256(artifact.read_bytes()).hexdigest()})
    (out / 'issue4-keylog-receipt.json').write_text(json.dumps(meta,ensure_ascii=False,indent=2)+'\n')
    print('\n'.join(lines[:30]))
    for i,line in enumerate(lines):
        if any(s in line for s in ('沒實作的服務','按鍵去向','硬體鍵盤','鍵盤輸入','停止原因','開過的檔')):
            print('\n'.join(lines[i:i+8]))
    print('\n'.join(lines[-24:]))
