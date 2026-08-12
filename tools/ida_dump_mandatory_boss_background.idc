#include <idc.idc>

// 必經頭目背景盤點的非破壞性 IDA 9.4 匯出。它保留原始 linear 位址與函式名；
// 語意名稱與推論等級另放在受版控的 evidence ledger，不覆寫原始定位。
//
// 本腳本會在 IDA 目前工作目錄寫出明確的 sidecar
// mandatory-boss-background-ida.txt；它包含輸入路徑、大小、IDA 的 input MD5
// 與位址基準。外層 Docker runner 另將 SHA-256 寫入同一份 sidecar header，且
// 必須在使用前確認檔案非空。

static audit_line(fp, line)
{
  msg(line + "\n");
  if ( fp != 0 )
    fprintf(fp, "%s\n", line);
}

static func_label(ea)
{
  auto start;
  start = get_func_attr(ea, FUNCATTR_START);
  if ( start == BADADDR )
    return "<no-func>";
  return get_func_name(start);
}

static dump_func(ea, fp)
{
  auto start, end, cur, x;

  start = get_func_attr(ea, FUNCATTR_START);
  if ( start == BADADDR )
  {
    audit_line(fp, "IDA_FUNC none requested=" + atoa(ea));
    return;
  }
  end = get_func_attr(ea, FUNCATTR_END);
  audit_line(fp, "IDA_FUNC start=" + atoa(start) + " end=" + atoa(end) +
             " name=" + get_func_name(start));
  for ( x = get_first_cref_to(start); x != BADADDR;
        x = get_next_cref_to(start, x) )
    audit_line(fp, "IDA_CALLER ea=" + atoa(x) + " func=" + func_label(x));
  for ( cur = start; cur != BADADDR && cur < end; cur = next_head(cur, end) )
    audit_line(fp, "IDA_DISASM " + atoa(cur) + " " + generate_disasm_line(cur, 0));
}

static main()
{
  auto audit_fp, input_fp, input_size;

  auto_wait();
  audit_fp = fopen("mandatory-boss-background-ida.txt", "w");
  if ( audit_fp == 0 )
  {
    msg("IDA_AUDIT_ERROR cannot open mandatory-boss-background-ida.txt\n");
    qexit(1);
  }
  input_size = -1;
  input_fp = fopen(get_input_file_path(), "rb");
  if ( input_fp != 0 )
  {
    input_size = filelength(input_fp);
    fclose(input_fp);
  }
  audit_line(audit_fp, "IDA_AUDIT mandatory-boss-background tool=IDA Pro 9.4 address_basis=linear");
  audit_line(audit_fp, "IDA_INPUT path=" + get_input_file_path());
  audit_line(audit_fp, "IDA_INPUT bytes=" + ltoa(input_size, 10));
  audit_line(audit_fp, "IDA_INPUT md5=" + retrieve_input_file_md5());
  audit_line(audit_fp, "IDA_ADDRESS_BASIS linear=logical+0x10000 file=logical+0x1370");
  dump_func(0x14312, audit_fp); // raw fixed record DS:4ee4: monster 0x59
  dump_func(0x1622a, audit_fp); // raw fixed record DS:4edf: monster 0x79
  dump_func(0x164cd, audit_fp); // raw fixed record DS:4ef0: monster 0x7c
  dump_func(0x165e7, audit_fp); // raw fixed record DS:4efa: monster 0x7a
  dump_func(0x1661e, audit_fp); // raw fixed record DS:4eff: monster 0x7b
  dump_func(0x1be89, audit_fp); // fixed runner
  dump_func(0x1bf35, audit_fp); // fixed formation loader/writer
  dump_func(0x1bfd1, audit_fp); // battle scene setup
  dump_func(0x1c688, audit_fp); // palette/page selector consumer
  dump_func(0x1c6e5, audit_fp); // PACKBG archive reader
  fclose(audit_fp);
  qexit(0);
}
