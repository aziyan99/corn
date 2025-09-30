@echo off
setlocal

SET "MAIN_FILE=%~dp0cmd\corn\main.go"
SET "OUTPUT_NAME=corn"


IF "%~1"=="" (
    ECHO ERROR: No build environment specified.
    ECHO.
    ECHO Usage: %~n0 [dev^|prod]
    GOTO :EOF
)

SET "BUILD_ENV=%~1"

ECHO Building for '%BUILD_ENV%' environment...

IF /I "%BUILD_ENV%"=="dev" (
    ECHO Compiling with debug symbols...
    go build -o "%~dp0build\%OUTPUT_NAME%-dev.exe" "%MAIN_FILE%"
    IF %ERRORLEVEL% EQU 0 (
        ECHO.
        ECHO Development build successful! Output: build\%OUTPUT_NAME%-dev.exe
    ) ELSE (
        ECHO.
        ECHO Development build FAILED.
    )
    GOTO :EOF
)

IF /I "%BUILD_ENV%"=="prod" (
    ECHO Compiling and optimizing for production...
    go build -ldflags="-s -w" -o "%~dp0build\%OUTPUT_NAME%.exe" "%MAIN_FILE%"
    IF %ERRORLEVEL% EQU 0 (
        ECHO.
        ECHO Production build successful! Output: build\%OUTPUT_NAME%.exe
    ) ELSE (
        ECHO.
        ECHO Production build FAILED.
    )
    GOTO :EOF
)

ECHO ERROR: Invalid argument '%BUILD_ENV%'.
ECHO.
ECHO Please use 'dev' or 'prod'.
ECHO Usage: %~n0 [dev^|prod]

:EOF
endlocal