cd /D "%~dp0"

if "%~1"=="" goto blank

rem Parameter 1 is path to cl.exe
%1 /LD main.cpp          
del main.obj
del main.exp
del main.lib
del ..\deploy\DEVOBJ.dll
move main.dll ..\deploy\DEVOBJ.dll
goto EOF

:blank
echo ".\build.bat <path-to-cl.exe>"
echo Specify path to cl.exe utility (provided by Visual Studio) for compilation

:EOF