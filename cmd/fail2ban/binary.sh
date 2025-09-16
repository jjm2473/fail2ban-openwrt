#!/bin/sh

#BASE_DIR=/Users/luyong/go/src/github.com/linkease/quick-start/istore-backend/cmd/backend
VERSION=`egrep -om1 '\s+VERSION\s+=\s+"[0-9\.]+"' ../../api/version.go | egrep -om1 '[0-9\.]+' | head -n1`

GV=`git log --pretty=format:'%H' -n 1`
DV=`date -u +.%Y%m%d.%H%M%S`

mkdir -p build

echo "version: $VERSION"
echo "amd64"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-X main.BuildVersion='${GV}' -X main.BuildDate='${DV}' -X main.Version='${VERSION}' -s -w -extldflags "-static"' -o build/fail2ban_op_bin.amd64 || exit 1
echo "arm64"
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -a -ldflags '-X main.BuildVersion='${GV}' -X main.BuildDate='${DV}' -X main.Version='${VERSION}' -s -w -extldflags "-static"' -o build/fail2ban_op_bin.arm64 || exit 1
echo "arm"
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 GOGC=60 go build -a -ldflags '-X main.BuildVersion='${GV}' -X main.BuildDate='${DV}' -X main.Version='${VERSION}' -s -w -extldflags "-static"' -o build/fail2ban_op_bin.armv7 || exit 1

mkdir -p build/fail2banop-binary-${VERSION}

cp build/fail2ban_op_bin.arm64 ./build/fail2banop-binary-${VERSION}/fail2ban_op_bin.aarch64
cp build/fail2ban_op_bin.amd64 ./build/fail2banop-binary-${VERSION}/fail2ban_op_bin.x86_64
cp build/fail2ban_op_bin.armv7 ./build/fail2banop-binary-${VERSION}/fail2ban_op_bin.arm

tar -C build -zcvf ./build/fail2banop-binary-${VERSION}.tar.gz fail2banop-binary-${VERSION}

if command -v shasum >/dev/null 2>&1; then
    SHA=$(shasum -a 256 "./build/fail2banop-binary-${VERSION}.tar.gz" | cut -d ' ' -f 1)
else 
    SHA=$(sha256sum "./build/fail2banop-binary-${VERSION}.tar.gz" | cut -d ' ' -f 1)
fi

echo ${SHA}

