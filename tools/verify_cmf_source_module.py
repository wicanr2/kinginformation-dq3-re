"""Compatibility CLI for complete CMF byte-layout source reconstruction; docs/25.

The shared SDK verifier defaults to CMFDRV.ASM and preserves the original
render_data function for historical validation scripts.
"""

from verify_sdk_source_module import main, render_data


if __name__ == "__main__":
    main()
