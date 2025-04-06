#!/bin/sh
set -e

echo "========== START link_brotli.sh =========="

test -d tmp/brotli || mkdir -p tmp/brotli

rm -rf vendor/github.com/google/brotli/dist/CMake*

cp -rf vendor/github.com/google/brotli/* tmp/brotli/
cp -rf submodules/brotli/* vendor/github.com/google/brotli/

readonly CWD=$PWD

cd vendor/github.com/google/brotli/go/cbrotli

readonly LIB_DIR=../../dist

echo "========== BEFORE PATCHING cgo.go =========="
cat cgo.go
echo "============================================"

# patch 1
sed -i -e "s|#cgo LDFLAGS: -lbrotlicommon|#cgo CFLAGS: -I../../c/include|" cgo.go

# patch 2
sed -i -e "s|\(#cgo LDFLAGS:\) \(-lbrotli.*\)|\1 -L$LIB_DIR \2-static -lbrotlicommon-static|" cgo.go

# patch 3: this one caused issues
echo "========== APPLYING -lm PATCH =========="
sed -i -e "/ -lm$/ n; /brotlienc/ s|$| -lm|" cgo.go

echo "========== AFTER PATCHING cgo.go =========="
cat cgo.go
echo "==========================================="

mkdir -p ${LIB_DIR}

cd ${LIB_DIR}
cmake ..
make

cd ${CWD}

echo "=========== END link_brotli.sh ============"