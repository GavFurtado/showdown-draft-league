
cmake_minimum_required(VERSION 3.16)

project(AsciiVideoFilter LANGUAGES CXX)
set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# Output directories
set(CMAKE_RUNTIME_OUTPUT_DIRECTORY ${CMAKE_SOURCE_DIR}/bin)
set(CMAKE_LIBRARY_OUTPUT_DIRECTORY ${CMAKE_SOURCE_DIR}/bin)
set(CMAKE_ARCHIVE_OUTPUT_DIRECTORY ${CMAKE_SOURCE_DIR}/bin)

# Optimization flags
if(CMAKE_BUILD_TYPE STREQUAL "Release")
    set(CMAKE_CXX_FLAGS_RELEASE "-O3")
endif()

# Detect platform
if(UNIX AND NOT APPLE)
    message(STATUS "Using pkg-config to find FFmpeg on Unix")
    find_package(PkgConfig REQUIRED)
    pkg_check_modules(FFMPEG REQUIRED IMPORTED_TARGET
        libavformat
        libavcodec
        libavutil
        libswscale
        libswresample
    )
    set(FFMPEG_LIBS PkgConfig::FFMPEG)
elseif(WIN32)
    message(STATUS "Configuring FFmpeg manually for Windows")

    # Change these paths as per your local FFmpeg installation on Windows
    set(FFMPEG_INCLUDE_DIR "C:/ffmpeg/include")
    set(FFMPEG_LIB_DIR "C:/ffmpeg/lib")

    set(FFMPEG_LIBS
        "${FFMPEG_LIB_DIR}/avformat.lib"
        "${FFMPEG_LIB_DIR}/avcodec.lib"
        "${FFMPEG_LIB_DIR}/avutil.lib"
        "${FFMPEG_LIB_DIR}/swscale.lib"
        "${FFMPEG_LIB_DIR}/swresample.lib"
    )
endif()

# Source files
file(GLOB_RECURSE SRC_FILES CONFIGURE_DEPENDS src/*.cpp)

# Static library
add_library(AsciiVideoFilterLib STATIC ${SRC_FILES})
target_include_directories(AsciiVideoFilterLib
    PRIVATE ${CMAKE_SOURCE_DIR}/src
    PUBLIC ${CMAKE_SOURCE_DIR}/include
)

if(WIN32)
    target_include_directories(AsciiVideoFilterLib PRIVATE ${FFMPEG_INCLUDE_DIR})
else()
    target_compile_options(AsciiVideoFilterLib PRIVATE ${FFMPEG_CFLAGS_OTHER})
endif()

target_link_libraries(AsciiVideoFilterLib ${FFMPEG_LIBS})

# Executable
add_executable(AsciiVideoFilter main.cpp)
target_link_libraries(AsciiVideoFilter AsciiVideoFilterLib)

# Debug macro
target_compile_definitions(AsciiVideoFilterLib PRIVATE DEBUG)
target_compile_definitions(AsciiVideoFilter PRIVATE DEBUG)
