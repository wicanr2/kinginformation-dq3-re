# dq3_remake — 精訊 DQ3 跨平台重製(C99 + SDL2)。
# 跨平台:Linux / Windows(MinGW 或 MSVC)/ macOS。
# docker build(不污染 host):見 scripts/docker_build.sh + scripts/Dockerfile。
cmake_minimum_required(VERSION 3.10)
project(dq3_remake C)

set(CMAKE_C_STANDARD 99)
set(CMAKE_C_STANDARD_REQUIRED ON)
if(NOT CMAKE_BUILD_TYPE)
    set(CMAKE_BUILD_TYPE Release)
endif()

find_package(SDL2 REQUIRED)

add_executable(dq3_remake
    src/main.c
    src/dq3_runtime.c
    src/dq3_pcx.c
    src/dq3_pal.c
    src/dq3_font.c
    src/dq3_blk.c
    src/dq3_scene.c
    src/dq3_field.c
    src/dq3_town.c
    src/dq3_stats.c
)

# 數值/升級系統單元測試(不依賴 SDL2,純邏輯;示範 #4/#5/#6 修正)
add_executable(dq3_stats_test src/dq3_stats.c src/test_stats.c)
target_include_directories(dq3_stats_test PRIVATE ${CMAKE_CURRENT_SOURCE_DIR}/include)
if(MSVC)
    target_compile_options(dq3_stats_test PRIVATE /W3)
else()
    target_compile_options(dq3_stats_test PRIVATE -Wall -Wextra)
endif()

target_include_directories(dq3_remake PRIVATE
    ${CMAKE_CURRENT_SOURCE_DIR}/include
    ${SDL2_INCLUDE_DIRS}
)
target_link_libraries(dq3_remake PRIVATE ${SDL2_LIBRARIES})

# 部分平台 SDL2 用 imported target 而非變數
if(TARGET SDL2::SDL2)
    target_link_libraries(dq3_remake PRIVATE SDL2::SDL2)
endif()
if(TARGET SDL2::SDL2main)
    target_link_libraries(dq3_remake PRIVATE SDL2::SDL2main)
endif()

if(MSVC)
    target_compile_options(dq3_remake PRIVATE /W3)
else()
    target_compile_options(dq3_remake PRIVATE -Wall -Wextra)
endif()
