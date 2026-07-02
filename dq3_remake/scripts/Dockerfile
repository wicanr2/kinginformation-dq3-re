# dq3_remake build/驗證環境(不污染 host)。SDL2 + cmake + Xvfb + 截圖工具。
#   docker build -t dq3-remake dq3_remake/scripts
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
        build-essential cmake pkg-config libsdl2-dev \
        libvorbis-dev libogg-dev \
        xvfb x11-apps imagemagick netpbm \
    && rm -rf /var/lib/apt/lists/*
# libvorbis-dev:CMake 偵測到 → 定義 DQ3_HAVE_VORBIS → remake 直接解 .ogg 播 MT-32(否則退 SB)。
# 烤進 image(取代每次 build 臨時 apt,修 flaky;Tier 3 dev-setup)。
WORKDIR /src
CMD ["bash"]
