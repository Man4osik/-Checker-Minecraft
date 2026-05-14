# Checker Go

Fast standalone Windows checker written in Go. It scans local system artifacts for Minecraft cheat/client indicators and creates TXT + JSON reports. The program does not delete files, clean traces or change system settings.

## What It Checks

- Running processes via WMI/PowerShell fallback.
- Windows services used in screenshare-style checks: `PcaSvc`, `CDPSvc`, `DPS`, `SSDPSRV`, `DiagTrack`, `SysMain`, `EventLog`, `Appinfo`, `WSearch`, `DusmSvc`.
- Startup registry keys and Startup folders.
- Scheduled tasks.
- DNS cache.
- Focused files in Desktop, Downloads, Documents, AppData, ProgramData and Program Files.
- Minecraft folders, launcher folders, logs and configs.
- TEMP suspicious executable/archive files.
- Browser artifacts from Chrome, Edge and Firefox.
- Recent and Prefetch.
- Manual indicator files on Desktop, treated as context only, not evidence.

## Risk Levels

- `high` - strong cheat/client/domain indicator.
- `medium` - suspicious indicator that needs manual review.
- `low` - weak/noisy indicator, launcher/proxy/generic match or service state.
- `info` - context only, not a confirmed detection.

Findings are indicators, not automatic proof. Always check risk level, file path, modified date and source.

## Download And Run

If you downloaded a ready `checker_go.exe`, Go is not required.

Focused scan:

```powershell
.\checker_go.exe
```

Telegram:

```powershell
.\checker_go.exe --telegram
```

Telegram + broad mode:

```powershell
.\checker_go.exe --telegram --broad
```

Deep scan with Telegram + broad mode:

```powershell
.\checker_go.exe --telegram --broad --deep
```

Write reports to a specific folder:

```powershell
.\checker_go.exe --out "C:\Reports"
```

## Build From Source

Go is required only if you build from source.

```powershell
go build -o checker_go.exe .\checker_go.go
```

Build optimized Windows executable:

```powershell
go build -trimpath -ldflags="-s -w" -o checker_go.exe .\checker_go.go
```

## Flags

- `--telegram` - send short summary and full TXT report to Telegram.
- `--broad` - wider noisy search. More detections, more false positives.
- `--deep` - scan more text files and larger logs.
- `--days 30` - dated artifact window.
- `--max-files 70000` - file enumeration limit.
- `--max-ms 180000` - soft scan timeout in milliseconds.
- `--out C:\Path\Reports` - report output directory.
- `--json=false` - disable JSON report.

## Reports

By default reports are written to `%TEMP%`:

- `scan_summary_go_*.txt` - short summary.
- `scan_report_go_*.txt` - full readable report.
- `scan_report_go_*.json` - structured report for parsing or bots.

## Telegram

The current build contains fallback Telegram settings. You can override them with environment variables:

```powershell
$env:TELEGRAM_BOT_TOKEN="your_bot_token"
$env:TELEGRAM_CHAT_ID="your_chat_id"
.\checker_go.exe --telegram
```

## Modes

Focused mode is the default and tries to reduce false positives.

Broad mode enables additional weak/generic indicators. Use it when you want more coverage and are ready to review noise manually.

## Notes

The Go version is faster and cleaner than the WSH JavaScript version. The older PowerShell full scanner can still provide deeper forensic sources such as SRUDB, ShellBags and Everything/ES checks.
