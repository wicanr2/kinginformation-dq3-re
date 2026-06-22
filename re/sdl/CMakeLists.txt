# DQ3 (精訊) SDL2 port — 跨平台 build(Linux / Windows / macOS)。
# docker 內驗證:見 tools/build/Dockerfile.sdl + sdl_build.sh。
cmake_minimum_required(VERSION 3.10)
project(dq3_sdl C)

set(CMAKE_C_STANDARD 99)
set(CMAKE_C_STANDARD_REQUIRED ON)

find_package(SDL2 REQUIRED)

add_executable(dq3_sdl
    sdl_main.c
    dq3_runtime.c
    dq3_pcx.c
)

target_include_directories(dq3_sdl PRIVATE ${SDL2_INCLUDE_DIRS})
target_link_libraries(dq3_sdl PRIVATE ${SDL2_LIBRARIES})

# 部分平台 SDL2 用 imported target 而非變數
if(TARGET SDL2::SDL2)
    target_link_libraries(dq3_sdl PRIVATE SDL2::SDL2)
endif()
if(TARGET SDL2::SDL2main)
    target_link_libraries(dq3_sdl PRIVATE SDL2::SDL2main)
endif()
