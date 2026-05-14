# Checker Go

Готовая Go-версия чекера для Windows. Программа проверяет локальные артефакты на Minecraft cheat/client индикаторы, создает TXT и JSON отчеты и может отправить результат в Telegram.

Программа ничего не удаляет, не чистит и не меняет в системе.

## Запуск

Скачать готовый exe с GitHub Releases и сразу запустить focused-проверку:

```powershell
$u="https://github.com/Man4osik/Checker-Go/releases/latest/download/checker_go.exe"; $p="$env:TEMP\checker_go.exe"; curl.exe -L -f --retry 3 -o $p $u; if ($LASTEXITCODE -eq 0) { & $p }
```

Скачать и запустить с Telegram:

```powershell
$u="https://github.com/Man4osik/Checker-Go/releases/latest/download/checker_go.exe"; $p="$env:TEMP\checker_go.exe"; curl.exe -L -f --retry 3 -o $p $u; if ($LASTEXITCODE -eq 0) { & $p --telegram }
```

Скачать и запустить с Telegram и broad-режимом:

```powershell
$u="https://github.com/Man4osik/Checker-Go/releases/latest/download/checker_go.exe"; $p="$env:TEMP\checker_go.exe"; curl.exe -L -f --retry 3 -o $p $u; if ($LASTEXITCODE -eq 0) { & $p --telegram --broad }
```

Скачать и запустить максимальную проверку с Telegram, broad и deep:

```powershell
$u="https://github.com/Man4osik/Checker-Go/releases/latest/download/checker_go.exe"; $p="$env:TEMP\checker_go.exe"; curl.exe -L -f --retry 3 -o $p $u; if ($LASTEXITCODE -eq 0) { & $p --telegram --broad --deep }
```

Скачать и сохранить отчеты в выбранную папку:

```powershell
$u="https://github.com/Man4osik/Checker-Go/releases/latest/download/checker_go.exe"; $p="$env:TEMP\checker_go.exe"; curl.exe -L -f --retry 3 -o $p $u; if ($LASTEXITCODE -eq 0) { & $p --out "C:\Reports" }
```

Если репозиторий будет называться иначе, замени `Man4osik/Checker-Go` в ссылках на свой `username/repository`.

Важно: ссылка заработает только после того, как в GitHub Releases будет опубликован файл с точным именем `checker_go.exe`.

## Что Проверяет

- Запущенные процессы.
- Важные Windows-службы для screenshare-проверок.
- Автозагрузку в реестре и Startup-папках.
- Запланированные задачи.
- DNS cache.
- Файлы в Desktop, Downloads, Documents, AppData, ProgramData и Program Files.
- Minecraft-папки, лаунчеры, логи и конфиги.
- Подозрительные файлы в TEMP.
- Browser artifacts Chrome, Edge и Firefox.
- Recent и Prefetch.
- Manual-файлы на Desktop как контекст, не как доказательство.

## Режимы

- `focused` - режим по умолчанию, меньше шума и false positive.
- `--broad` - расширенный поиск, больше детектов и больше false positive.
- `--deep` - проверяет больше текстовых файлов и более крупные логи.

## Флаги

- `--telegram` - отправить краткий и полный отчет в Telegram.
- `--broad` - включить широкий шумный поиск.
- `--deep` - включить более глубокую проверку логов и текстовых файлов.
- `--days 30` - окно датированных находок.
- `--max-files 70000` - лимит обхода файлов.
- `--max-ms 180000` - мягкий лимит времени проверки в миллисекундах.
- `--out C:\Path\Reports` - папка для отчетов.
- `--json=false` - не создавать JSON-отчет.

## Отчеты

По умолчанию отчеты сохраняются в `%TEMP%`:

- `scan_summary_go_*.txt` - краткий отчет.
- `scan_report_go_*.txt` - полный TXT-отчет.
- `scan_report_go_*.json` - структурированный JSON-отчет.

## Уровни Риска

- `high` - сильный индикатор чита, клиента или домена.
- `medium` - подозрительный индикатор, требует ручной проверки.
- `low` - слабый или шумный индикатор.
- `info` - только контекст, не детект.

Находки не являются автоматическим доказательством. Смотри `risk`, путь файла, дату изменения и источник.

## Telegram

В готовом exe уже есть fallback Telegram-настройки. Их можно переопределить через переменные окружения:

```powershell
$env:TELEGRAM_BOT_TOKEN="your_bot_token"
$env:TELEGRAM_CHAT_ID="your_chat_id"
.\checker_go.exe --telegram
```

## Исходники

Исходный код лежит в файле `checker_go.go`.

Описание исходников: `SOURCES.md`.
