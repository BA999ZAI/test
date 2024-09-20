cmake_minimum_required(VERSION 3.10)
project(test)

set(GO_EXECUTABLE "/usr/bin/go")

set(SOURCE_DIR "${CMAKE_SOURCE_DIR}")

add_custom_target(test ALL
    COMMAND ${GO_EXECUTABLE} build -o ${CMAKE_BINARY_DIR}/test ${SOURCE_DIR}/cmd/main.go
    WORKING_DIRECTORY ${SOURCE_DIR}
    COMMENT "Building Go application"
)

set_target_properties(test PROPERTIES
    OUTPUT_NAME "test"
    RUNTIME_OUTPUT_DIRECTORY "${CMAKE_BINARY_DIR}"
)