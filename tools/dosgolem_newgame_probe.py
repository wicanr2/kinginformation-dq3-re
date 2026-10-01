"""原版冷啟動、主選單與命名操作收據；創角測試在執行前明定種子；母親開場仍待閉合。

入口與證據分級見 docs/113-newgame-geometry-re.md、GitHub Issue #4。
只在一次性 Docker 執行；原版、上游來源唯讀，工作產物不入 Git。
"""
from pathlib import Path
import hashlib, json, os, re, shutil, subprocess, tempfile

repo = Path('/repo')
out = Path('/work/dosgolem-opening')
scenario = os.environ.get('DQ3_NEWGAME_PROBE_SCENARIO', 'initial')
assert scenario in ('initial', 'name_navigation', 'name_function_mode', 'name_creation'), scenario
prefix = {'initial':'issue4-keylog', 'name_navigation':'issue4-name',
          'name_function_mode':'issue4-mode', 'name_creation':'issue4-creation'}[scenario]
keys = [(710000000, 0x1c), (731000000, 0x1c)]
captures = [(729000000, 'menu'), (740000000, 'create'),
            (750000000, 'final' if scenario == 'initial' else 'initial')]
stop = 750000000
if scenario == 'name_navigation':
    # 正式 IRQ1 四方向邊界，再左、左、Enter、Enter；不改遊戲記憶體。
    keys += [(760000000, 0x4b), (780000000, 0x4d), (800000000, 0x48),
             (820000000, 0x50), (840000000, 0x4b), (860000000, 0x4b),
             (880000000, 0x1c), (900000000, 0x1c)]
    captures += [(771000000, 'left-wrap'), (791000000, 'right-wrap'),
                 (811000000, 'up-wrap'), (831000000, 'down-wrap'),
                 (851000000, 'left-again'), (871000000, 'raw43'),
                 (891000000, 'candidate-empty'), (911000000, 'candidate-dismissed')]
    stop = 920000000
if scenario in ('name_function_mode', 'name_creation'):
    # raw0 → 上 raw36 → 左 raw35 → Enter 進功能列 → Enter 選英數。
    # raw35 才是語意 cell43；raw43 是聲調，不能混為功能格。
    keys += [(760000000, 0x48), (780000000, 0x4b), (800000000, 0x1c), (820000000, 0x1c)]
    captures += [(771000000, 'up-wrap'), (791000000, 'function-cell'),
                 (811000000, 'function-focus'), (831000000, 'alnum')]
    stop = 840000000
if scenario == 'name_creation':
    # raw35 → 右raw36 → 下raw0 → Enter輸入0；上、左、Enter進功能列，
    # 上鍵環繞至完成，Enter進性別，最後Enter選預設男性。seed在執行前定義。
    keys += [(840000000,0x4d),(860000000,0x50),(880000000,0x1c),
             (900000000,0x48),(920000000,0x4b),(940000000,0x1c),
             (960000000,0x48),(980000000,0x1c),(1020000000,0x1c)]
    captures += [(851000000,'alnum-right'),(871000000,'alnum-zero'),
                 (891000000,'name-zero'),(911000000,'finish-up'),
                 (931000000,'finish-cell'),(951000000,'finish-focus'),
                 (971000000,'finish-choice'),(991000000,'gender'),
                 (1031000000,'ability')]
    # 能力面板 sub_1834E→sub_2111B 等待輸入，然後才返回創角確認。
    keys += [(1060000000,0x1c)]
    captures += [(1051000000,'ability-waiting'),(1071000000,'ability-confirm')]
    stop = 1080000000
assert out.is_dir() and out.stat().st_uid == os.getuid()
exe = repo / 'assets_raw/DQ3.EXE'
assert len(exe.read_bytes()) == 115282
assert hashlib.sha256(exe.read_bytes()).hexdigest() == '5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c'
for p in out.glob('issue4-*'):
    assert p.stat().st_uid == os.getuid(), str(p)
# 重生前以內容 hash 保留上一批，避免下輪覆寫唯一的正式收據與圖像。
# 沿用既有工作目錄，不另建交付或研究目錄。
previous_files = set(out.glob(prefix + '-*'))
if (out / (prefix + '.log')).exists():
    previous_files.add(out / (prefix + '.log'))
for previous in sorted(previous_files):
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
    queue_cases = ''.join(f'\t\tcase {step}: m.QueueKey(0x{scan:02x}); fmt.Printf("DQ3_KEY_QUEUED step=%d scan={scan:02x}\\n", m.Steps)\n' for step, scan in keys)
    observed_steps = ', '.join(str(step) for step, _ in captures)
    probetext = probetext.replace(marker, '''\tm.KeyEvery = 500000\n\tvar previousKeyIRQs uint64\n'''+marker+'''\t\tswitch m.Steps {\n'''+queue_cases+'''\t\t}\n\t\tif m.KeyIRQs != previousKeyIRQs {\n\t\t\tfmt.Printf("DQ3_KEY_DELIVERED step=%d count=%d port60=%02x CSIP=%04x:%04x\\n", m.Steps, m.KeyIRQs, m.In8(0x60), m.CPU.Seg[cpu.CS], m.CPU.IP)\n\t\t\tpreviousKeyIRQs = m.KeyIRQs\n\t\t}\n'''+f'''\t\tswitch m.Steps {{\n\t\tcase {observed_steps}:\n\t\t\tfmt.Printf("DQ3_NAME_OBSERVED step=%d DS=%04x raw_cursor=%d name_mode=%04x\\n", m.Steps, m.CPU.Seg[cpu.DS], m.Read16(cpu.Addr(m.CPU.Seg[cpu.DS], 0x26fe)), m.Read16(cpu.Addr(m.CPU.Seg[cpu.DS], 0x26fc)))\n\t\t}}\n''')
    if scenario == 'name_creation':
        seed_hook = r"""
        // 原版由正常玩家輸入抵達此處；只在已審查的Lv1交易套用預先固定種子。
        // IDA linear → runtime physical = linear−0xEF00；DGROUP DS=15ED。
        pc := uint32(m.CPU.Seg[cpu.CS])*16+uint32(m.CPU.IP)
        if pc == 0x1d9cc-0xef00 && !dq3SeedApplied {
            ds, si := m.CPU.Seg[cpu.DS], m.CPU.R[cpu.SI]
            ret := m.Read16(cpu.Addr(m.CPU.Seg[cpu.SS],m.CPU.R[cpu.SP]))
            if ds != 0x15ed || si != 0x507f || ret != 0x08cf ||
                m.Read8(cpu.Addr(ds,si+1)) != 0 || m.Read8(cpu.Addr(ds,si+0x15)) != 1 {
                panic("創角種子入口前提不符，停止原版收據")
            }
            before := m.Read16(cpu.Addr(ds,0x0b5a))
            m.Write16(cpu.Addr(ds,0x0b5a),0x1357)
            dq3SeedApplied, dq3AbilityRNG = true, true
            fmt.Printf("DQ3_CREATION_SEED step=%d DS=%04x SI=%04x return=%04x previous=%04x fixed=%04x\n",m.Steps,ds,si,ret,before,m.Read16(cpu.Addr(ds,0x0b5a)))
        }
        if dq3AbilityRNG && pc == 0x1e6e7-0xef00 {
            fmt.Printf("DQ3_ABILITY_RNG_ENTER step=%d delta=%d seed=%04x\n",m.Steps,m.CPU.R[cpu.AX],m.Read16(cpu.Addr(m.CPU.Seg[cpu.DS],0x0b5a)))
        }
        if dq3AbilityRNG && pc == 0x1e710-0xef00 {
            fmt.Printf("DQ3_ABILITY_RNG_RESULT step=%d delta=%d value=%d seed=%04x\n",m.Steps,m.CPU.R[cpu.BX],m.CPU.R[cpu.AX],m.Read16(cpu.Addr(m.CPU.Seg[cpu.DS],0x0b5a)))
        }
        if dq3AbilityRNG && pc == 0x108cf-0xef00 {
            dq3AbilityRNG=false
        }
        if dq3SeedApplied && pc == 0x108de-0xef00 {
            ds := m.CPU.Seg[cpu.DS]
            base := uint16(0x507f)
            fmt.Printf("DQ3_CREATION_RESULT step=%d DS=%04x base=%04x class_raw=%d gender_raw=%d level=%d current_hp=%d current_mp=%d str=%d vit=%d agi=%d max_hp=%d max_mp=%d int=%d luck=%d seed=%04x\n",m.Steps,ds,base,
                m.Read8(cpu.Addr(ds,base+1)),m.Read8(cpu.Addr(ds,base+2)),m.Read8(cpu.Addr(ds,base+0x15)),
                m.Read16(cpu.Addr(ds,base+0x16)),m.Read16(cpu.Addr(ds,base+0x18)),
                m.Read16(cpu.Addr(ds,base+0x1a)),m.Read16(cpu.Addr(ds,base+0x1e)),m.Read16(cpu.Addr(ds,base+0x24)),
                m.Read16(cpu.Addr(ds,base+0x2a)),m.Read16(cpu.Addr(ds,base+0x2c)),m.Read16(cpu.Addr(ds,base+0x26)),
                m.Read16(cpu.Addr(ds,base+0x28)),m.Read16(cpu.Addr(ds,0x0b5a)))
        }
        if dq3SeedApplied && (pc == 0x108eb-0xef00 || pc == 0x108ee-0xef00 || pc == 0x108f5-0xef00) {
            fmt.Printf("DQ3_CREATION_FLOW step=%d ida_linear=%05x\n",m.Steps,pc+0xef00)
        }
"""
        marker = '\tfor m.Steps < *steps && !m.CPU.Halted && !d.Exited {\n'
        assert probetext.count(marker)==1
        probetext=probetext.replace(marker,'\tvar dq3SeedApplied, dq3AbilityRNG bool\n'+marker+seed_hook,1)
    probe.write_text(probetext)
    subprocess.run(['gofmt','-w',str(probe)],cwd=src,check=True)
    binary = out / (prefix.replace('issue4-', 'issue4-probe-'))
    subprocess.run(['go','build','-trimpath','-p','2','-o',str(binary),'./cmd/probe'], cwd=src, check=True)
    args = [str(binary),'-exe',str(exe),'-root',str(exe.parent),'-steps',str(stop+1),'-trace','16','-log-calls',
            '-dump-at',';'.join(f'{step}:{out}/{prefix}-{name}.png' for step, name in captures),
            '-save-state',f'{stop}:{out}/{prefix}-final.state']
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
            'probe_sha256':hashlib.sha256(binary.read_bytes()).hexdigest(),'args':args,
            'scenario':scenario,'player_input':[{'queued_step':step,'scan':hex(scan)} for step,scan in keys], 'minimum_scan_interval':500000,'game_state_injection':False}
    (out / f'{prefix}-meta.json').write_text(json.dumps(meta,ensure_ascii=False,indent=2)+'\n')
    print('原版自然探測開始',flush=True)
    with (out / f'{prefix}.log').open('w') as log:
        result = subprocess.run(args,stdout=log,stderr=subprocess.STDOUT,timeout=180)
    print('原版探測結束',result.returncode,flush=True)
    lines = (out / f'{prefix}.log').read_text().splitlines()
    assert result.returncode == 0
    assert '沒實作的服務（0 種）：' in lines
    assert not any('找不到的檔（' in line for line in lines)
    key_events = [line for line in lines if line.startswith('DQ3_KEY_DELIVERED ')]
    expected_steps = [event for step, _ in keys for event in (step+1, step+500001)]
    assert len(key_events) == len(expected_steps), key_events
    for line, step in zip(key_events, expected_steps):
        if scenario == 'initial':
            assert f'step={step} ' in line, line
    # 後續 IRQ1 要等 IF 放行；契約是完整 make/break 在下一張收據前送達，
    # 不是每次都恰好 queued_step+1。記錄實際步數，不猜 ISR wall-clock。
    for i,(queued,scan) in enumerate(keys):
        make,release = key_events[2*i:2*i+2]
        make_step = int(re.search(r'step=(\d+)', make).group(1))
        release_step = int(re.search(r'step=(\d+)', release).group(1))
        deadline = min(step for step,_ in captures if step > queued)
        assert queued < make_step < release_step < deadline, (make,release)
        assert release_step - make_step >= 500000
        assert f'port60={scan:02x} ' in make and f'port60={scan|0x80:02x} ' in release
    meta['actual_irq1_events'] = key_events
    meta['name_observations'] = [line for line in lines if line.startswith('DQ3_NAME_OBSERVED ')]
    expected_cursors = {'initial':[0,0,0], 'name_navigation':[0,0,0,44,0,36,0,44,43,43,0],
                        'name_function_mode':[0,0,0,36,35,35,35], 'name_creation':[0,0,0,36,35,35,35]}[scenario]
    observations = [re.search(r'DS=([0-9a-f]+) raw_cursor=(\d+) name_mode=([0-9a-f]+)', line)
                    for line in meta['name_observations']]
    assert len(observations) == len(captures) and all(observations)
    assert all(match.group(1) == '15ed' for match in observations)
    assert [int(match.group(2)) for match in observations[:len(expected_cursors)]] == expected_cursors
    expected_modes = [0] + [1]*(len(expected_cursors)-1)
    if scenario in ('name_function_mode', 'name_creation'):
        expected_modes[-2:] = [5,2]
    assert [int(match.group(3),16) for match in observations[:len(expected_modes)]] == expected_modes
    meta['observation_contract'] = {'raw_cursor_dgroup_offset':'0x26fe',
                                    'name_mode_dgroup_offset':'0x26fc',
                                    'expected_raw_cursors':expected_cursors,
                                    'expected_modes':expected_modes}
    meta['test_rng_seed_control'] = None
    if scenario == 'name_creation':
        seed_events=[line for line in lines if line.startswith('DQ3_CREATION_SEED ')]
        results=[line for line in lines if line.startswith('DQ3_CREATION_RESULT ')]
        assert len(seed_events)==1 and 'fixed=1357' in seed_events[0], seed_events
        assert len(results)==1, results
        meta['test_rng_seed_control']={'seed':'0x1357','configured_before_execution':True,
            'applied_once_at':'natural IDA linear 0x1D9CC, caller return logical08CF',
            'field':'DGROUP0B5A, actual DS15ED','event':seed_events[0],
            'only_rng_state_modified':True,'other_gameplay_state_injection':False}
        meta['creation_results']=results
        meta['ability_random_events']=[line for line in lines if line.startswith('DQ3_ABILITY_RNG_')]
        meta['creation_flow_events']=[line for line in lines if line.startswith('DQ3_CREATION_FLOW ')]
    meta['rng_comparison'] = False
    meta['scope'] = '只到主選單及命名導航／模式選擇；沒有創角能力、出生點或母親開場 parity'
    if scenario == 'name_creation':
        meta['scope']='固定原版測試種子後，自然創角能力的原版收據；尚未與重製對拍，母親仍未知'
    meta['artifacts'] = []
    artifact_names = [f'{prefix}-{name}.{suffix}' for _,name in captures for suffix in ('png','bin')] + [f'{prefix}.log']
    for name in artifact_names:
        artifact = out / name
        assert artifact.is_file() and artifact.stat().st_size > 0
        assert artifact.stat().st_uid == os.getuid()
        meta['artifacts'].append({'path':name,'size':artifact.stat().st_size,'sha256':hashlib.sha256(artifact.read_bytes()).hexdigest()})
    (out / f'{prefix}-receipt.json').write_text(json.dumps(meta,ensure_ascii=False,indent=2)+'\n')
    print('\n'.join(lines[:30]))
    for i,line in enumerate(lines):
        if any(s in line for s in ('沒實作的服務','按鍵去向','硬體鍵盤','鍵盤輸入','停止原因','開過的檔')):
            print('\n'.join(lines[i:i+8]))
    print('\n'.join(lines[-24:]))
