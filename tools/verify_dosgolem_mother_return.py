"""正常39次輸入的原版南側提示收據核對；入口 docs/188。"""
import argparse,hashlib,json,re,struct,zlib
from pathlib import Path

EXPECTED_CAPTURES = {
  "before-down": {
    ".png": "26c026a0828220cf66e81e3be256e228dc11f0816c901da5941738261b6f9356",
    ".bin": "01ef25fdc7a9db2ab59bc490b939eaf3ebebb8e49dc39bbf255e4fce0e605464"
  },
  "pc-1020b": {
    ".png": "7f6ef98f6b920efd82bd3df88e684296125fc86dae07c8994b5d44744d4ad30c",
    ".bin": "bcf841f90555a614361bfc026a514a32744cf199cf9095f822ea2c328795956b"
  },
  "pc-10221": {
    ".png": "c39063647541155319f56c6eaef0a3d2538a284d52aef317b33c3fe296ec593f",
    ".bin": "df6e69900cd46af52f50827b3dad9ac03ae5897bad5835d15e44cefbc473e5ac"
  },
  "pc-1022d": {
    ".png": "11b76ab13ec9990703b7ddadf675dedb8073da430b287821ad356b3c3a42425c",
    ".bin": "2ce95a604702ea71045859fb46d73915e7ac769290097163df158d8460338b57"
  },
  "pc-10232": {
    ".png": "5478f2ee08a54999bbcea0ff652695d991aa76e60be05879ec2d18540c9d88b5",
    ".bin": "717495e4324608d3465ac25e8017670d0a4b29d48a20c038c9ea75c6aef678a7"
  },
  "pc-1023c": {
    ".png": "c39063647541155319f56c6eaef0a3d2538a284d52aef317b33c3fe296ec593f",
    ".bin": "df6e69900cd46af52f50827b3dad9ac03ae5897bad5835d15e44cefbc473e5ac"
  },
  "pc-10242": {
    ".png": "c39063647541155319f56c6eaef0a3d2538a284d52aef317b33c3fe296ec593f",
    ".bin": "df6e69900cd46af52f50827b3dad9ac03ae5897bad5835d15e44cefbc473e5ac"
  },
  "pc-10245": {
    ".png": "5901138c060e62ba103a18922010d3a658071b43c165c64b34f95d8fe50372cf",
    ".bin": "2a79c8d9da963063b6d3a912f7bb46c0dbf668d31134bc2a5d2b57982f743dac"
  },
  "pc-1024b": {
    ".png": "5901138c060e62ba103a18922010d3a658071b43c165c64b34f95d8fe50372cf",
    ".bin": "2a79c8d9da963063b6d3a912f7bb46c0dbf668d31134bc2a5d2b57982f743dac"
  },
  "after-down": {
    ".png": "0192fed7305362b8d3cfb22045f1cb25274bb72e21e97bdea7ca38d6cf8c0d20",
    ".bin": "f6d164c1887467411e2e6d1130d10f450ce71332ae2336d28e04b2425e5e3bc9"
  }
}

def png(path):
    """沿用已稽核的8-bit PNG濾波還原，另核對signature／CRC及資料長度。"""
    raw = path.read_bytes()
    assert raw[:8] == b'\x89PNG\r\n\x1a\n'
    offset, compressed, palette = 8, b'', None
    while offset < len(raw):
        size = struct.unpack_from('>I', raw, offset)[0]
        kind, value = raw[offset+4:offset+8], raw[offset+8:offset+8+size]
        crc = struct.unpack_from('>I', raw, offset+8+size)[0]
        assert zlib.crc32(kind+value) & 0xffffffff == crc
        offset += size+12
        if kind == b'IHDR':
            width, height, depth, typ, compression, filtering, interlace = struct.unpack('>IIBBBBB', value)
            assert depth == 8 and typ in (2, 3, 6) and (compression, filtering, interlace) == (0, 0, 0)
        elif kind == b'PLTE':
            assert len(value) % 3 == 0
            palette = [tuple(value[i:i+3]) for i in range(0, len(value), 3)]
        elif kind == b'IDAT':
            compressed += value
        elif kind == b'IEND':
            assert offset == len(raw)
            break
    channels = {2: 3, 3: 1, 6: 4}[typ]
    stride = width*channels
    decoded = zlib.decompress(compressed)
    assert len(decoded) == height*(stride+1)
    prior, rows = bytearray(stride), []
    for y in range(height):
        mode = decoded[y*(stride+1)]
        row = bytearray(decoded[y*(stride+1)+1:(y+1)*(stride+1)])
        for x in range(stride):
            left = row[x-channels] if x >= channels else 0
            upper_left = prior[x-channels] if x >= channels else 0
            above = prior[x]
            if mode == 0:
                predictor = 0
            elif mode == 1:
                predictor = left
            elif mode == 2:
                predictor = above
            elif mode == 3:
                predictor = (left+above)//2
            else:
                assert mode == 4
                q = left+above-upper_left
                values = [left, above, upper_left]
                distances = [abs(q-v) for v in values]
                predictor = values[distances.index(min(distances))]
            row[x] = (row[x]+predictor) & 255
        rows.append(bytes(row))
        prior = row
    indices = b''.join(rows) if typ == 3 else None
    rgb = [palette[i] for i in indices] if typ == 3 else [tuple(row[x:x+3]) for row in rows for x in range(0, stride, channels)]
    return width, height, indices, rgb
def digest(p): return hashlib.sha256(p.read_bytes()).hexdigest()
def fields(line): return dict(re.findall(r'(\w+)=(\S+)',line))

def source_receipt(root):
    prefix='issue4-mother-return-gate-r3'
    parent_path=root/'issue4-mother-finish-receipt.json'
    assert digest(parent_path)=='9358ce6e7a8e5c4547555daeccf74a2339f003ab106ebe4d57b32111602c0f3f'
    parent=json.loads(parent_path.read_text())
    meta_path=root/f'{prefix}-meta.json';meta=json.loads(meta_path.read_text())
    assert meta['original_sha256']==parent['original_sha256'] and meta['original_size']==115282
    assert meta['upstream_revision']=='2f44a68ebfc54b28fb15dd4a34510b0b04a5415d'
    assert meta['scenario']=='mother_return_first_down_cold' and meta['state_restore'] is False and meta['gameplay_state_injection'] is False
    assert meta['seed']=='1357' and meta['seed_configured_before_execution'] is True
    assert meta['parent_receipt_sha256']==digest(parent_path) and meta['build_flags']==['-trimpath','-p','2']
    assert meta['producer_sha256']==digest(Path(__file__).with_name('dosgolem_mother_return.py'))
    assert meta['generator_sha256']==digest(Path(__file__).with_name('dosgolem_newgame_probe.py'))
    assert meta['probe_source_sha256']==digest(root/f'{prefix}-probe-source.go')=='ce0565d27a10303b5793bb856755452b25083b9b1b24c0c7aa10dfc6e9d9ed32'
    assert meta['probe_sha256']==digest(root/f'{prefix}-probe')
    retained=[]
    for a in parent['artifacts']:
        if a['path'].endswith(('.png','.bin')):
            p=root/a['path'].replace('issue4-mother-finish-',prefix+'-',1)
            assert p.stat().st_size==a['size'] and digest(p)==a['sha256']
            retained.append(p.name)
    assert len(retained)==174
    lines=(root/(prefix+'.log')).read_text().splitlines()
    assert '沒實作的服務（0 種）：' in lines
    assert not any('找不到的檔（' in l for l in lines)
    irq=[l for l in lines if l.startswith('DQ3_KEY_DELIVERED ')]
    assert len(irq)==78 and irq[:76]==parent['actual_irq1_events']
    assert [int(fields(l)['count']) for l in irq]==list(range(1,79))
    q=[fields(l) for l in lines if l.startswith('DQ3_GATE_QUEUED ')]
    assert len(q)==1 and q[0]['scan']=='50'
    events=[fields(l) for l in lines if l.startswith('DQ3_GATE_CAPTURE ')]
    assert [e['label'] for e in events]==list(EXPECTED_CAPTURES)
    assert [int(e['step']) for e in events]==sorted(int(e['step']) for e in events)
    assert int(q[0]['step'])<int(fields(irq[76])['step'])<int(fields(irq[77])['step'])<int(events[1]['step'])
    assert all(e['flags']==events[0]['flags'] and e['raw0b24']=='000c' and e['player_x']=='21' for e in events)
    assert [e['player_y'] for e in events]==['17']+['18']*6+['17']*3
    assert events[3]['DI']=='0c07' and events[4]['raw259b']=='2'
    assert events[6]['raw4f1f']=='0002' and events[8]['raw0b34']=='00'
    done=[fields(l) for l in lines if l.startswith('DQ3_GATE_DONE ')]
    assert len(done)==1 and done[0]['total_inputs']=='39' and done[0]['irqs']=='78'
    seeds=[l for l in lines if l.startswith('DQ3_CREATION_SEED ')]
    assert len(seeds)==1 and 'fixed=1357' in seeds[0]
    artifacts=[]
    paths=sorted(root.glob(prefix+'-*'))+[root/(prefix+'.log')]
    for p in paths:
        if p.name==prefix+'-receipt.json': continue
        assert p.is_file()
        if p.suffix=='.png':
            w,h,idx,_=png(p);assert (w,h)==(640,350) and idx==p.with_suffix('.bin').read_bytes()
        artifacts.append({'path':p.name,'size':p.stat().st_size,'sha256':digest(p)})
    for label,expected in EXPECTED_CAPTURES.items():
        for suffix,hash_value in expected.items(): assert digest(root/(prefix+'-'+label+suffix))==hash_value
    return {'scope':'正常母親返回後首次Down；自動文字、關窗及強制北行有限來源',
            'original_sha256':parent['original_sha256'],'parent_receipt_sha256':digest(parent_path),
            'parent_174_png_bin_unchanged':True,'normal_inputs':39,'actual_irq1_events':irq,
            'events':events,'artifacts':artifacts,'seed':'1357','seed_control_once':True,
            'no_state_restore':True,'flags_unchanged':True,'no_extra_dismiss_input':True,
            'registration_parity':False,'remake_parity':False,'audio_parity':False,'original_save_load_parity':False}

def validate(path):
    assert path.name=='issue4-mother-return-gate-r3-receipt.json'
    received=json.loads(path.read_text());expected=source_receipt(path.parent)
    assert received==expected,'receipt differs from independently verified source'
    return expected

def main():
    p=argparse.ArgumentParser();p.add_argument('receipt',type=Path);p.add_argument('--write',action='store_true');args=p.parse_args()
    if args.write:
        assert not args.receipt.exists()
        data=source_receipt(args.receipt.parent)
        args.receipt.write_text(json.dumps(data,ensure_ascii=False,indent=2)+'\n')
    data=validate(args.receipt)
    print(f"原版39次正常輸入／78 IRQ1／174父圖像及10個事件邊界 PASS；收據SHA-256 {digest(args.receipt)}")

if __name__=='__main__': main()
