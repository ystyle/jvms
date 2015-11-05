@echo off
SET INNOSETUP=%CD%\jvms.iss
SET ORIG=%CD%
SET GOPATH=%CD%\src
SET GOBIN=%CD%\bin
SET GOARCH=386

REM Get the version number from the setup file
for /f "tokens=*" %%i in ('findstr /n . %INNOSETUP% ^| findstr ^4:#define') do set L=%%i
set version=%L:~24,-1%

REM Get the version number from the core executable
for /f "tokens=*" %%i in ('findstr /n . %GOPATH%\jvms.go ^| findstr ^JvmsVersion^| findstr ^21^') do set L=%%i
set goversion=%L:~19,-1%

IF NOT %version%==%goversion% GOTO VERSIONMISMATCH

SET DIST=%CD%\dist\%version%

REM Build the executable
echo Building JVMS for Windows
rm %GOBIN%\jvms.exe
cd %GOPATH%
goxc -arch="386" -os="windows" -n="jvms" -d="%GOBIN%" -o="%GOBIN%\jvms{{.Ext}}" -tasks-=package
cd %ORIG%
rm %GOBIN%\src.exe
rm %GOPATH%\src.exe
rm %GOPATH%\jvms.exe

REM Clean the dist directory
rm -rf "%DIST%"
mkdir "%DIST%"

REM Create the "noinstall" zip
echo Generating jvms-noinstall.zip
for /d %%a in (%GOBIN%) do (buildtools\zip -j -9 -r "%DIST%\jvms-noinstall.zip" "%CD%\LICENSE" "%%a\*" -x "%GOBIN%\java.ico")

REM Create the installer
echo Generating jvms-setup.zip
buildtools\iscc %INNOSETUP% /o%DIST%
buildtools\zip -j -9 -r "%DIST%\jvms-setup.zip" "%DIST%\jvms-setup.exe"
REM rm "%DIST%\jvms-setup.exe"
echo --------------------------
echo Release %version% available in %DIST%
GOTO COMPLETE

:VERSIONMISMATCH
echo The version number in jvms.iss does not match the version in src\jvms.go
echo   - jvms.iss line #4: %version%
echo   - jvms.go line #21: %goversion%
EXIT /B

:COMPLETE
@echo on
