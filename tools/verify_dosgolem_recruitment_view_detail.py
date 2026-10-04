"""稽核冷啟動觀看名單詳細狀況、關閉與返回來源；入口 docs/188。"""
import argparse
import json
from pathlib import Path

from verify_dosgolem_recruitment_view import source_receipt
from verify_dosgolem_mother_return import digest, fields, png

PREFIX = "issue4-recruit-view-detail-r1"


def validate(root, producer, view_producer):
    parent_path = root / "issue4-recruit-view-r1-source-r1-receipt.json"
    parent = source_receipt(root, view_producer)
    assert json.loads(parent_path.read_text()) == parent, "view parent receipt changed"
    meta = json.loads((root / (PREFIX + "-meta.json")).read_text())
    for key, value in parent["meta"].items():
        if key == "args":
            assert meta[key] == [arg.replace("issue4-recruit-view-r1", PREFIX) for arg in value], key
            continue
        if key not in ("producer_sha256", "probe_source_sha256", "probe_sha256"):
            assert meta[key] == value, key
    assert meta["producer_sha256"] == digest(producer)
    for suffix, key in (("-probe-source.go", "probe_source_sha256"), ("-probe", "probe_sha256")):
        assert digest(root / (PREFIX + suffix)) == meta[key]
    lines = (root / (PREFIX + ".log")).read_text().splitlines()
    groups = {}
    for name, tag in (("queued", "DQ3_QUIESCENT_QUEUED"), ("consumed", "DQ3_QUIESCENT_INPUT"),
                      ("states", "DQ3_QUIESCENT_CAPTURE"), ("actual_irq1_events", "DQ3_KEY_DELIVERED")):
        groups[name] = [fields(line) for line in lines if line.startswith(tag + " ")]
        assert groups[name][:len(parent[name])] == parent[name], name
    queued, consumed, states, irq = (groups[k] for k in ("queued", "consumed", "states", "actual_irq1_events"))
    assert len(queued) == len(consumed) == len(states) == 205 and len(irq) == 486
    assert [int(q["scan"], 16) for q in queued[199:]] == [0x1c, 0x1c, 0x01, 0x4d, 0x1c, 0x1c]
    assert [q["kind"] for q in queued[199:]] == ["recruit_view_pick", "recruit_wait", "recruit_view_detail_close", "recruit_view_decline_cursor", "recruit_view_decline", "recruit_wait"]
    assert [s["phase"] for s in states[199:]] == ["waiting", "view_detail_wait", "choice", "choice", "waiting", "ready"]
    assert [s["last_record"] for s in states[199:]] == ["528", "528", "540", "540", "541", "541"]
    assert [s["choice_count"] for s in states[199:]] == ["1", "1", "2", "2", "2", "2"]
    assert [s["choice_cursor"] for s in states[199:]] == ["5", "5", "1", "2", "2", "2"]
    assert [c["ida_linear"] for c in consumed[199:]] == ["210ca", "2113a", "2110b", "210ca", "210ca", "2113a"]
    assert [s["ida_linear"] for s in states[199:]] == ["21133", "21103", "1f7b7", "1f7b7", "21133", "1997c"]
    scans = [scan for _, scan in meta["normal_prefix_inputs"]] + [int(q["scan"], 16) for q in queued]
    assert [int(x["port60"], 16) for x in irq] == [v for scan in scans for v in (scan, scan | 128)]
    assert [int(x["count"]) for x in irq] == list(range(1, 487))
    artifacts = []
    for number, (q, c, s) in enumerate(zip(queued, consumed, states), 1):
        assert q["packet"] == c["packet"] == s["packet"] == str(number)
        assert q["scan"] == c["scan"] == s["scan"] and q["kind"] == s["kind"]
        assert int(q["step"]) == int(s["queued_step"]) < int(c["step"]) < int(s["step"])
        assert int(c["irqs"]) == 76 + number * 2 - 1
        assert int(irq[76 + number * 2 - 1]["step"]) <= int(s["step"])
        assert int(s["raw0013"], 16) & 0x4000 == 0
        if number > 1:
            assert q["step"] == states[number - 2]["step"] and q["phase"] == states[number - 2]["phase"]
        name = PREFIX + f"-packet-{number:03d}-" + s["phase"]
        width, height, indices, _ = png(root / (name + ".png"))
        assert (width, height) == (640, 350) and indices == (root / (name + ".bin")).read_bytes()
        for ext in (".png", ".bin"):
            path = root / (name + ext)
            artifacts.append({"path": path.name, "size": path.stat().st_size, "sha256": digest(path)})
            if number <= 199:
                prior = root / ("issue4-recruit-view-r1" + f"-packet-{number:03d}-" + s["phase"] + ext)
                assert path.read_bytes() == prior.read_bytes()
        if number >= 172:
            assert s["gold_lo"] == "0032" and s["gold_hi"] == "0000" and s["flags"] == states[171]["flags"]
        if number > 199:
            assert all(s[k] == parent["states"][-1][k] for k in ("player_x", "player_y", "raw0b24", "actor"))
    observations = [fields(line) for line in lines if line.startswith("DQ3_RECRUIT_STATE ")]
    assert len(observations) == 34
    assert all(o["roster_flags"] == "020100000000000000000000" and o["slot1"] == parent["recruit_observations"][0]["slot1"] for o in observations)
    for key in ("raw4f1f", "raw4f15", "raw4f17", "raw4f19", "raw4f1b"):
        assert len({o[key] for o in observations}) == 1, key
    detail = [fields(line) for line in lines if line.startswith("DQ3_VIEW_DETAIL ")]
    assert [o["ida_linear"] for o in detail] == ["10668", "10671", "1834e", "1068e"]
    assert [o["packet"] for o in detail] == ["200", "200", "200", "202"]
    assert [o["choice_cursor"] for o in detail] == ["1", "1", "5", "5"]
    assert all(o["DS"] == "15ed" and o["raw4f1d"] == "520b" for o in detail)
    assert all(o["actor520b"] == observations[0]["slot1"] for o in detail)
    assert int(consumed[199]["step"]) < int(detail[0]["step"]) < int(detail[1]["step"]) < int(detail[2]["step"]) < int(states[199]["step"])
    assert int(consumed[201]["step"]) < int(detail[3]["step"]) < int(states[201]["step"])
    entry = [fields(line) for line in lines if line.startswith("DQ3_VIEW_ENTRY ")]
    assert entry == [parent["view_entry"]]
    done = [fields(line) for line in lines if line.startswith("DQ3_RECRUIT_DONE ")]
    assert len(done) == 1 and done[0]["step"] == states[-1]["step"]
    assert done[0]["packets"] == "205" and done[0]["irqs"] == "486"
    assert states[-1]["ida_linear"] == "1997c"
    for ext in (".png", ".bin"):
        assert (root / (PREFIX + "-packet-200-waiting" + ext)).read_bytes() == (root / (PREFIX + "-packet-201-view_detail_wait" + ext)).read_bytes(), "ability canvas changed between waits"
    return {"original_sha256": meta["original_sha256"], "upstream_revision": meta["upstream_revision"],
            "seed": meta["seed"], "scope": "正常出生戰士男性、觀看詳細狀況、能力窗按鍵、Esc關閉、No告別與場景返回；不含改名或有法術角色",
            "meta": meta, "log_sha256": digest(root / (PREFIX + ".log")), "parent_sha256": digest(parent_path),
            "normal_inputs": 243, "artifacts": artifacts, **groups, "recruit_observations": observations,
            "view_entry": entry[0], "view_detail": detail, "done": done[0], "prefix199_unchanged": True,
            "transaction_unchanged": True, "complete_view": False, "remake_parity": False,
            "original_save_load_parity": False, "view_spec_ready": False}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("root", type=Path)
    parser.add_argument("producer", type=Path)
    parser.add_argument("view_producer", type=Path)
    parser.add_argument("receipt", type=Path)
    args = parser.parse_args()
    assert not args.receipt.exists(), "refuse to overwrite existing receipt"
    report = validate(args.root, args.producer, args.view_producer)
    with args.receipt.open("x") as stream:
        json.dump(report, stream, ensure_ascii=False, indent=2)
        stream.write("\n")
    print("原版觀看名單詳細狀況來源 PASS", len(report["states"]), len(report["artifacts"]), digest(args.receipt))


if __name__ == "__main__":
    main()
