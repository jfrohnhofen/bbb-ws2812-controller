#!/usr/bin/env bash
set -ex

BUILD_DIR="build"
PASM_DIR="third_party/am335x_pru_package/pru_sw/utils/pasm_source"

rm -rf ${BUILD_DIR}
mkdir ${BUILD_DIR}

gcc -Wall -Werror -D_UNIX_ -o ${BUILD_DIR}/pasm ${PASM_DIR}/pasm.c ${PASM_DIR}/pasmpp.c ${PASM_DIR}/pasmexp.c ${PASM_DIR}/pasmop.c ${PASM_DIR}/pasmdot.c ${PASM_DIR}/pasmstruct.c ${PASM_DIR}/pasmmacro.c ${PASM_DIR}/path_utils.c

go generate ./cmd/controller
env GOOS=linux GOARCH=arm go build -o ${BUILD_DIR}/controller ./cmd/controller

go build -o ${BUILD_DIR}/compiler ./cmd/compiler
