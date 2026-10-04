"""依已審查範圍轉單次 MIDI 事件，不猜檔頭；入口 docs/188。只在 Docker 執行。"""
import argparse
import hashlib
import json
from pathlib import Path
import struct


def parse_events(data, start, end):
    if not 0 <= start < end <= len(data):
        raise ValueError('event range outside input')
    pos, ticks, running, events = start, 0, None, []
    def vlq():
        nonlocal pos
        value = 0
        for _ in range(4):
            if pos >= end:
                raise ValueError('truncated delta')
            b = data[pos]
            pos += 1
            value = (value << 7) | (b & 127)
            if b < 128:
                return value
        raise ValueError('delta exceeds four bytes')
    while pos < end:
        at = pos
        delta = vlq()
        ticks += delta
        if pos >= end:
            raise ValueError('missing event status')
        status = data[pos]
        if status >= 128:
            pos += 1
            if status < 240:
                running = status
        else:
            status = running
        if status is None:
            raise ValueError('running status without channel event')
        if 128 <= status < 240:
            size = 1 if status >> 4 in (12, 13) else 2
            payload = data[pos:pos + size]
            if pos + size > end or any(b >= 128 for b in payload):
                raise ValueError('invalid channel payload')
            pos += size
            message = bytes([status]) + payload
        elif status == 255:
            if pos >= end or data[pos] != 47:
                raise ValueError('unsupported meta event')
            pos += 1
            if vlq() != 0:
                raise ValueError('end marker has payload')
            message = b'\xff\x2f\x00'
        else:
            raise ValueError('unsupported system event')
        events.append({'offset': at, 'delta': delta, 'ticks': ticks,
                       'bytes': data[at:pos].hex(), 'message': message.hex()})
        if status == 255:
            if pos != end:
                raise ValueError('bytes follow end marker inside reviewed range')
            return events
    raise ValueError('missing end marker')


def encode_vlq(value):
    out = bytearray([value & 127])
    value >>= 7
    while value:
        out.insert(0, (value & 127) | 128)
        value >>= 7
    return bytes(out)


def to_smf(events, divisor, reference_hz, programs, ppq=48):
    if divisor <= 0 or reference_hz <= 0:
        raise ValueError('clock parameters must be positive')
    tempo = (divisor * 1000000 * ppq + reference_hz // 2) // reference_hz
    if not 0 < tempo < 1 << 24:
        raise ValueError('tempo outside SMF field')
    track = b'\x00\xff\x51\x03' + tempo.to_bytes(3, 'big')
    for event in events:
        message = bytes.fromhex(event['message'])
        if programs and message[0] >> 4 == 12:
            if message[1] >= len(programs):
                raise ValueError('program missing from explicit mapping')
            message = bytes([message[0], programs[message[1]]])
        track += encode_vlq(event['delta']) + message
    return b'MThd' + struct.pack('>IHHH', 6, 0, 1, ppq) + b'MTrk' + struct.pack('>I', len(track)) + track, tempo


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('input', type=Path)
    parser.add_argument('--start', type=lambda v: int(v, 0), required=True)
    parser.add_argument('--end', type=lambda v: int(v, 0), required=True)
    parser.add_argument('--expected-sha256', required=True)
    parser.add_argument('--clock-divisor', type=int, required=True)
    parser.add_argument('--reference-hz', type=int, required=True)
    parser.add_argument('--program-map', default='')
    parser.add_argument('--output', type=Path, required=True)
    parser.add_argument('--receipt', type=Path, required=True)
    args = parser.parse_args()
    if not Path('/.dockerenv').is_file():
        parser.error('run in the project Docker toolchain')
    if args.output.exists() or args.receipt.exists() or args.output == args.receipt:
        parser.error('preserve existing output and receipt')
    raw = args.input.read_bytes()
    digest = hashlib.sha256(raw).hexdigest()
    if digest != args.expected_sha256:
        parser.error('input hash differs')
    programs = [int(v) for v in args.program_map.split(',')] if args.program_map else []
    if any(not 0 <= v < 128 for v in programs):
        parser.error('invalid program mapping')
    events = parse_events(raw, args.start, args.end)
    smf, tempo = to_smf(events, args.clock_divisor, args.reference_hz, programs)
    report = {'input_path': str(args.input), 'input_size': len(raw), 'input_sha256': digest,
              'address_space': 'input file offsets', 'start': args.start, 'end': args.end,
              'event_count': len(events), 'delta_ticks': events[-1]['ticks'], 'events': events,
              'clock_divisor': args.clock_divisor, 'reference_hz': args.reference_hz,
              'program_map': programs, 'smf_ppq': 48, 'smf_tempo_us': tempo,
              'approximation': 'hardware-spec timing; configured synthesis when mapping supplied',
              'output_sha256': hashlib.sha256(smf).hexdigest(),
              'producer_sha256': hashlib.sha256(Path(__file__).read_bytes()).hexdigest()}
    with args.output.open('xb') as stream:
        stream.write(smf)
    with args.receipt.open('x') as stream:
        json.dump(report, stream, indent=2)
        stream.write('\n')
    print(json.dumps({k: v for k, v in report.items() if k != 'events'}))


if __name__ == '__main__':
    main()
