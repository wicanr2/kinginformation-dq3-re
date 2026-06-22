# dq3_remake build/驗證環境(不污染 host)。SDL2 + cmake + Xvfb + 截圖工具。
#   docker build -t dq3-remake dq3_remake/scripts
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
        build-essential cmake pkg-config libsdl2-dev \
        xvfb x11-apps imagemagick netpbm \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /src
CMD ["bash"]
