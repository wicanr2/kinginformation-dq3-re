"""非破壞創角匯出；證據台帳與推論等級見 docs/113-newgame-geometry-re.md。"""
import hashlib,json,os
import ida_auto,ida_bytes,ida_funcs,ida_kernwin,ida_lines,ida_loader,ida_nalt,idautils,idc
ida_auto.auto_wait()
path=ida_nalt.get_input_file_path()
blob=open(path,'rb').read()
sha=hashlib.sha256(blob).hexdigest()
assert len(blob)==115282 and sha=='5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c'
# 受版控的有限審查台帳；其餘資料保持 unknown。範圍以原始定位為 key，
# 不從散文自動抽取或覆蓋 IDA 名稱。strong 一律附醒目限制。
ledger=[
 (0x10854,0x10924,'strong','主角創角 caller：命名／性別／生成／確認；本批只動態閉合生成及提示出現，接受／重來尚待驗'),
 (0x16f4b,0x16fcf,'strong','預設名稱的 BIOS tick→CS:701B；不能當成能力 DGROUP0B5A 種子'),
 (0x1d9cc,0x1db48,'strong','class／level與DS4366→間接角色record的七能力writer→sub_181B1；其他職業／等級未由本批驗收'),
 (0x1e6b9,0x1e6c9,'strong','DGROUP0B5A 的16-bit加9018／旋轉3 writer；不代表全遊戲骰序'),
 (0x1e6e7,0x1e713,'strong','AX正delta→同一種子前進→餘數0改1；原版自然Lv1七次呼叫已記錄'),
 (0x1f4e3,0x1f590,'strong','raw window的frame／shadow／record／選項consumer；本批性別預設男性全畫布已闭合'),
 (0x2111b,0x21148,'strong','等待輸入再返回的far helper；自然能力面板Enter已閉合，不追硬體時序'),
 (0x28bc6,0x28be4,'confirmed','性別raw window：record556及原始游標；與能力選項共用結構但不可共用畫面背景結論'),
]
ledger=[(a,b,level,semantic.replace('闭合','閉合')) for a,b,level,semantic in ledger]
def row(ea):
 item={'ida_linear':hex(ea),'file_offset':hex(ida_loader.get_fileregion_offset(ea)),
 'original_name':idc.get_name(ea),'bytes':(ida_bytes.get_bytes(ea,ida_bytes.get_item_size(ea)) or b'').hex(),
 'disassembly':ida_lines.generate_disasm_line(ea,ida_lines.GENDSM_REMOVE_TAGS),
 'source_path':path,'source_size':len(blob),'source_sha256':sha,
 'inference_level':'unknown','semantic':'UNKNOWN：未審查創角／亂數證據',
 'evidence':'本次非破壞 IDA database 匯出；舊文件只供定位'}
 for start,end,level,semantic in ledger:
  if start<=ea<end:
   item.update(inference_level=level,semantic=semantic,
    reviewed_range={'ida_linear_start':hex(start),'ida_linear_end_exclusive':hex(end)},
    evidence='docs/113-newgame-geometry-re.md：2026-10-01有限證據審查；原版issue4-creation-receipt.json固定種子自然IRQ1；原始bytes及database xref')
   break
 item['warning']='' if item['inference_level']=='confirmed' else '⚠ '+item['inference_level']+'：未達已證實，不得把語意名稱當事實'
 return item
def refs(ea):
 return [{**row(r.frm),'xref_type':r.type,'original_function':ida_funcs.get_func_name(r.frm)} for r in idautils.XrefsTo(ea)]
anchors=[0x16f4b,0x1e6b9,0x10854,0x1d9cc,0x1e6e7,0x1834e,0x184a1,0x2111b,0x1f4e3,0x1f63c,0x28bc6,0x2592a]
starts=set()
for a in anchors:
 f=ida_funcs.get_func(a)
 if f:starts.add(f.start_ea)
 for r in idautils.XrefsTo(a):
  f=ida_funcs.get_func(r.frm)
  if f and a==0x2592a:starts.add(f.start_ea)
functions=[]
for start in sorted(starts):
 f=ida_funcs.get_func(start)
 functions.append({**row(start),'original_function':ida_funcs.get_func_name(start),
 'end_ida_linear':hex(f.end_ea),'xrefs':refs(start),'instructions':[row(ea) for ea in idautils.FuncItems(start)]})
result={'input':{'path':path,'size':len(blob),'sha256':sha},'tool':{'name':'IDA Pro','version':ida_kernwin.get_kernel_version(),'address_space':'IDA linear；MZ file=linear−0xEC90；DGROUP基底linear0x24DD0'},
 'warning':'保留原名、位址、bytes及運算元；不改名／patch／函式邊界；直接xref不排除間接寫入。',
 'review_ledger':[{'ida_linear_start':hex(a),'ida_linear_end_exclusive':hex(b),'inference_level':level,'semantic':semantic} for a,b,level,semantic in ledger],
 'anchors':[{**row(a),'xrefs':refs(a)} for a in anchors], 'functions':functions}
with open(idc.ARGV[1]+'.tmp','w',encoding='utf-8') as stream:json.dump(result,stream,ensure_ascii=False,indent=2)
os.replace(idc.ARGV[1]+'.tmp',idc.ARGV[1])
idc.qexit(0)
