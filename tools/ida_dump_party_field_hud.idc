#include <idc.idc>

// 地表隊伍 HUD 的非破壞性 IDA 9.4 匯出。只輸出原始函式名、linear 位址、
// callers 與指令；語意與推論等級另存於受版控 evidence 文件，不改名原資料庫。
static audit_line(fp, text)
{
  msg(text + "\n");
  if (fp != 0)
    fprintf(fp, "%s\n", text);
}

static func_label(ea)
{
  auto start;
  start = get_func_attr(ea, FUNCATTR_START);
  if (start == BADADDR)
    return "<no-func>";
  return get_func_name(start);
}

static dump_func(ea, fp)
{
  auto start, end, cur, x;
  start = get_func_attr(ea, FUNCATTR_START);
  if (start == BADADDR)
  {
    audit_line(fp, "IDA_FUNC none requested=" + atoa(ea));
    return;
  }
  end = get_func_attr(ea, FUNCATTR_END);
  audit_line(fp, "IDA_FUNC start=" + atoa(start) + " end=" + atoa(end) +
             " name=" + get_func_name(start));
  for (x = get_first_cref_to(start); x != BADADDR;
       x = get_next_cref_to(start, x))
    audit_line(fp, "IDA_CALLER ea=" + atoa(x) + " func=" + func_label(x));
  for (cur = start; cur != BADADDR && cur < end; cur = next_head(cur, end))
    audit_line(fp, "IDA_DISASM " + atoa(cur) + " " + generate_disasm_line(cur, 0));
}

static main()
{
  auto fp, input_fp, input_size;
  auto_wait();
  fp = fopen("party-field-hud-ida.txt", "w");
  if (fp == 0)
    qexit(1);
  input_size = -1;
  input_fp = fopen(get_input_file_path(), "rb");
  if (input_fp != 0)
  {
    input_size = filelength(input_fp);
    fclose(input_fp);
  }
  audit_line(fp, "IDA_AUDIT party-field-hud tool=IDA Pro 9.4 address_basis=linear");
  audit_line(fp, "IDA_INPUT path=" + get_input_file_path());
  audit_line(fp, "IDA_INPUT bytes=" + ltoa(input_size, 10));
  audit_line(fp, "IDA_INPUT md5=" + retrieve_input_file_md5());
  audit_line(fp, "IDA_ADDRESS_BASIS linear=logical+0x10000 file=logical+0x1370 DS=DGROUP");
  dump_func(0x17C83, fp); // field command owner and DS:3E9C geometry writer
  dump_func(0x18222, fp); // party name/HP/MP/class/level writer
  dump_func(0x1C572, fp); // battle caller of the same HUD writer
  dump_func(0x1F590, fp); // window consumer
  dump_func(0x211B6, fp); // glyph writer
  dump_func(0x215EE, fp); // four-glyph name writer
  dump_func(0x219D2, fp); // numeric writer
  fclose(fp);
  qexit(0);
}
