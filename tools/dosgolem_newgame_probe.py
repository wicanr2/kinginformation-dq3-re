"""原版冷啟動、主選單、創角與生日首頁收據；執行前明定種子；續頁／母親仍待閉合。

入口與證據分級見 docs/113-newgame-geometry-re.md、GitHub Issue #4。
只在一次性 Docker 執行；原版、上游來源唯讀，工作產物不入 Git。
"""
from pathlib import Path
import hashlib, json, os, re, shutil, subprocess, tempfile

repo = Path('/repo')
out = Path('/work/dosgolem-opening')
generation_script = Path(__file__).read_bytes()
scenario = os.environ.get('DQ3_NEWGAME_PROBE_SCENARIO', 'initial')
home_scenarios = ('mother_home_entry', 'mother_home_contract', 'mother_home_navigation', 'mother_home_animation')
mother_scenarios = ('mother_approach', 'mother_finish', 'king_approach') + home_scenarios
creation_scenarios = ('name_creation', 'opening_accept', 'birthday_continue') + mother_scenarios
assert scenario in ('initial', 'name_navigation', 'name_function_mode') + creation_scenarios, scenario
prefix = {'initial':'issue4-keylog', 'name_navigation':'issue4-name',
          'name_function_mode':'issue4-mode', 'name_creation':'issue4-creation',
          'opening_accept':'issue4-opening',
          'birthday_continue':'issue4-birthday-pages',
          'mother_approach':'issue4-mother-approach',
          'mother_finish':'issue4-mother-finish',
          'king_approach':'issue4-king-approach',
          'mother_home_entry':'issue4-home-entry',
          'mother_home_contract':'issue4-home-contract',
          'mother_home_navigation':'issue4-home-navigation',
          'mother_home_animation':'issue4-home-animation'}[scenario]
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
if scenario in ('name_function_mode',) + creation_scenarios:
    # raw0 → 上 raw36 → 左 raw35 → Enter 進功能列 → Enter 選英數。
    # raw35 才是語意 cell43；raw43 是聲調，不能混為功能格。
    keys += [(760000000, 0x48), (780000000, 0x4b), (800000000, 0x1c), (820000000, 0x1c)]
    captures += [(771000000, 'up-wrap'), (791000000, 'function-cell'),
                 (811000000, 'function-focus'), (831000000, 'alnum')]
    stop = 840000000
if scenario in creation_scenarios:
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
if scenario in ('opening_accept', 'birthday_continue') + mother_scenarios:
    # 第17次正式 Enter 接受角色；後續只觀察，不寫入出生位置或故事狀態。
    keys += [(1100000000, 0x1c)]
    captures += [(1111000000, 'accepted'), (1151000000, 'birthday-wait'),
                 (1191000000, 'stable')]
    stop = 1200000000
if scenario in ('birthday_continue',) + mother_scenarios:
    # 在已自然抵達的生日等待狀態追加兩次Enter；未知續頁不預先命名為房間。
    keys += [(1220000000, 0x1c), (1340000000, 0x1c)]
    captures += [(1231000000, 'continue-1'), (1271000000, 'continue-1-wait'),
                  (1311000000, 'continue-1-stable'), (1351000000, 'continue-2'),
                  (1391000000, 'continue-2-wait'), (1431000000, 'continue-2-stable')]
    # 已觀測的自然consumer邊界，僅擷取，不跳轉或寫入文字／場景狀態。
    captures += [(1221357922, 'birthday-tail-before-eof'),
                 (1221370959, 'birthday-scroll-1'),
                 (1221491508, 'birthday-scroll-2'),
                 (1221612130, 'birthday-scroll-3'),
                 (1221732757, 'birthday-scroll-4'),
                 (1221840427, 'birthday-scroll-complete'),
                 (1222142894, 'room-background-before-actors')]
    stop = 1440000000
if scenario in mother_scenarios:
    # 第19次正常確認後，原版必經進入選單、三次圖像選擇與結果確認。
    # sub_21F79接受預設選項；保留原始EXE內NOP比較，不修改驗證或旗標。
    modal_scans = [0x1c] * 5
    if scenario == 'mother_home_navigation':
        # Escape實際選定第一輪；第二輪四方向環繞，完成選圖及結果確認後多送一次Enter。
        modal_scans = [0x1c, 0x01, 0x4b, 0x48, 0x4d, 0x50] + [0x1c] * 4
    for i, scan in enumerate(modal_scans):
        step = 1440000000 + i*20000000
        keys.append((step, scan))
        captures.append((step+11000000, f'opening-modal-{i+1:02d}'))
    # 主角(5,5)左側是床；先下2、左2、下3、右6接近(9,10)。
    # 逐格遵循原始CTY通路，沒有狀態注入或談話快捷入口。
    directions = [0x50]*2 + [0x4b]*2 + [0x50]*3 + [0x4d]*6
    for i, scan in enumerate(directions):
        step = 1440000000 + len(modal_scans)*20000000 + i*20000000
        keys.append((step, scan))
        captures.append((step+11000000, f'approach-{i+1:02d}'))
    extra_steps = (len(modal_scans) - 5) * 20000000
    captures.append((1851000000 + extra_steps, 'after-approach'))
    stop = 1860000000 + extra_steps
if scenario in ('mother_finish', 'king_approach'):
    # 原始record80只有一個FFFC；追加一次正常Enter，觀察EOF後的自然續行。
    keys.append((1900000000, 0x1c))
    captures += [(1911000000, 'castle-confirm'), (1931000000, 'after-mother-return')]
    stop = 1940000000
if scenario == 'king_approach':
    # 保留母親返回的38次正常輸入，再沿城鎮通路向北走9格。
    # 這是待驗路線；只記錄原版結果，不注入城堡／王座狀態。
    for i in range(9):
        step = 1960000000 + i * 20000000
        keys.append((step, 0x48))
        captures.append((step + 11000000, f'castle-north-{i+1:02d}'))
    stop = 2140000000
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
# 已完整按hash歸檔後移除本情境舊產物；未重生的圖不能混入新收據。
for previous in sorted(previous_files):
    previous.unlink()
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
    if scenario == 'mother_home_animation':
        queue_cases = ''.join(f'\t\tcase {step}: m.QueueKey(0x{scan:02x}); fmt.Printf("DQ3_KEY_QUEUED step=%d scan={scan:02x}\\n", m.Steps); fmt.Printf("DQ3_INPUT_CLOCK step=%d scan={scan:02x} ticks=%d pit_divisor=%d counter0002=%d phase0004=%d\\n",m.Steps,m.Ticks,m.PITDivisor(),m.Read16(cpu.Addr(0x15ed,2)),m.Read8(cpu.Addr(0x15ed,4)))\n' for step, scan in keys)
    observed_steps = ', '.join(str(step) for step, _ in captures)
    probetext = probetext.replace(marker, '''\tm.KeyEvery = 500000\n\tvar previousKeyIRQs uint64\n'''+marker+'''\t\tswitch m.Steps {\n'''+queue_cases+'''\t\t}\n\t\tif m.KeyIRQs != previousKeyIRQs {\n\t\t\tfmt.Printf("DQ3_KEY_DELIVERED step=%d count=%d port60=%02x CSIP=%04x:%04x\\n", m.Steps, m.KeyIRQs, m.In8(0x60), m.CPU.Seg[cpu.CS], m.CPU.IP)\n\t\t\tpreviousKeyIRQs = m.KeyIRQs\n\t\t}\n'''+f'''\t\tswitch m.Steps {{\n\t\tcase {observed_steps}:\n\t\t\tfmt.Printf("DQ3_NAME_OBSERVED step=%d DS=%04x raw_cursor=%d name_mode=%04x\\n", m.Steps, m.CPU.Seg[cpu.DS], m.Read16(cpu.Addr(m.CPU.Seg[cpu.DS], 0x26fe)), m.Read16(cpu.Addr(m.CPU.Seg[cpu.DS], 0x26fc)))\n\t\t}}\n''')
    if scenario in creation_scenarios:
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
    if scenario in ('birthday_continue',) + mother_scenarios:
        # 唯讀觀測IDA已定位的caller與文字等待；raw欄位不在探測器猜命名。
        flow_hook = r"""
        flowPC := uint32(m.CPU.Seg[cpu.CS])*16+uint32(m.CPU.IP)+0xef00
        if dq3SeedApplied {
            switch flowPC {
            case 0x100ab, 0x100b5, 0x100c4, 0x100d5, 0x100d8, 0x100e9, 0x100fa,
                 0x21558, 0x21501, 0x216c3, 0x21726, 0x21a8b:
                ds := m.CPU.Seg[cpu.DS]
                fmt.Printf("DQ3_BIRTHDAY_FLOW step=%d ida_linear=%05x DS=%04x SI=%04x BP=%04x DX=%04x raw259b=%d raw0716=%04x raw0718=%04x raw4f33=%d raw4f35=%d seed=%04x raw251d=%d raw0b2d=%d raw25d1=%04x raw4f2d=%d raw0004=%d raw26ad=%d ticks=%d pit_divisor=%d\n",
                    m.Steps,flowPC,ds,m.CPU.R[cpu.SI],m.CPU.R[cpu.BP],m.CPU.R[cpu.DX],
                    m.Read8(cpu.Addr(ds,0x259b)),m.Read16(cpu.Addr(ds,0x0716)),m.Read16(cpu.Addr(ds,0x0718)),
                    m.Read16(cpu.Addr(ds,0x4f33)),m.Read16(cpu.Addr(ds,0x4f35)),m.Read16(cpu.Addr(ds,0x0b5a)),
                    m.Read16(cpu.Addr(ds,0x251d)),m.Read8(cpu.Addr(ds,0x0b2d)),m.Read16(cpu.Addr(ds,0x25d1)),m.Read16(cpu.Addr(ds,0x4f2d)),
                    m.Read8(cpu.Addr(ds,0x0004)),m.Read8(cpu.Addr(ds,0x26ad)),m.Ticks,m.PITDivisor())
            }
            if (flowPC == 0x11ee8 || flowPC == 0x1e30b) &&
                m.Read16(cpu.Addr(m.CPU.Seg[cpu.DS],0x4f33)) == 5 &&
                m.Read16(cpu.Addr(m.CPU.Seg[cpu.DS],0x4f35)) == 5 {
                ds := m.CPU.Seg[cpu.DS]
                fmt.Printf("DQ3_ROOM_SPRITE step=%d ida_linear=%05x DS=%04x BX=%04x SI=%04x DI=%04x raw0004=%d seed=%04x\n",
                    m.Steps,flowPC,ds,m.CPU.R[cpu.BX],m.CPU.R[cpu.SI],m.CPU.R[cpu.DI],
                    m.Read8(cpu.Addr(ds,0x0004)),m.Read16(cpu.Addr(ds,0x0b5a)))
            }
            if flowPC == 0x167c5 || flowPC == 0x167f6 || flowPC == 0x16822 ||
                flowPC == 0x122cd || flowPC == 0x100e9 || flowPC == 0x100fa || flowPC == 0x1010a {
                ds := m.CPU.Seg[cpu.DS]
                if flowPC == 0x167f6 { dq3NPCSequenceStep++ }
                npc := uint16(0x0b66)
                fmt.Printf("DQ3_NPC_SEQUENCE step=%d ida_linear=%05x ordinal=%d ticks=%d pit_divisor=%d DS=%04x SI=%04x DX=%04x AX=%04x CX=%04x npc0=%02x%02x%02x%02x%02x%02x%02x%02x raw3d4c=%d raw0004=%d seed=%04x player_x=%d player_y=%d\n",
                    m.Steps,flowPC,dq3NPCSequenceStep,m.Ticks,m.PITDivisor(),ds,
                    m.CPU.R[cpu.SI],m.CPU.R[cpu.DX],m.CPU.R[cpu.AX],m.CPU.R[cpu.CX],
                    m.Read8(cpu.Addr(ds,npc)),m.Read8(cpu.Addr(ds,npc+1)),
                    m.Read8(cpu.Addr(ds,npc+2)),m.Read8(cpu.Addr(ds,npc+3)),
                    m.Read8(cpu.Addr(ds,npc+4)),m.Read8(cpu.Addr(ds,npc+5)),
                    m.Read8(cpu.Addr(ds,npc+6)),m.Read8(cpu.Addr(ds,npc+7)),
                    m.Read16(cpu.Addr(ds,0x3d4c)),m.Read8(cpu.Addr(ds,0x0004)),
                    m.Read16(cpu.Addr(ds,0x0b5a)),m.Read16(cpu.Addr(ds,0x4f33)),m.Read16(cpu.Addr(ds,0x4f35)))
                name := ""
                if flowPC == 0x16822 { name = fmt.Sprintf("npc-sequence-%02d",dq3NPCSequenceStep) }
                if flowPC == 0x100e9 { name = "npc-sequence-complete" }
                if flowPC == 0x100fa { name = "mother-dialogue-eof" }
                if flowPC == 0x1010a { name = "mother-dialogue-return" }
                if name != "" && !dq3NPCCaptures[name] {
                    dq3NPCCaptures[name] = true
                    path := "/work/dosgolem-opening/__PREFIX__-"+name
                    if err := writeScreen(m,path+".png",0,0); err != nil { die(err) }
                    if err := os.WriteFile(path+".bin",m.Indexed(),0o644); err != nil { die(err) }
                }
            }
            if flowPC >= 0x216d0 && flowPC <= 0x21718 &&
                (flowPC == 0x216d0 || flowPC == 0x216f3 || flowPC == 0x21718) {
                ds := m.CPU.Seg[cpu.DS]
                location := "birthday"
                if m.Read16(cpu.Addr(ds,0x4f33)) == 5 && m.Read16(cpu.Addr(ds,0x4f35)) == 5 {
                    location = "room"
                }
                __OTHER_LOCATION__
                phase := "visible"
                if flowPC != 0x216d0 { phase = "hidden" }
                fmt.Printf("DQ3_WAIT_PHASE step=%d ida_linear=%05x location=%s phase=%s DS=%04x BP=%04x DX=%04x raw0005=%d raw0004=%d raw0b34=%d seed=%04x ticks=%d pit_divisor=%d\n",
                    m.Steps,flowPC,location,phase,ds,m.CPU.R[cpu.BP],m.CPU.R[cpu.DX],
                    m.Read16(cpu.Addr(ds,0x0005)),m.Read8(cpu.Addr(ds,0x0004)),
                    m.Read16(cpu.Addr(ds,0x0b34)),m.Read16(cpu.Addr(ds,0x0b5a)),m.Ticks,m.PITDivisor())
                name := location+"-wait-arrow-"+phase
                capture := false
                if flowPC != 0x21718 {
                    dq3PhaseSeen[name]++
                    capture = dq3PhaseSeen[name] == 1
                    // 首次生日箭頭早於可見頁換入；保留為換頁診斷，第二次才是完整畫面。
                    if location == "birthday" && phase == "visible" {
                        capture = dq3PhaseSeen[name] <= 2
                        if dq3PhaseSeen[name] == 1 { name += "-transient" }
                    }
                }
                if capture {
                    path := "/work/dosgolem-opening/__PREFIX__-"+name
                    if err := writeScreen(m,path+".png",0,0); err != nil { die(err) }
                    if err := os.WriteFile(path+".bin",m.Indexed(),0o644); err != nil { die(err) }
                }
            }
        }
"""
        if scenario == 'mother_home_animation':
            # 只觀測已定位的六tick翻轉參數及取圖consumer，不追IRQ driver細節。
            flow_hook += r"""
        if flowPC == 0x1fea4 || flowPC == 0x1fea9 || (dq3SeedApplied &&
            (flowPC == 0x11ed0 || flowPC == 0x11ee8 || flowPC == 0x1e2ff || flowPC == 0x1e30b)) {
            ds := m.CPU.Seg[cpu.DS]
            if ds != 0x15ed { panic("animation observer DS differs") }
            fmt.Printf("DQ3_NPC_ANIMATION step=%d ida_linear=%05x DS=%04x AX=%04x BX=%04x DX=%04x SI=%04x DI=%04x BP=%04x counter0002=%d phase0004=%d ticks=%d pit_divisor=%d player_x=%d player_y=%d npc0=%02x%02x%02x%02x%02x%02x%02x%02x\n",
                m.Steps,flowPC,ds,m.CPU.R[cpu.AX],m.CPU.R[cpu.BX],m.CPU.R[cpu.DX],
                m.CPU.R[cpu.SI],m.CPU.R[cpu.DI],m.CPU.R[cpu.BP],
                m.Read16(cpu.Addr(ds,2)),m.Read8(cpu.Addr(ds,4)),m.Ticks,m.PITDivisor(),
                m.Read16(cpu.Addr(ds,0x4f33)),m.Read16(cpu.Addr(ds,0x4f35)),
                m.Read8(cpu.Addr(ds,0x0b66)),m.Read8(cpu.Addr(ds,0x0b67)),
                m.Read8(cpu.Addr(ds,0x0b68)),m.Read8(cpu.Addr(ds,0x0b69)),
                m.Read8(cpu.Addr(ds,0x0b6a)),m.Read8(cpu.Addr(ds,0x0b6b)),
                m.Read8(cpu.Addr(ds,0x0b6c)),m.Read8(cpu.Addr(ds,0x0b6d)))
        }
"""
        flow_hook = flow_hook.replace('__PREFIX__', prefix)
        flow_hook = flow_hook.replace('__OTHER_LOCATION__', '''if dq3MotherEntered {
                    location = "other"
                }''' if scenario in mother_scenarios else '')
        if scenario in mother_scenarios:
            flow_hook = flow_hook.replace('if dq3SeedApplied {',
                'if dq3SeedApplied && flowPC == 0x1010b { dq3MotherEntered = true }\n        if dq3SeedApplied {',1)
            flow_hook += """
        if dq3SeedApplied && ((flowPC == 0x19530 &&
            (m.Read16(cpu.Addr(m.CPU.Seg[cpu.DS],0x4f1f)) != 0xff ||
             m.Read16(cpu.Addr(m.CPU.Seg[cpu.DS],0x4f46))&4 != 0)) || flowPC == 0x196d2 || flowPC == 0x1970b ||
            flowPC == 0x1010b || flowPC == 0x10121 || flowPC == 0x10130 || flowPC == 0x101c5 ||
            flowPC == 0x194c3 || flowPC == 0x10148 || flowPC == 0x10163 || flowPC == 0x1017e ||
            flowPC == 0x10199 || flowPC == 0x101b4 || flowPC == 0x101ce || flowPC == 0x101d3 ||
            flowPC == 0x101e6 || flowPC == 0x101ef || flowPC == 0x101f8 || flowPC == 0x1020a ||
            flowPC == 0x21ddc || flowPC == 0x21e94) {
            ds := m.CPU.Seg[cpu.DS]
            fmt.Printf("DQ3_MOTHER_ENTRY step=%d ida_linear=%05x ticks=%d pit_divisor=%d player_x=%d player_y=%d raw0b24=%04x raw0b55=%d raw258a=%04x raw258c=%04x SI=%04x AX=%04x npc0=%02x%02x%02x%02x%02x%02x%02x%02x seed=%04x flag_byte50=%02x clock=%d rawflag_byte17=%02x raw0b34=%d\\n",
                m.Steps,flowPC,m.Ticks,m.PITDivisor(),m.Read16(cpu.Addr(ds,0x4f33)),m.Read16(cpu.Addr(ds,0x4f35)),
                m.Read16(cpu.Addr(ds,0x0b24)),m.Read8(cpu.Addr(ds,0x0b55)),m.Read16(cpu.Addr(ds,0x258a)),m.Read16(cpu.Addr(ds,0x258c)),
                m.CPU.R[cpu.SI],m.CPU.R[cpu.AX],m.Read8(cpu.Addr(ds,0x0b66)),m.Read8(cpu.Addr(ds,0x0b67)),
                m.Read8(cpu.Addr(ds,0x0b68)),m.Read8(cpu.Addr(ds,0x0b69)),m.Read8(cpu.Addr(ds,0x0b6a)),m.Read8(cpu.Addr(ds,0x0b6b)),
                m.Read8(cpu.Addr(ds,0x0b6c)),m.Read8(cpu.Addr(ds,0x0b6d)),m.Read16(cpu.Addr(ds,0x0b5a)),m.Read8(cpu.Addr(ds,0x4f7a)),m.Read16(cpu.Addr(ds,0x251d)),
                m.Read8(cpu.Addr(ds,0x4f72)),m.Read8(cpu.Addr(ds,0x0b34)))
            if flowPC == 0x1010b || flowPC == 0x10130 || flowPC == 0x101c5 || flowPC == 0x1020a {
                path := fmt.Sprintf("/work/dosgolem-opening/__PREFIX__-mother-entry-%05x",flowPC)
                if err := writeScreen(m,path+".png",0,0); err != nil { die(err) }
                if err := os.WriteFile(path+".bin",m.Indexed(),0o644); err != nil { die(err) }
            }
        }
        if dq3MotherEntered && !dq3MotherReturned && (flowPC == 0x21414 || flowPC == 0x21441) {
            ds := m.CPU.Seg[cpu.DS]
            textSeg := m.Read16(cpu.Addr(ds,0x252e))
            fmt.Printf("DQ3_MOTHER_TEXT step=%d ida_linear=%05x DS=%04x DI=%04x SI=%04x text_segment=%04x text_word=%04x\\n",
                m.Steps,flowPC,ds,m.CPU.R[cpu.DI],m.CPU.R[cpu.SI],textSeg,m.Read16(cpu.Addr(textSeg,m.CPU.R[cpu.SI])))
        }
        if flowPC == 0x1020a { dq3MotherReturned = true }
"""
            flow_hook = flow_hook.replace('__PREFIX__', prefix)
        if scenario in home_scenarios:
            # 只讀取正常選單入口、圖塊 consumer 與選擇結果；原始 NOP 比較保持。
            flow_hook += r"""
        if flowPC == 0x21ddc { dq3HomeModal = true }
        if dq3HomeModal {
            switch flowPC {
            case 0x21ddc, 0x21e1f, 0x21e26, 0x21ec1, 0x21f79, 0x21fd2,
                 0x22016, 0x21f11, 0x21e3e, 0x21e43, 0x21e94, 0x1ea8c,
                 0x1f590, 0x1f779, 0x1f908:
                ds := m.CPU.Seg[cpu.DS]
                if ds != 0x15ed { panic("home modal DS differs") }
                fmt.Printf("DQ3_HOME_MODAL step=%d ida_linear=%05x DS=%04x AX=%04x BX=%04x CX=%04x DX=%04x BP=%04x SI=%04x DI=%04x raw26fe=%d raw0722=%d raw0726=%d raw4f46=%04x choices=%02x%02x%02x raw4f09=%04x raw2534=%04x\n",
                    m.Steps,flowPC,ds,m.CPU.R[cpu.AX],m.CPU.R[cpu.BX],m.CPU.R[cpu.CX],
                    m.CPU.R[cpu.DX],m.CPU.R[cpu.BP],m.CPU.R[cpu.SI],m.CPU.R[cpu.DI],
                    m.Read16(cpu.Addr(ds,0x26fe)),m.Read16(cpu.Addr(ds,0x0722)),m.Read16(cpu.Addr(ds,0x0726)),
                    m.Read16(cpu.Addr(ds,0x4f46)),m.Read8(cpu.Addr(ds,0x2606)),m.Read8(cpu.Addr(ds,0x2607)),
                    m.Read8(cpu.Addr(ds,0x2608)),m.Read16(cpu.Addr(ds,0x4f09)),m.Read16(cpu.Addr(ds,0x2534)))
                if flowPC == 0x21e26 {
                    for _, region := range []struct{start, size uint16}{{0x09f1,19},{0x4348,32},{0x2b7a,132},{0x00db,16}} {
                        fmt.Printf("DQ3_HOME_MODAL_DATA step=%d DGROUP=%04x raw=",m.Steps,region.start)
                        for i:=uint16(0);i<region.size;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(ds,region.start+i)))}
                        fmt.Println()
                    }
                }
            }
        }
        if flowPC == 0x21e94 { dq3HomeModal = false }
"""
            if scenario != 'mother_home_entry':
                # 觀測另一個自然clock初始化的generator；不固定第二個seed、不寫入原版欄位。
                flow_hook += r"""
        switch flowPC {
        case 0x16f4b, 0x16f56, 0x16f65, 0x16fce:
            ds := m.CPU.Seg[cpu.DS]
            if ds != 0x15ed { panic("picture initialization DS differs") }
            fmt.Printf("DQ3_PICTURE_INIT step=%d ida_linear=%05x DS=%04x AX=%04x bios046c=%04x challenge=%d options=",
                m.Steps,flowPC,ds,m.CPU.R[cpu.AX],m.Read16(0x046c),m.Read8(cpu.Addr(ds,0x09f1)))
            for i:=uint16(0);i<18;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(ds,0x09f2+i)))}
            fmt.Printf(" expected=%02x%02x%02x\n",m.Read8(cpu.Addr(ds,0x0745)),m.Read8(cpu.Addr(ds,0x0746)),m.Read8(cpu.Addr(ds,0x0747)))
        }
"""
        declarations = '\tvar dq3PhaseSeen = map[string]int{}\n\tvar dq3NPCSequenceStep int\n\tvar dq3NPCCaptures = map[string]bool{}\n'
        if scenario in home_scenarios:
            declarations += '\tvar dq3HomeModal bool\n'
        if scenario in mother_scenarios:
            declarations += '\tvar dq3MotherEntered, dq3MotherReturned bool\n'
        probetext = probetext.replace(marker, declarations + marker + flow_hook, 1)
    if scenario == 'king_approach':
        marker = '\tfor m.Steps < *steps && !m.CPU.Halted && !d.Exited {\n'
        observed_steps = ', '.join(str(step) for step, _ in captures if step > 1940000000)
        observation = f'''\t\tswitch m.Steps {{
        case {observed_steps}:
            ds := uint16(0x15ed)
            fmt.Printf("DQ3_KING_APPROACH step=%d actual_ds=%04x dgroup=%04x player_x=%d player_y=%d raw0b24=%04x raw0b55=%d raw4f25=%d raw4f27=%d rawflag_byte17=%02x rawflag_byte18=%02x seed=%04x\\n",
                m.Steps,m.CPU.Seg[cpu.DS],ds,m.Read16(cpu.Addr(ds,0x4f33)),m.Read16(cpu.Addr(ds,0x4f35)),
                m.Read16(cpu.Addr(ds,0x0b24)),m.Read8(cpu.Addr(ds,0x0b55)),
                int16(m.Read16(cpu.Addr(ds,0x4f25))),int16(m.Read16(cpu.Addr(ds,0x4f27))),
                m.Read8(cpu.Addr(ds,0x4f72)),m.Read8(cpu.Addr(ds,0x4f73)),m.Read16(cpu.Addr(ds,0x0b5a)))
        }}
        if m.Steps > 1960000000 {{
            pc := uint32(m.CPU.Seg[cpu.CS])*16+uint32(m.CPU.IP)+0xef00
            ordinal := int((m.Steps-1960000000)/20000000)+1
            if ordinal <= 9 && pc == 0x11994 {{
                ds := uint16(0x15ed)
                fmt.Printf("DQ3_KING_CAMERA step=%d ordinal=%d ida_linear=%05x player_x=%d player_y=%d origin_x=%d origin_y=%d\\n",m.Steps,ordinal,pc,
                    m.Read16(cpu.Addr(ds,0x4f33)),m.Read16(cpu.Addr(ds,0x4f35)),
                    int16(m.Read16(cpu.Addr(ds,0x4f25))),int16(m.Read16(cpu.Addr(ds,0x4f27))))
            }}
            if ordinal <= 9 && pc == 0x2111b && !dq3KingReady[ordinal] {{
                dq3KingReady[ordinal] = true
                ds := uint16(0x15ed)
                ss, sp := m.CPU.Seg[cpu.SS],m.CPU.R[cpu.SP]
                fmt.Printf("DQ3_KING_READY step=%d ordinal=%d ida_linear=%05x player_x=%d player_y=%d return_cs=%04x return_ip=%04x raw0b24=%04x raw0b55=%d\\n",m.Steps,ordinal,pc,
                    m.Read16(cpu.Addr(ds,0x4f33)),m.Read16(cpu.Addr(ds,0x4f35)),
                    m.Read16(cpu.Addr(ss,sp+2)),m.Read16(cpu.Addr(ss,sp)),m.Read16(cpu.Addr(ds,0x0b24)),m.Read8(cpu.Addr(ds,0x0b55)))
                path := fmt.Sprintf("/work/dosgolem-opening/{prefix}-king-ready-%02d",ordinal)
                if err := writeScreen(m,path+".png",0,0); err != nil {{ die(err) }}
                if err := os.WriteFile(path+".bin",m.Indexed(),0o644); err != nil {{ die(err) }}
            }}
        }}
'''
        assert probetext.count(marker) == 1
        probetext = probetext.replace(marker, '\tvar dq3KingReady = map[int]bool{}\n' + marker + observation, 1)
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
            'upstream_source_host_path':os.environ.get('DQ3_DOSGOLEM_SOURCE_HOST_PATH','unknown'),
            'docker_image':'dq3-ebiten-test:20260822-r1',
            'go_version':subprocess.check_output(['go','version'],text=True).strip(),
            'generation_script_sha256':hashlib.sha256(generation_script).hexdigest(),
            'build_flags':['-trimpath','-p','2'],
            'upstream_files_sha256':hashlib.sha256(raw).hexdigest(),'patched_files_sha256':hashlib.sha256(files.read_bytes()).hexdigest(),
            'original_bios_sha256':hashlib.sha256(biosraw).hexdigest(),'patched_bios_sha256':hashlib.sha256(bios.read_bytes()).hexdigest(),
            'original_vga_sha256':hashlib.sha256(vgaraw).hexdigest(),'patched_vga_sha256':hashlib.sha256(vga.read_bytes()).hexdigest(),
            'probe_source_sha256':hashlib.sha256(probe.read_bytes()).hexdigest(),
            'probe_sha256':hashlib.sha256(binary.read_bytes()).hexdigest(),'args':args,
            'scenario':scenario,'player_input':[{'queued_step':step,'scan':hex(scan)} for step,scan in keys], 'minimum_scan_interval':500000,'game_state_injection':False}
    (out / f'{prefix}-meta.json').write_text(json.dumps(meta,ensure_ascii=False,indent=2)+'\n')
    (out / f'{prefix}-generation.py').write_bytes(generation_script)
    print('原版自然探測開始',flush=True)
    with (out / f'{prefix}.log').open('w') as log:
        result = subprocess.run(args,stdout=log,stderr=subprocess.STDOUT,timeout=900 if scenario == 'king_approach' else 840 if scenario == 'mother_home_animation' else 780 if scenario in ('mother_finish', 'mother_home_navigation') else 660 if scenario in mother_scenarios else 360 if scenario == 'birthday_continue' else 180)
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
                        'name_function_mode':[0,0,0,36,35,35,35],
                        'name_creation':[0,0,0,36,35,35,35],
                        'opening_accept':[0,0,0,36,35,35,35],
                        'birthday_continue':[0,0,0,36,35,35,35],
                        'mother_approach':[0,0,0,36,35,35,35],
                        'mother_finish':[0,0,0,36,35,35,35],
                        'king_approach':[0,0,0,36,35,35,35],
                        'mother_home_entry':[0,0,0,36,35,35,35],
                        'mother_home_contract':[0,0,0,36,35,35,35],
                        'mother_home_navigation':[0,0,0,36,35,35,35],
                        'mother_home_animation':[0,0,0,36,35,35,35]}[scenario]
    observations = [re.search(r'DS=([0-9a-f]+) raw_cursor=(\d+) name_mode=([0-9a-f]+)', line)
                    for line in meta['name_observations']]
    assert len(observations) == len(captures) and all(observations)
    # 正式移動的任意抓圖點可能正處於讀取CTY的暫時DS；命名契約只套創角段。
    checked_observations = observations[:len(expected_cursors)] if scenario in mother_scenarios else observations
    assert all(match.group(1) == '15ed' for match in checked_observations)
    assert [int(match.group(2)) for match in observations[:len(expected_cursors)]] == expected_cursors
    expected_modes = [0] + [1]*(len(expected_cursors)-1)
    if scenario in ('name_function_mode',) + creation_scenarios:
        expected_modes[-2:] = [5,2]
    assert [int(match.group(3),16) for match in observations[:len(expected_modes)]] == expected_modes
    meta['observation_contract'] = {'raw_cursor_dgroup_offset':'0x26fe',
                                    'name_mode_dgroup_offset':'0x26fc',
                                    'expected_raw_cursors':expected_cursors,
                                    'expected_modes':expected_modes}
    meta['test_rng_seed_control'] = None
    if scenario in creation_scenarios:
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
    if scenario == 'mother_home_animation':
        meta['npc_animation_events'] = [line for line in lines if line.startswith('DQ3_NPC_ANIMATION ')]
        meta['input_clock_events'] = [line for line in lines if line.startswith('DQ3_INPUT_CLOCK ')]
        assert len(meta['input_clock_events']) == len(keys)
    meta['scope'] = '只到主選單及命名導航／模式選擇；沒有創角能力、出生點或母親開場 parity'
    if scenario in creation_scenarios:
        meta['scope']='固定原版測試種子後，自然創角能力的原版收據；尚未與重製對拍，母親仍未知'
    if scenario in ('opening_accept', 'birthday_continue') + mother_scenarios:
        meta['scope']='冷啟動17次正式IRQ1輸入，固定創角種子並正常接受角色後的原版畫面；出生位置與母親開場尚未對拍'
        meta['observation_contract']['post_creation_cursor_semantics'] = '創角後僅保留原始欄位值，不將命名游標解讀為場景狀態'
    if scenario in ('birthday_continue',) + mother_scenarios:
        meta['scope']='冷啟動19次正式IRQ1輸入，創角種子固定一次，再正常接受角色與兩次生日續頁；續頁與下一場景尚待重製比較'
        meta['birthday_flow_events'] = [line for line in lines if line.startswith('DQ3_BIRTHDAY_FLOW ')]
        meta['room_sprite_events'] = [line for line in lines if line.startswith('DQ3_ROOM_SPRITE ')]
        meta['wait_phase_events'] = [line for line in lines if line.startswith('DQ3_WAIT_PHASE ')]
        meta['npc_sequence_events'] = [line for line in lines if line.startswith('DQ3_NPC_SEQUENCE ')]
    if scenario in mother_scenarios:
        meta['scope']='冷啟動37次正式IRQ1輸入；正常創角、生日及原始選單進入／三次圖像選擇／結果確認後，下2／左2／下3／右6接近家中原始事件格；母親入口尚未與重製對拍'
        meta['mother_entry_events'] = [line for line in lines if line.startswith('DQ3_MOTHER_ENTRY ')]
        meta['mother_text_events'] = [line for line in lines if line.startswith('DQ3_MOTHER_TEXT ')]
    if scenario in home_scenarios:
        meta['home_modal_events'] = [line for line in lines if line.startswith('DQ3_HOME_MODAL ')]
        meta['home_modal_data'] = [line for line in lines if line.startswith('DQ3_HOME_MODAL_DATA ')]
        assert meta['home_modal_events'] and len(meta['home_modal_data']) == 4
        meta['scope'] = '原版37次正式IRQ1冷啟動家中流程、三次圖像選擇與正常走近；只讀幾何／consumer／選擇結果，不代表remake parity'
        if scenario != 'mother_home_entry':
            meta['picture_init_events'] = [line for line in lines if line.startswith('DQ3_PICTURE_INIT ')]
            assert meta['picture_init_events'], '尚未觀測圖像選擇的自然初始化'
        if scenario == 'mother_home_navigation':
            meta['scope'] = '原版42次正式IRQ1冷啟動家中流程，含取消與四方向邊界；不代表remake parity'
    if scenario in ('mother_finish', 'king_approach'):
        meta['scope']='冷啟動38次正式IRQ1輸入；保留母親接近路徑，再一次Enter解除城門record80內嵌等待；後續移動與旗標尚未與重製對拍'
        text = exe.parent / 'D3TXT01.TXT'
        meta['original_text']={'path':str(text),'size':text.stat().st_size,'sha256':hashlib.sha256(text.read_bytes()).hexdigest()}
    if scenario == 'king_approach':
        meta['scope']='保留原版冷啟動38次母親返回輸入，再以9次正常上鍵走向城堡；只讀原版結果，尚未與重製對拍'
        meta['king_approach_events']=[line for line in lines if line.startswith('DQ3_KING_APPROACH ')]
        meta['king_ready_events']=[line for line in lines if line.startswith('DQ3_KING_READY ')]
        meta['king_camera_events']=[line for line in lines if line.startswith('DQ3_KING_CAMERA ')]
        assert len(meta['king_approach_events']) == 9
    meta['artifacts'] = []
    artifact_names = [f'{prefix}-{name}.{suffix}' for _,name in captures for suffix in ('png','bin')] + [f'{prefix}.log', f'{prefix}-generation.py']
    if scenario in ('birthday_continue',) + mother_scenarios:
        artifact_names += [f'{prefix}-{location}-wait-arrow-{phase}.{suffix}'
                           for location in ('birthday', 'room') for phase in ('visible', 'hidden')
                           for suffix in ('png', 'bin')]
        artifact_names += [f'{prefix}-birthday-wait-arrow-visible-transient.{suffix}'
                           for suffix in ('png', 'bin')]
        artifact_names += [f'{prefix}-{name}.{suffix}'
                           for name in [f'npc-sequence-{i:02d}' for i in range(1,17)] +
                                       ['npc-sequence-complete', 'mother-dialogue-eof', 'mother-dialogue-return']
                           for suffix in ('png', 'bin')]
    if scenario in mother_scenarios:
        # 只登錄本次重生的額外圖像；與固定清單去重，不把manifest列數當檔案數。
        artifact_names += [p.name for pattern in (f'{prefix}-*.png', f'{prefix}-*.bin')
                           for p in sorted(out.glob(pattern)) if p.name not in artifact_names]
    assert len(artifact_names) == len(set(artifact_names)), '收據不得重複列出產物'
    for name in artifact_names:
        artifact = out / name
        assert artifact.is_file() and artifact.stat().st_size > 0
        assert artifact.stat().st_uid == os.getuid()
        if artifact.suffix == '.png':
            assert str(artifact) in '\n'.join(lines), '不是本次重生的圖像：'+name
        meta['artifacts'].append({'path':name,'size':artifact.stat().st_size,'sha256':hashlib.sha256(artifact.read_bytes()).hexdigest()})
    (out / f'{prefix}-receipt.json').write_text(json.dumps(meta,ensure_ascii=False,indent=2)+'\n')
    print('\n'.join(lines[:30]))
    for i,line in enumerate(lines):
        if any(s in line for s in ('沒實作的服務','按鍵去向','硬體鍵盤','鍵盤輸入','停止原因','開過的檔')):
            print('\n'.join(lines[i:i+8]))
    print('\n'.join(lines[-24:]))
