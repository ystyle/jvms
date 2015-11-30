@echo off
set /P JVMS_PATH="Enter the absolute path where the zip file is extracted/copied to: "
setx /M JVMS_HOME "%JVMS_PATH%"
setx /M JAVA_HOME "%SYSTEMDRIVE%\Program Files\jdk"
setx /M PATH "%PATH%;%JVMS_HOME%;%JAVA_HOME%\bin;"

if exist "%SYSTEMDRIVE%\Program Files (x86)\" (
set SYS_ARCH=64
) else (
set SYS_ARCH=32
)
(echo root: %JVMS_HOME% && echo path: %JAVA_HOME% && echo arch: %SYS_ARCH% && echo proxy: none) > %JVMS_HOME%\settings.txt

notepad %JVMS_HOME%\settings.txt
@echo on
