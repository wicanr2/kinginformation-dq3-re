"""稽核入隊後等待的有限 DRAFT 證據；容器用法與限制見 docs/188。"""
import argparse
import json
from pathlib import Path

from verify_dosgolem_recruitment_entry import validate as validate_entry
from verify_dosgolem_mother_return import digest, fields, png

PREFIX = "issue4-recruit-normal-r5"
PREFIX_HASH = "8e49fb70b2b2a4428585a2391eac6cc56fef3dca30cc9206b93cd9741f0c2f0d"
LOG_HASH = "3973569b328d810dc0f0d1da368da170716f41424f94062044c36cfe4bc52134"
META_HASH = "b11a2e2eb9fb41610c7667704c9a107c68f75a287c538108f12e1c77e15309f7"
IDA_HASH = "1274a0628dfaf8f28465cbd064e16fe8d3de44ecf3173f2e9477bba589bb9658"
REMAKE_HASH = "5563edb23e4860922013b9b238ad6dd01a479da07bec016043a7d5340c998132"


def validate(root, ida, remake):
    parent = validate_entry(root / "issue4-recruit-normal-r4-source-r2-receipt.json")
    prefix = root / (PREFIX + "-prefix198-draft-r1-receipt.json")
    assert digest(prefix) == PREFIX_HASH
    d = json.loads(prefix.read_text())
    assert d["complete_join"] is False and d["audio_parity"] is False
    assert d["parent_sha256"] == digest(root / "issue4-recruit-normal-r4-source-r2-receipt.json")
    log, meta_path = root / (PREFIX + ".log"), root / (PREFIX + "-meta.json")
    assert digest(log) == LOG_HASH and digest(meta_path) == META_HASH
    meta = json.loads(meta_path.read_text())
    assert meta == d["meta"]
    for suffix, key in [("-probe-source.go", "probe_source_sha256"), ("-probe", "probe_sha256")]:
        assert digest(root / (PREFIX + suffix)) == meta[key]
    assert not meta["state_restore"] and not meta["gameplay_state_injection"]
    assert meta["seed"] == "1357" and meta["seed_configured_before_execution"]
    lines = log.read_text().splitlines()
    for name, tag in [("queued", "DQ3_QUIESCENT_QUEUED"), ("consumed", "DQ3_QUIESCENT_INPUT"),
                      ("states", "DQ3_QUIESCENT_CAPTURE"), ("actual_irq1_events", "DQ3_KEY_DELIVERED")]:
        rows = [fields(l) for l in lines if l.startswith(tag + " ")]
        count = 472 if name == "actual_irq1_events" else 198
        assert rows[:count] == d[name]
        assert d[name][:len(parent[name])] == parent[name]
        assert len(rows) == {"queued": 199, "consumed": 199, "states": 198,
                             "actual_irq1_events": 474}[name]
    for item in d["artifacts"]:
        p = root / item["path"]
        assert p.name == item["path"] and p.stat().st_size == item["size"]
        assert digest(p) == item["sha256"]
        if p.suffix == ".png":
            width, height, indices, _ = png(p)
            assert (width, height) == (640, 350) and indices == p.with_suffix(".bin").read_bytes()
    assert len(d["artifacts"]) == 396
    timer = [fields(l) for l in lines if l.startswith("DQ3_TIMER_DRAFT ")]
    stop = [fields(l) for l in lines if l.startswith("DQ3_RECRUIT_DONE ")]
    assert len(timer) == len(stop) == 1
    assert timer[0] == {
        "step": "3776635393", "packet": "199", "pc": "1ff02", "previous_pc": "1fefc",
        "flag": "4002", "counter": "4e20", "record": "538", "irqs": "474",
    }
    assert "DRAFT_TIMER_STOP" in next(l for l in lines if l.startswith("DQ3_RECRUIT_DONE "))
    assert stop[0]["step"] == "3796635394" and stop[0]["pc"] == "208f3"
    assert stop[0]["record"] == "538" and stop[0]["packet"] == "199"
    assert digest(ida) == IDA_HASH
    sidecar = json.loads(ida.read_text())
    assert sidecar["tool"]["version"] == "9.4"
    assert sidecar["input"]["sha256"] == meta["original_sha256"]
    tools = Path(__file__).parent
    assert sidecar["tool"]["script_sha256"] == digest(tools / "ida_dump_recruitment_join_wait_contract.py")
    ledger = tools / "ida_recruitment_join_wait_ledger.json"
    assert sidecar["reviewed_ledger"]["sha256"] == digest(ledger)
    assert sidecar["reviewed_ledger"]["annotations"] == json.loads(ledger.read_text())["annotations"]
    rows = {r["ida_linear"]: r for sec in sidecar["ranges"] for r in sec["instructions"]}
    annotations = sidecar["reviewed_ledger"]["annotations"]
    assert len(annotations) == 14
    for a in annotations:
        r = rows[a["ida_linear"]]
        assert r["file_offset"] == a["file_offset"] and r["file_bytes"].startswith(a["bytes"])
        assert r["inference_level"] == a["inference_level"]
        assert r["added_semantic"] == a["semantic"] and r["source_sha256"] == meta["original_sha256"]
    assert digest(remake) == REMAKE_HASH
    runtime = json.loads(remake.read_text())
    assert runtime["complete_join_parity"] is False and runtime["original_join_dialogue_missing"] is True
    assert runtime["packet198_remake_stage"] == 1 and not runtime["packet198_remake_dialogue_open"]
    assert runtime["packet198_remake_roster"] == 0 and runtime["packet198_remake_companions"] == 1
    assert [s["full_rgb_difference"] for s in runtime["samples"][-2:]] == [51535, 56790]
    return {
        "status": "DRAFT evidence verified; complete join and playback root cause unresolved",
        "prefix_packets": 198, "complete_irq1": 472, "uncompleted_packet": 199,
        "first_flag_writer": "IDA linear1FEFC; file1126C; DGROUP0013",
        "wait_stop": "IDA linear208F3; file11C63",
        "full_rgb_difference": [51535, 56790],
        "production_changed": False, "complete_join_parity": False,
    }


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("source_root", type=Path)
    p.add_argument("ida", type=Path)
    p.add_argument("remake", type=Path)
    args = p.parse_args()
    print(json.dumps(validate(args.source_root, args.ida, args.remake), ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
