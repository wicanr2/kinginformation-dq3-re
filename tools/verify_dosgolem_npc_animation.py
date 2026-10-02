"""核對 dosgolem 自行重生的 NPC 動畫觀測；不宣稱 remake 畫面通過。

在 Docker 內執行。入口、原始位址及證據限制見 docs/188。
"""

import argparse
from bisect import bisect_right
import json
from pathlib import Path

from verify_dosgolem_home_entry import fields, require, verify


def verify_animation(receipt, assets):
    report = verify(receipt, assets)
    source = json.loads(receipt.read_text())
    require(source["scenario"] == "mother_home_animation", "NPC動畫情境不符")
    prefix = receipt.name.removesuffix("-receipt.json")
    lines = (receipt.parent / (prefix + ".log")).read_text().splitlines()
    require(source["npc_animation_events"] == [line for line in lines if line.startswith("DQ3_NPC_ANIMATION ")],
            "動畫事件與當次原始日誌不符")
    require(source["input_clock_events"] == [line for line in lines if line.startswith("DQ3_INPUT_CLOCK ")],
            "輸入時鐘與當次原始日誌不符")
    events = [fields(line) for line in source["npc_animation_events"]]
    require(events, "缺動畫事件")
    require(all(event["DS"] == "15ed" and event["pit_divisor"] in ("65536", "12428")
                and int(event["phase0004"]) in (0, 1) for event in events),
            "動畫位址、相位或啟動／遊戲計時參數不符")
    previous_step = -1
    pending_flip = pending_reader = None
    flips, reader_count, frozen_count = [], 0, 0
    accepted_ticks = []
    for event in events:
        step = int(event["step"])
        require(step > previous_step, "動畫事件次序不符")
        previous_step = step
        pc = event["ida_linear"]
        if pc == "1fea4":
            require(pending_flip is None and event["counter0002"] == "0", "翻轉前計數未歸零")
            require(accepted_ticks and accepted_ticks[-1]["counter0002"] == "5"
                    and accepted_ticks[-1]["phase0004"] == event["phase0004"], "翻轉缺少第五次計數")
            pending_flip = event
        elif pc == "1fea9":
            counter = int(event["counter0002"])
            if accepted_ticks:
                previous = accepted_ticks[-1]
                require(counter == (int(previous["counter0002"]) + 1) % 6, "遊戲計數未連續遞增")
                require(int(event["ticks"]) > int(previous["ticks"]), "虛擬tick未遞增")
                if counter:
                    require(previous["phase0004"] == event["phase0004"], "非第六次計數改變相位")
            else:
                require(counter == 1 and event["phase0004"] == "0", "缺少原始啟動計數原點")
            if counter == 0:
                require(pending_flip is not None and pending_flip["ticks"] == event["ticks"], "翻轉返回缺失")
                require(int(pending_flip["phase0004"]) ^ int(event["phase0004"]) == 1, "共用位元未翻轉")
                flips.append(len(accepted_ticks) + 1)
                pending_flip = None
            else:
                require(pending_flip is None, "翻轉後計數未保持歸零")
            accepted_ticks.append(event)
        elif pc == "11ed0":
            require(event["pit_divisor"] == "12428", "NPC玩家取圖計時參數不符")
            require(pending_reader is None, "NPC取圖巢狀入口尚未審查")
            pending_reader = event
        elif pc == "11ee8":
            require(pending_reader is not None, "NPC取圖入口缺失")
            base = int(pending_reader["BX"], 16)
            frame_word = int(event["BX"], 16)
            require(frame_word % 2 == 0, "frame指標索引未對齊")
            delta = frame_word // 2 - base
            frozen = int(pending_reader["DX"], 16) & 128 != 0
            require(delta in (0, 1) and (not frozen or delta == 0), "原始frame索引不符")
            if not frozen and pending_reader["ticks"] == event["ticks"]:
                require(delta == int(pending_reader["phase0004"]), "穩定tick的共用phase未消費")
            reader_count += 1
            frozen_count += int(frozen)
            pending_reader = None
    require(pending_flip is None and pending_reader is None, "動畫觀測未完整返回")
    require(len(flips) > 1 and reader_count > 0, "動畫觀測不足")
    require(all(b - a == 6 for a, b in zip(flips, flips[1:])), "共用翻轉間隔不是六次遊戲計數")
    clocks = [fields(line) for line in source["input_clock_events"]]
    require(len(clocks) == len(source["player_input"]), "正常輸入缺少虛擬時鐘")
    for clock, key in zip(clocks, source["player_input"]):
        require(int(clock["step"]) == key["queued_step"] and int(clock["scan"], 16) == int(key["scan"], 16),
                "輸入與虛擬時鐘未對應")
        require(clock["pit_divisor"] == "12428" and int(clock["phase0004"]) in (0, 1)
                and 0 <= int(clock["counter0002"]) < 6, "輸入時鐘狀態不符")
        index = bisect_right([int(e["step"]) for e in accepted_ticks], int(clock["step"])) - 1
        require(index >= 0 and all(clock[k] == accepted_ticks[index][k]
                                  for k in ("counter0002", "phase0004")), "輸入時鐘與自然遊戲計數不符")
    gaps = [{"previous_virtual_tick": int(a["ticks"]), "next_virtual_tick": int(b["ticks"]),
             "previous_step": int(a["step"]), "next_step": int(b["step"])}
            for a, b in zip(accepted_ticks, accepted_ticks[1:])
            if int(b["ticks"]) - int(a["ticks"]) != 1]
    report.update({"animation_flip_pairs": len(flips), "animation_period_accepted_ticks": 6,
                   "accepted_game_tick_events": len(accepted_ticks),
                   "virtual_tick_gaps": gaps,
                   "virtual_tick_is_game_counter": not gaps,
                   "npc_frame_reader_pairs": reader_count, "frozen_dynamic_samples": frozen_count,
                   "input_clock_events_verified": len(clocks),
                   "frame_phase_comparison": "消費原始共用phase；重製相位／modal保持仍待對拍",
                   "original_full_campaign_parity": False})
    return report


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--receipt", type=Path, required=True)
    parser.add_argument("--assets", type=Path, default=Path("/repo/assets_raw"))
    parser.add_argument("--output", type=Path)
    args = parser.parse_args()
    report = verify_animation(args.receipt, args.assets)
    text = json.dumps(report, ensure_ascii=False, indent=2) + "\n"
    if args.output:
        args.output.write_text(text)
    print(text, end="")
