"""正常觀看名單第一個選人等待的來源稽核；入口與限制見 docs/188。"""
import argparse
import json
from pathlib import Path

from verify_dosgolem_recruitment_entry import validate as validate_entry
from verify_dosgolem_mother_return import digest, fields, png

PREFIX = "issue4-recruit-view-r1"


def source_receipt(root, producer):
    parent_path = root / "issue4-recruit-normal-r4-source-r2-receipt.json"
    parent = validate_entry(parent_path)
    meta = json.loads((root / (PREFIX + "-meta.json")).read_text())
    for key in ("original_size", "original_sha256", "upstream_revision", "seed",
                "seed_configured_before_execution", "parent_receipt_sha256", "generator_sha256",
                "build_flags", "normal_prefix_inputs", "state_restore", "gameplay_state_injection",
                "upstream_files_sha256", "patched_files_sha256", "upstream_bios_sha256",
                "patched_bios_sha256", "upstream_vga_sha256", "patched_vga_sha256"):
        assert meta[key] == parent["meta"][key], key
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
    assert len(queued) == len(consumed) == len(states) == 199 and len(irq) == 474
    assert [int(q["scan"], 16) for q in queued[196:]] == [0x50, 0x50, 0x1c]
    assert [q["kind"] for q in queued[196:]] == ["recruit_view_cursor", "recruit_view_cursor", "recruit_view"]
    assert [s["phase"] for s in states[196:]] == ["choice"] * 3
    assert [s["last_record"] for s in states[196:]] == ["528"] * 3
    assert [s["choice_count"] for s in states[196:]] == ["3", "3", "1"]
    assert [s["choice_cursor"] for s in states[196:]] == ["2", "3", "1"]
    scans = [scan for _, scan in meta["normal_prefix_inputs"]] + [int(q["scan"], 16) for q in queued]
    assert [int(x["port60"], 16) for x in irq] == [v for scan in scans for v in (scan, scan | 128)]
    assert [int(x["count"]) for x in irq] == list(range(1, 475))
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
            if number <= 196:
                prior = root / ("issue4-recruit-normal-r4" + f"-packet-{number:03d}-" + s["phase"] + ext)
                assert path.read_bytes() == prior.read_bytes()
        if number >= 172:
            assert s["gold_lo"] == "0032" and s["gold_hi"] == "0000" and s["flags"] == states[171]["flags"]
        if number > 196:
            assert all(s[k] == parent["states"][-1][k] for k in ("player_x", "player_y", "raw0b24", "actor"))
    observations = [fields(line) for line in lines if line.startswith("DQ3_RECRUIT_STATE ")]
    assert len(observations) == 28
    assert all(o["roster_flags"] == "020100000000000000000000" for o in observations)
    assert len({o["slot1"] for o in observations}) == 1
    assert observations[0]["slot1"] == parent["recruit_observations"][0]["slot1"]
    entry = [fields(line) for line in lines if line.startswith("DQ3_VIEW_ENTRY ")]
    assert len(entry) == 1 and entry[0]["packet"] == "199" and entry[0]["DS"] == "15ed"
    assert entry[0]["ida_linear"] == "10624" and entry[0]["choice_cursor"] == "3"
    assert int(consumed[-1]["step"]) < int(entry[0]["step"]) < int(states[-1]["step"])
    done = [fields(line) for line in lines if line.startswith("DQ3_RECRUIT_DONE ")]
    assert len(done) == 1 and done[0]["step"] == states[-1]["step"]
    assert done[0]["packets"] == "199" and done[0]["irqs"] == "474"
    return {"original_sha256": meta["original_sha256"], "upstream_revision": meta["upstream_revision"],
            "seed": meta["seed"], "scope": "正常出生戰士男性、觀看名單第一個選人等待；未接受詳細狀況或返回",
            "meta": meta, "log_sha256": digest(root / (PREFIX + ".log")), "parent_sha256": digest(parent_path),
            "normal_inputs": 237, "artifacts": artifacts, **groups, "recruit_observations": observations,
            "view_entry": entry[0], "done": done[0], "prefix196_unchanged": True,
            "transaction_unchanged": True, "complete_view": False, "remake_parity": False,
            "original_save_load_parity": False}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("root", type=Path)
    parser.add_argument("producer", type=Path)
    parser.add_argument("receipt", type=Path)
    args = parser.parse_args()
    assert not args.receipt.exists(), "refuse to overwrite existing receipt"
    report = source_receipt(args.root, args.producer)
    with args.receipt.open("x") as stream:
        json.dump(report, stream, ensure_ascii=False, indent=2)
        stream.write("\n")
    print("原版觀看名單來源 PASS", len(report["states"]), len(report["artifacts"]), digest(args.receipt))


if __name__ == "__main__":
    main()
