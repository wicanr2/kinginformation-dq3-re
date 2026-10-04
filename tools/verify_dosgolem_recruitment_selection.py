"""正常選人取消來源稽核。入口、範圍及 Docker 用法見 docs/188。"""
import argparse
import json
from pathlib import Path

from verify_dosgolem_recruitment_entry import validate as validate_entry
from verify_dosgolem_mother_return import digest, fields, png

PREFIX = "issue4-recruit-selection-cancel-r1"


def source_receipt(root, producer, continue_yes=False):
    prefix = "issue4-recruit-yes-r1" if continue_yes else PREFIX
    packets = 204 if continue_yes else 201
    irq_count = 76 + packets * 2
    parent = validate_entry(root / "issue4-recruit-normal-r4-source-r2-receipt.json")
    meta = json.loads((root / (prefix + "-meta.json")).read_text())
    for key in ("original_size", "original_sha256", "upstream_revision", "seed",
                "seed_configured_before_execution", "parent_receipt_sha256", "generator_sha256",
                "build_flags", "normal_prefix_inputs", "state_restore", "gameplay_state_injection",
                "upstream_files_sha256", "patched_files_sha256", "upstream_bios_sha256",
                "patched_bios_sha256", "upstream_vga_sha256", "patched_vga_sha256"):
        assert meta[key] == parent["meta"][key], key
    assert meta["producer_sha256"] == digest(producer)
    for suffix, key in (("-probe-source.go", "probe_source_sha256"), ("-probe", "probe_sha256")):
        assert digest(root / (prefix + suffix)) == meta[key]
    lines = (root / (prefix + ".log")).read_text().splitlines()
    groups = {}
    for name, tag in (("queued", "DQ3_QUIESCENT_QUEUED"), ("consumed", "DQ3_QUIESCENT_INPUT"),
                      ("states", "DQ3_QUIESCENT_CAPTURE"), ("actual_irq1_events", "DQ3_KEY_DELIVERED")):
        groups[name] = [fields(line) for line in lines if line.startswith(tag + " ")]
        assert groups[name][:len(parent[name])] == parent[name], name
    queued, consumed, states, irq = (groups[key] for key in ("queued", "consumed", "states", "actual_irq1_events"))
    assert len(queued) == len(consumed) == len(states) == packets
    assert len(irq) == irq_count
    scans_expected = [0x1c, 0x01, 0x1c, 0x1c, 0x01, 0x4d, 0x1c, 0x1c] if continue_yes else [0x1c, 0x01, 0x4d, 0x1c, 0x1c]
    records = ["530", "540", "528", "530", "540", "540", "541", "541"] if continue_yes else ["530", "540", "540", "541", "541"]
    counts = ["1", "2", "3", "1", "2", "2", "2"] if continue_yes else ["1", "2", "2", "2"]
    cursors = ["1", "1", "1", "1", "1", "2", "2"] if continue_yes else ["1", "1", "2", "2"]
    assert [int(q["scan"], 16) for q in queued[196:]] == scans_expected
    assert [s["phase"] for s in states[196:]] == ["choice"] * (packets - 198) + ["waiting", "ready"]
    assert [s["last_record"] for s in states[196:]] == records
    assert [s["choice_count"] for s in states[196:packets - 1]] == counts
    assert [s["choice_cursor"] for s in states[196:packets - 1]] == cursors
    for state in states[196:]:
        for key in ("player_x", "player_y", "raw0b24"):
            assert state[key] == parent["states"][-1][key], key
    scans = [scan for _, scan in meta["normal_prefix_inputs"]] + [int(q["scan"], 16) for q in queued]
    assert [int(x["port60"], 16) for x in irq] == [v for scan in scans for v in (scan, scan | 128)]
    assert [int(x["count"]) for x in irq] == list(range(1, irq_count + 1))
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
        name = prefix + f"-packet-{number:03d}-" + s["phase"]
        width, height, indices, _ = png(root / (name + ".png"))
        assert (width, height) == (640, 350) and indices == (root / (name + ".bin")).read_bytes()
        for ext in (".png", ".bin"):
            path = root / (name + ext)
            artifacts.append({"path": path.name, "size": path.stat().st_size, "sha256": digest(path)})
            if number <= 196:
                prior = root / ("issue4-recruit-normal-r4" + f"-packet-{number:03d}-" + s["phase"] + ext)
                assert path.read_bytes() == prior.read_bytes()
            if number == 197:
                prior = root / ("issue4-recruit-normal-r5-packet-197-choice" + ext)
                assert path.read_bytes() == prior.read_bytes()
        if number >= 172:
            assert s["gold_lo"] == "0032" and s["gold_hi"] == "0000" and s["flags"] == states[171]["flags"]
    observations = [fields(line) for line in lines if line.startswith("DQ3_RECRUIT_STATE ")]
    assert len(observations) == packets - 171
    assert all(row["roster_flags"] == "020100000000000000000000" for row in observations)
    assert len({row["slot1"] for row in observations}) == 1
    assert observations[0]["slot1"] == parent["recruit_observations"][0]["slot1"]
    done = [fields(line) for line in lines if line.startswith("DQ3_RECRUIT_DONE ")]
    assert len(done) == 1 and done[0]["step"] == states[-1]["step"]
    assert done[0]["packets"] == str(packets) and done[0]["irqs"] == str(irq_count)
    return {"original_sha256": meta["original_sha256"], "upstream_revision": meta["upstream_revision"], "seed": meta["seed"],
            "scope": "正常出生一名戰士男性、清單Esc、Yes、再進清單Esc、No、告別及場景返回" if continue_yes else "正常出生一名戰士男性、下樓、入隊清單Esc、繼續詢問No、告別等待、場景返回",
            "meta": meta, "log_sha256": digest(root / (prefix + ".log")),
            "parent_sha256": digest(root / "issue4-recruit-normal-r4-source-r2-receipt.json"),
            "normal_inputs": 38 + packets, "artifacts": artifacts, **groups,
            "recruit_observations": observations, "done": done[0],
            "prefix196_unchanged": True, "selection197_matches_join_prefix": True,
            "transaction_unchanged": True, "remake_parity": False,
            "complete_join": False, "audio_parity": False, "original_save_load_parity": False}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("root", type=Path)
    parser.add_argument("producer", type=Path)
    parser.add_argument("receipt", type=Path)
    parser.add_argument("--continue-yes", action="store_true", help="核對Yes後重新進選人、再取消No的204包正常路線")
    args = parser.parse_args()
    report = source_receipt(args.root, args.producer, args.continue_yes)
    if args.receipt.exists():
        assert json.loads(args.receipt.read_text()) == report
    else:
        with args.receipt.open("x") as stream:
            json.dump(report, stream, ensure_ascii=False, indent=2)
            stream.write("\n")
    print("原版正常選人取消 PASS", len(report["states"]), len(report["actual_irq1_events"]), len(report["artifacts"]), digest(args.receipt))


if __name__ == "__main__":
    main()
