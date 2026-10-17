@echo off
rem =========================================================================
rem FitGirl Repack CLI - Windows Launcher (Go Native Binary)
rem =========================================================================
setlocal EnableDelayedExpansion

cd /d "%~dp0"
title FitGirl Repack CLI

rem Check for local fitgirl-cli.exe
if exist "%~dp0fitgirl-cli.exe" (
    "%~dp0fitgirl-cli.exe" %*
    goto :Done
)

rem Check in dist folder
if exist "%~dp0dist\fitgirl-cli-windows-amd64.exe" (
    "%~dp0dist\fitgirl-cli-windows-amd64.exe" %*
    goto :Done
)

rem Check if Go is installed to build on the fly
where go >nul 2>&1
if %ERRORLEVEL% equ 0 (
    echo Compiling fitgirl-cli.exe from Go source...
    go build -ldflags="-s -w" -o "%~dp0fitgirl-cli.exe" .
    "%~dp0fitgirl-cli.exe" %*
    goto :Done
)

echo [ERROR] fitgirl-cli.exe was not found.
echo Please run: go build -o fitgirl-cli.exe .
echo.
pause
exit /b 1

:Done
set "EXIT_CODE=%ERRORLEVEL%"
if "%~1"=="" (
    if %EXIT_CODE% neq 0 (
        echo.
        echo Application exited with code %EXIT_CODE%.
        pause
    )
)
exit /b %EXIT_CODE%
