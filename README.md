# APK Scanner

Find out which technology an Android app was built with, straight from its APK.

## Features

- Detects the framework behind an APK (Flutter, React Native, Unity, native Java/Kotlin and more), with a confidence score and the evidence behind it.
- Analyzes the app's manifest: permissions, `debuggable`, `allowBackup` and exported components.

## Installation

Requires **Go 1.22 or newer**. The first build downloads the dependencies, so you need an internet connection.

There are two ways to install: **build from source** (clone the repo) or **install with `go install`** (one command, then move the binary somewhere on your `PATH`).

### Option 1: Build from source

#### Windows

```powershell
git clone https://github.com/Esam-atef/APK_Scanner.git
cd APK_Scanner
go build -o apkscan.exe .\cmd\apkscan
.\apkscan.exe -h
```

#### Linux

```bash
git clone https://github.com/Esam-atef/APK_Scanner.git
cd APK_Scanner
go build -o apkscan ./cmd/apkscan
./apkscan -h
```

### Option 2: Install with `go install`

`go install` downloads, builds and places the binary in Go's `bin` folder (`$GOPATH/bin`, which is `~/go/bin` on Linux and `%USERPROFILE%\go\bin` on Windows by default). Then move it to a system-wide folder so you can run `apkscan` from anywhere.

#### Linux

```bash
go install github.com/Esam-atef/APK_Scanner/cmd/apkscan@latest
sudo mv "$(go env GOPATH)/bin/apkscan" /usr/local/bin/
apkscan -h
```

#### Windows

Run PowerShell **as Administrator**:

```powershell
go install github.com/Esam-atef/APK_Scanner/cmd/apkscan@latest

# Windows has no /usr/local/bin, so create an equivalent folder and add it to PATH
New-Item -ItemType Directory -Force -Path "C:\Program Files\apkscan" | Out-Null
Move-Item -Force "$(go env GOPATH)\bin\apkscan.exe" "C:\Program Files\apkscan\"
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";C:\Program Files\apkscan", "Machine")
```

Then **open a new terminal** (the updated `PATH` is only picked up by new sessions) and run:

```powershell
apkscan -h
```

> **Note:** If you skip the move step, the binary still works from Go's `bin` folder as long as that folder is on your `PATH`.

## Usage

```
apkscan <flag> <value>
```

Use one flag per run.

| Flag | What it does |
|---|---|
| `-f, --file <apk>` | Scan a single APK file. |
| `-m, --multi <apk> <apk> ...` | Scan several APK files together as one app (a base APK plus its `split_config` files). Needs at least two files, separated by spaces. |
| `-d, --dir <folder>` | Scan every `.apk` file inside a folder as one app. The folder should contain a single app. |
| `-a, --manifest <apk>` | Analyze the `AndroidManifest.xml` of one APK (permissions, debuggable, backup, exported components). |
| `-h, --help` | Show help. |

### Examples

```
apkscan -f app.apk
apkscan -m base.apk split_config.arm64_v8a.apk
apkscan -d ./my_app_folder
apkscan -a base.apk
```

Run it as `.\apkscan.exe` on Windows or `./apkscan` on Linux, unless you added it to your `PATH` (this is already the case if you used Option 2 above). If a path contains spaces, wrap it in quotes.