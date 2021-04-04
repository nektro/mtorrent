#!/usr/bin/env bash

build_template() {
    export CGO_ENABLED=0
    export GOOS=$1
    export GOARCH=$2
    EXT=$3
    TAG=$(date +'%Y%m%d.%H%M')
    echo $TAG-$GOOS-$GOARCH
    go build -ldflags="-s -w" -o ./bin/mtorrent-$GOOS-$GOARCH$EXT
    # -$TAG-$GOOS-$GOARCH$EXT
}

#
# build_template aix ppc64
# build_template android 386
# build_template android amd64
# build_template android arm
# build_template android arm64
build_template darwin amd64
build_template darwin arm64
build_template dragonfly amd64
build_template freebsd 386
build_template freebsd amd64
build_template freebsd arm
build_template freebsd arm64
build_template illumos amd64
# build_template ios amd64
# build_template ios arm64
# build_template js wasm
build_template linux 386
build_template linux amd64
build_template linux arm
build_template linux arm64
build_template linux mips
build_template linux mips64
build_template linux mips64le
build_template linux mipsle
build_template linux ppc64
build_template linux ppc64le
build_template linux riscv64
build_template linux s390x
build_template netbsd 386
build_template netbsd amd64
build_template netbsd arm
build_template netbsd arm64
build_template openbsd 386
build_template openbsd amd64
build_template openbsd arm
build_template openbsd arm64
build_template openbsd mips64
# build_template plan9 386
# build_template plan9 amd64
# build_template plan9 arm
build_template solaris amd64
build_template windows 386 .exe
build_template windows amd64 .exe
build_template windows arm .exe
