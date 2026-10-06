# APK Scanner

**Find out which framework an Android app was built with, and check its manifest for common security issues, straight from the APK.**

APK Scanner (`apkscan`) is a small command-line tool written in Go. You give it an APK and it tells you whether the app was built with **Flutter, React Native, Unity, Xamarin / .NET MAUI, Cordova, Capacitor, NativeScript, Godot** or plain **native Android** (Java/Kotlin), with a confidence score and the evidence behind the answer.

Knowing the technology is the first step of mobile analysis: it decides which tools to reach for next (for example, a Flutter app keeps its logic in `libapp.so`, so decompiling the Java code tells you very little).

## Features

- **Framework detection** from three kinds of evidence: file names inside the APK, class names inside `classes*.dex`, and (as a fallback) names declared in the manifest.
- **Confidence score and evidence.** Every result lists the exact signals that matched, so you can check the reasoning instead of trusting a label.
- **Split APK support.** Pass a base APK together with its `split_config.*.apk` files (or a whole folder) and they are analysed as one app.
- **Manifest analysis** (`-a`): requested permissions, `debuggable`, `allowBackup`, and exported components.
- **JSON output**, ready for scripts. The banner and errors go to stderr, so stdout stays clean JSON.
- **Offline and deterministic.** No network access at run time, nothing is decompiled or executed, and the same APK always gives the same result.
- **Single binary.** Pure Go, no cgo.

## Installation

You need **Go 1.22 or newer** and an internet connection while installing. Go downloads the dependencies (two small modules) automatically and compiles them into the binary, so the finished tool is a single standalone file with nothing else to install.

### Prerequisite: install Go

**Windows**

1. Download the `.msi` installer from <https://go.dev/dl/> and run it.
2. Open a **new** PowerShell window and check that it works:

   ```powershell
   go version
   ```

**Linux**

Some distributions ship an older Go than 1.22, so the official tarball is the safest option. Download the Linux archive for your CPU from <https://go.dev/dl/>, then:

```bash
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go*.linux-amd64.tar.gz      # use ...linux-arm64... on ARM
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile
```

Log out and back in (or run `source ~/.profile`), then check:

```bash
go version
```

**macOS**

```bash
brew install go        # or use the .pkg installer from https://go.dev/dl/
go version
```

### Option 1: Install with `go install` (recommended)

`go install` downloads, builds and places the binary in Go's `bin` folder (`~/go/bin` on Linux and macOS, `%USERPROFILE%\go\bin` on Windows, unless `GOPATH` is customized). Then move it to a system-wide folder so you can run `apkscan` from anywhere.

#### Linux and macOS

```bash
go install github.com/Esam-atef/APK_Scanner/cmd/apkscan@latest
sudo mkdir -p /usr/local/bin
sudo mv "$(go env GOPATH)/bin/apkscan" /usr/local/bin/
apkscan -h
```

#### Windows

Windows has no `/usr/local/bin`, so this creates an equivalent folder and adds it to the system `PATH`. Run PowerShell **as Administrator**:

```powershell
go install github.com/Esam-atef/APK_Scanner/cmd/apkscan@latest

$dir = "C:\Program Files\apkscan"
New-Item -ItemType Directory -Force -Path $dir | Out-Null
Move-Item -Force "$(go env GOPATH)\bin\apkscan.exe" $dir

$machinePath = [Environment]::GetEnvironmentVariable("Path", "Machine")
if ($machinePath -notlike "*$dir*") {
    [Environment]::SetEnvironmentVariable("Path", "$machinePath;$dir", "Machine")
}
```

Then **open a new terminal** (the updated `PATH` is only picked up by new sessions) and run:

```powershell
apkscan -h
```

> **Note:** Moving the binary is optional. If Go's `bin` folder is already on your `PATH`, `apkscan` works from there as is.

### Option 2: Build from source

Clone the repository, then build.

#### Windows

```powershell
git clone https://github.com/Esam-atef/APK_Scanner.git
cd APK_Scanner
go build -o apkscan.exe .\cmd\apkscan
.\apkscan.exe -h
```

*Optional:* to run `apkscan` from any folder, add the folder that contains `apkscan.exe` to your `PATH` (System Properties, Environment Variables, `Path`, Edit, New), then open a new terminal.

#### Linux and macOS

```bash
git clone https://github.com/Esam-atef/APK_Scanner.git
cd APK_Scanner
go build -o apkscan ./cmd/apkscan
./apkscan -h
```

*Optional:* install it system-wide:

```bash
sudo mkdir -p /usr/local/bin
sudo install -m 0755 apkscan /usr/local/bin/apkscan
```

### Building for another system

The tool is pure Go, so you can build for any supported system from any machine.

| Target | `GOOS` | `GOARCH` |
|---|---|---|
| Windows (Intel/AMD, ARM) | `windows` | `amd64`, `arm64` |
| Linux (Intel/AMD, ARM) | `linux` | `amd64`, `arm64` |
| macOS (Intel, Apple Silicon) | `darwin` | `amd64`, `arm64` |

```bash
# Linux / macOS shell
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o apkscan.exe ./cmd/apkscan
```

```powershell
# Windows PowerShell
$env:GOOS="linux"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"
go build -o apkscan ./cmd/apkscan
Remove-Item Env:GOOS, Env:GOARCH, Env:CGO_ENABLED     # back to normal builds
```

### Troubleshooting

| Problem | Fix |
|---|---|
| `go` is not recognized / command not found | Open a new terminal after installing Go, or add Go's `bin` folder to your `PATH`. |
| `could not import github.com/avast/apkparser` or `no required module provides package` | Run `go mod tidy` in the folder that contains `go.mod`. In VS Code, if the red underline stays, run **Go: Restart Language Server**. |
| `apkscan` is not recognized right after `go install` | Go's `bin` folder (or the folder you moved the binary to) is not on your `PATH`. Open a new terminal, or add that folder to `PATH`. |
| Errors mentioning an unsupported Go version | Upgrade Go to 1.22 or newer. |
| `unexpected argument` or a "got 2 values" error | A path contains a space. Wrap it in quotes: `-f "C:\My Apps\app.apk"`. |

## Usage

```
apkscan <flag> <value>
```

Exactly one of `-f`, `-m`, `-d`, `-a` is used per run.

| Flag | What it does |
|---|---|
| `-f, --file <apk>` | Scan a single APK file. |
| `-m, --multi <apk> <apk> ...` | Scan several APK files together as one app (a base APK plus its `split_config` files). Needs at least two files, separated by spaces. |
| `-d, --dir <folder>` | Scan every `.apk` file directly inside a folder as one app (sub-folders are not searched). The folder should contain a single app. |
| `-a, --manifest <apk>` | Analyse the `AndroidManifest.xml` of one APK (permissions, debuggable, backup, exported components). |
| `-h, --help` | Show help. |

Long flags work with one or two dashes (`-dir` and `--dir` are the same).

### Examples

```bash
apkscan -f app.apk                                          # one APK
apkscan -m base.apk split_config.arm64_v8a.apk              # a split app
apkscan -d ./my_app_folder                                  # a whole folder
apkscan -a base.apk                                         # manifest analysis
apkscan -f app.apk > result.json                            # save the JSON (the banner still shows on screen)
```

When a scan starts, the banner is printed to **stderr**:

```
+----------------------------------------------+
|  APK Scanner  v0.1.0                         |
|  Detect the framework behind an Android APK  |
+----------------------------------------------+
```

It is not shown for `-h`, for usage mistakes, or when an input path is wrong.

### Output and exit codes

- **stdout:** the JSON result.
- **stderr:** the banner, errors and hints.
- **Exit code:** `0` success, `1` the input could not be read or the scan failed, `2` the command was used incorrectly.

### Tips

- **Apps installed from the Play Store are usually split APKs.** Scanning only `base.apk` misses the native libraries (they live in `split_config.<abi>.apk`), so the result looks weak and the tool warns you. Use `-m` or `-d` with all the files.
- **Windows:** do not end a quoted folder path with a backslash (`"C:\my apps\"`). Windows treats `\"` as an escaped quote.
- **Windows PowerShell 5.1:** `>` saves files as UTF-16. Use PowerShell 7+, or `cmd /c "apkscan.exe -f app.apk > result.json"`.
- **Pulling an app from a phone with adb**, all splits at once (replace `com.example.app`):

  ```bash
  # Linux / macOS
  mkdir -p out
  adb shell pm path com.example.app | tr -d '\r' | sed 's/^package://' | while read -r p; do adb pull "$p" out/; done
  ```

  ```bat
  :: Windows cmd (use %%i and %%j inside a .bat file)
  mkdir out
  for /f "tokens=1,* delims=:" %i in ('adb shell pm path com.example.app') do adb pull "%j" out\
  ```

  Then run `apkscan -d out`.

### Example output: framework scan

```json
{
  "apk_path": "app.apk",
  "framework": "Flutter",
  "confidence": 100,
  "source": "rule-engine",
  "signals_matched": [
    "lib/*/libflutter.so → lib/arm64-v8a/libflutter.so",
    "assets/flutter_assets/ → assets/flutter_assets/a",
    "lib/*/libapp.so (Flutter AOT snapshot) → lib/arm64-v8a/libapp.so",
    "class io.flutter.* → io.flutter.embedding.android.FlutterActivity (dex)"
  ]
}
```

A native app that ships its own C/C++ library:

```json
{
  "apk_path": "native_app.apk",
  "framework": "Native",
  "confidence": 50,
  "source": "rule-engine",
  "signals_matched": [
    "classes.dex present; no known framework signals found"
  ],
  "unrecognized_native_libs": [
    "libmyapplication.so"
  ],
  "warnings": [
    "Native libraries that match no known framework were found. This may be custom JNI code, or an engine/framework that has no rules yet."
  ]
}
```

### Example output: manifest analysis (abridged)

```json
{
  "apk_path": "app.apk",
  "manifest": {
    "package": "io.appium.settings",
    "version_name": "8.4.0",
    "version_code": 198,
    "min_sdk": 26,
    "target_sdk": 35,
    "permissions": [
      "android.permission.ACCESS_BACKGROUND_LOCATION",
      "android.permission.ACCESS_COARSE_LOCATION",
      "..."
    ],
    "debuggable": true,
    "allow_backup": false,
    "exported_components": [
      {
        "type": "activity",
        "name": "io.appium.settings.Settings",
        "reason": "exported=true",
        "launcher": true,
        "actions": ["android.intent.action.MAIN"]
      },
      {
        "type": "service",
        "name": "io.appium.settings.LocationService",
        "reason": "exported=true"
      }
    ],
    "component_counts": { "activity": 2, "receiver": 10, "service": 9 },
    "network_security_config": false,
    "findings": [
      {
        "id": "debuggable",
        "message": "android:debuggable is true. In a release build this is a vulnerability: a debugger can be attached to the app and its data can be read. This tool cannot tell a debug build from a release build.",
        "note": "Reportable in a pentest; not usually accepted in bug bounty programs."
      },
      {
        "id": "exported_components",
        "message": "12 exported component(s) have no permission protection (launcher activities excluded). A component is exported when it sets exported=true or has an intent-filter; see exported_components."
      }
    ]
  }
}
```

## How it works

An APK is just a ZIP archive. APK Scanner never installs, runs or decompiles the app. It collects **evidence**, scores each known framework against it, and reports the best match.

```
APK file(s)
    |
    v
1. Collect evidence   file names | class names in classes*.dex | manifest names (fallback)
    |
    v
2. Score frameworks   each matched signal adds its weight, giving 0-100 per framework
    |
    v
3. Decide             highest score wins; if nothing matches, "Native" by elimination
    |
    v
JSON: framework, confidence, signals matched, warnings
```

### 1. Evidence

**File names.** The tool lists the paths inside the ZIP without extracting anything. Every framework has to ship its runtime with the app, and that leaves a recognisable footprint:

- Flutter ships its engine as `libflutter.so` and keeps its resources in `assets/flutter_assets/`.
- React Native keeps the JavaScript bundle at `assets/index.android.bundle`.
- Unity ships `libunity.so`, Xamarin ships `libmonodroid.so`, Cordova keeps a web app in `assets/www/`, and so on.

The ABI folder (`arm64-v8a`, `x86_64`, ...) does not matter: `lib/*/libflutter.so` matches in any of them.

**Class names inside `classes*.dex`.** Names of files can be missing or renamed, so the tool also reads the class names in the app's code. A DEX file starts with a header that points to a *type table*, and each entry in it is a class descriptor such as `Lio/flutter/embedding/engine/FlutterEngine;`. The tool walks that table (all `classes.dex`, `classes2.dex`, ... files) and checks whether any descriptor **starts with** a known framework prefix, for example `Lio/flutter/` or `Lcom/facebook/react/`.

Two details make this reliable:

- Only descriptors that *start with* the prefix count. A class such as `Lcom/example/io/flutter/Fake;` is ignored.
- Framework classes are often left un-renamed even when an app is obfuscated, because the framework's own native code looks them up by name. This is common but not guaranteed.

This evidence helps when `base.apk` is scanned without its splits (the classes are in the base even when the `.so` files are not), or when library or asset names were changed.

**Manifest names (fallback only).** If a DEX file cannot be read (see [Limitations](#limitations)), the tool reads the class names declared in `AndroidManifest.xml` instead, silently. Such evidence is labelled `(manifest)` in `signals_matched`; evidence from the DEX files is labelled `(dex)`.

**Split APKs.** When you pass several files (`-m`) or a folder (`-d`), their evidence is merged before scoring, so `libflutter.so` in one split and `flutter_assets/` in the base add up as they should.

### 2. Scoring

Each framework has a list of **signals**. File signals have a weight, and the weights of one framework add up to 100. A matching class prefix adds a **bonus**.

| Framework | File signals (weight) | Class bonus |
|---|---|---|
| Flutter | `libflutter.so` (60), `assets/flutter_assets/` (35), `libapp.so` (5) | `io.flutter.*` +40 |
| React Native | `assets/index.android.bundle` (55), `libreactnativejni.so` (30), `libhermes.so` (15) | `com.facebook.react.*` +40 |
| Xamarin / .NET MAUI | `libmonodroid.so` (40), `libxamarin-app.so` (20), `libmonosgen-2.0.so` (20), `assemblies/` (20) | `mono.*` +30 |
| Unity | `libunity.so` (60), `assets/bin/Data/` (30), `libil2cpp.so` (10) | `com.unity3d.player.*` +40 |
| Cordova | `assets/www/cordova.js` (60), `assets/www/index.html` (40) | `org.apache.cordova.*` +40 |
| Capacitor | `assets/capacitor.config.json` (40), `assets/public/index.html` (40), `assets/capacitor.plugins.json` (20) | `com.getcapacitor.*` +40 |
| NativeScript | `libNativeScript.so` (70), `assets/metadata/` (30) | `com.tns.*` +40 |
| Godot | `libgodot_android.so` (70), `*.pck` (30) | `org.godotengine.godot.*` +40 |

```
confidence = min(100, matched file weights + class bonus)
```

Examples:

| Situation | Calculation | Result |
|---|---|---|
| A normal Flutter release | 60 + 35 + 5 = 100, bonus not needed | **100%** |
| Only Flutter's `base.apk` (no native libraries) | `flutter_assets` 35 + class bonus 40 | **75%**, plus a warning that `lib/` is missing |
| Libraries and assets renamed, classes intact | class bonus 40 only | **40%** |

The bonus is deliberately on top of the file signals, so apps whose files are complete are not penalised for being obfuscated, and apps whose files are missing are rescued by their class names.

### 3. Decision

1. The framework with the highest confidence wins. Ties are broken alphabetically, so results are stable.
2. If the best score is **below 20%** and the APK has a `classes.dex`, the answer is **Native** (plain Java/Kotlin), reached by elimination:
   - **70%** when the APK has no native libraries at all;
   - **50%** when it has `.so` libraries that match no rule (they are listed in `unrecognized_native_libs`);
   - never more than **40%** if any DEX file could not be read, because the code was not actually inspected.
3. If nothing matches and there is no `classes.dex`, the result is `Unknown`.

### 4. Reading the result

| Field | Meaning |
|---|---|
| `framework` | The detected technology. |
| `confidence` | How strongly the evidence points to it, from 0 to 100. This is **not a statistical probability**. It measures how much of the known footprint was found. |
| `source` | Always `rule-engine` for now. |
| `signals_matched` | Each rule that fired and the file or class it matched. |
| `other_candidates` | Other frameworks that also had some evidence. A high number here means the result is not clear-cut. |
| `unrecognized_native_libs` | `.so` libraries that match no rule (only shown for `Native`). Custom JNI code, or an engine the tool does not know. |
| `warnings` | Reasons the result may be incomplete: a split or config APK without code, native libraries missing because splits were not included, unreadable or packed DEX files. |

`Native` is an answer by elimination, not by evidence. It means "none of the known footprints were found".

### Manifest analysis (`-a`)

`AndroidManifest.xml` inside an APK is stored as binary XML. The tool reads it with the [`avast/apkparser`](https://github.com/avast/apkparser) library and reports:

1. **Permissions** requested by the app.
2. **`android:debuggable`.** A `true` value gives a finding. In a release build this is a vulnerability; the tool cannot tell a debug build from a release build, so the decision is yours.
3. **`android:allowBackup`.** A finding when it is `true`, or when it is not set (the default is `true`).
4. **Exported components** (activities, services, broadcast receivers, content providers). A component is reported as exported when:
   - it sets `android:exported="true"` (`reason: "exported=true"`), or
   - it has an `intent-filter` and `android:exported` is not set (`reason: "intent-filter"`).

   Content providers have no intent filters, so they count only with an explicit `exported="true"`. Disabled components are skipped. Launcher activities are listed (with `"launcher": true`) but are not counted as a problem. Components protected by `android:permission` are listed with that permission and are not counted as unprotected. For each exported component the tool also lists its intent-filter `actions`, which you need to test it (for example with `adb shell am start`).

Whether to report a finding in a pentest or a bug bounty program is a judgement call. The `note` field carries a rule of thumb only.

The manifest of the **base** APK holds these settings. If you pass a split APK, the tool warns you to use `base.apk`.

## Limitations

- **Confidence is a heuristic.** The weights and thresholds are estimates, not calibrated probabilities.
- **Tested mostly on Flutter and native apps.** The rules for the other frameworks follow their known file layouts but have only been checked against synthetic APKs, so real-world edge cases may show up.
- **Unknown engines look like `Native`.** Unreal, Qt, Cocos2d and similar engines have no rules yet. Compose Multiplatform and Kotlin Multiplatform leave no distinctive footprint on Android, so they cannot be told apart from native apps.
- **Packed or heavily obfuscated apps** can hide the evidence. If a `classes*.dex` file is encrypted or invalid, the tool says so in `warnings` and lowers the confidence instead of guessing. If obfuscation renames the framework's own classes as well, class evidence is lost.
- **Split apps need all their files.** Scanning only `base.apk` gives a weaker, warned result.
- **Not supported:** `.xapk`, `.apks` and `.aab` bundles (unpack them first), and APKs whose ZIP structure is broken (they are rejected before the manifest is read). DEX files larger than 256 MiB are skipped with a warning.
- **The manifest checks are a checklist, not a full audit.** They do not look at custom permission levels, deep-link data, signing certificates, or the contents of the network security configuration (only whether one is declared).

## Project structure

```
cmd/apkscan/          command-line interface: flags, help, banner, main
internal/extractor/   opens the APK (ZIP) and reads classes*.dex
internal/dex/         reads class names from a DEX file
internal/detector/    framework rules and the scoring engine
internal/manifest/    manifest parsing and security findings
internal/output/      JSON result types and printing
```

## Development

```bash
go vet ./...
go test ./...
```

To test the manifest parser against a real APK, set `APKSCAN_TEST_APK` (PowerShell: `$env:APKSCAN_TEST_APK="C:\path\base.apk"`) and run `go test ./internal/manifest -run Real -v`.

### Adding a framework

1. Add a constant in `internal/detector/types.go`:

   ```go
   MyEngine Framework = "My Engine"
   ```

2. Add its rules in `internal/detector/rules.go`:

   ```go
   MyEngine: {
       {Name: "lib/*/libmyengine.so", Pattern: "libmyengine.so", MatchType: MatchSuffix, Weight: 70},
       {Name: "assets/myengine/", Pattern: "assets/myengine/", MatchType: MatchPrefix, Weight: 30},
       {Name: "class com.myengine.*", Pattern: "Lcom/myengine/", MatchType: MatchClassPrefix, Weight: 40},
   },
   ```

   Match types: `MatchExact` (whole path), `MatchSuffix` (path ends with), `MatchPrefix` (path starts with), `MatchContains` (path contains), `MatchClassPrefix` (a class descriptor starts with, searched in the DEX files). By convention the file-signal weights of one framework add up to 100, and class signals are bonus weights on top.

3. Run `go test ./...`.

## Dependencies and license

- [`github.com/avast/apkparser`](https://github.com/avast/apkparser) reads the binary manifest. It is released under **LGPL-3.0**, and it is used only inside `internal/manifest`, so it can be replaced without touching the rest of the code. Go links dependencies statically, so check the LGPL obligations before you distribute a compiled binary.
- `github.com/klauspost/compress` is pulled in by `apkparser`.

This project has no license file yet. Add one before publishing it.