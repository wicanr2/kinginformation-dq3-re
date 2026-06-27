# CMake toolchain:Linux → Windows x86_64 交叉編譯(mingw-w64)。
#   cmake -S dq3_remake -B /build-win -DCMAKE_TOOLCHAIN_FILE=dq3_remake/cmake/mingw-w64-x86_64.cmake
# 搭配 Dockerfile.mingw(SDL2 已裝進 /usr/x86_64-w64-mingw32)。
set(CMAKE_SYSTEM_NAME Windows)
set(CMAKE_SYSTEM_PROCESSOR x86_64)

set(TOOLCHAIN_PREFIX x86_64-w64-mingw32)
set(CMAKE_C_COMPILER   ${TOOLCHAIN_PREFIX}-gcc)
set(CMAKE_CXX_COMPILER ${TOOLCHAIN_PREFIX}-g++)
set(CMAKE_RC_COMPILER  ${TOOLCHAIN_PREFIX}-windres)

# sysroot:SDL2 mingw dev 與 mingw runtime 都在此。
set(CMAKE_FIND_ROOT_PATH /usr/${TOOLCHAIN_PREFIX})
set(CMAKE_FIND_ROOT_PATH_MODE_PROGRAM NEVER)
set(CMAKE_FIND_ROOT_PATH_MODE_LIBRARY ONLY)
set(CMAKE_FIND_ROOT_PATH_MODE_INCLUDE ONLY)
set(CMAKE_FIND_ROOT_PATH_MODE_PACKAGE ONLY)

# 只靜態連 libgcc(免帶 libgcc_s_seh.dll);SDL2 仍動態連 → 打包帶 SDL2.dll。
set(CMAKE_EXE_LINKER_FLAGS_INIT "-static-libgcc")
