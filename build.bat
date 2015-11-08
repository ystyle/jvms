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
for /f "tokens=*" %%i in ('findstr /n . %GOPATH%\jvms.go ^| findstr ^JvmsVersion^| findstr ^23^') do set L=%%i
set goversion=%L:~20,-1%

IF NOT %version%==%goversion% GOTO VERSIONMISMATCH

SET DIST=%CD%\dist\%version%

REM Build the executable
echo Building JVMS for Windows
del /Q /F %GOBIN%\jvms.exe
cd %GOPATH%
go install jvms.go
cd %ORIG%


REM Clean the dist directory
del /Q /F "%DIST%"
mkdir "%DIST%"

REM Create the "noinstall" zip
echo Generating jvms-noinstall.zip
for /d %%a in (%GOBIN%) do (buildtools\zip -j -9 -r "%DIST%\jvms-noinstall.zip" "%CD%\LICENSE" "%%a\*" "%GOBIN%\java.ico")

REM Create the installer
echo Generating jvms-setup.zip
buildtools\iscc %INNOSETUP% /o%DIST%
buildtools\zip -j -9 -r "%DIST%\jvms-setup.zip" "%DIST%\jvms-setup.exe"
REM rm "%DIST%\jvms-setup.exe"
del /Q /F "%DIST%\jvms-setup.exe"
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
