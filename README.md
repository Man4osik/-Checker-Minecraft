# Checker Go

Готовая Go-версия чекера для Windows. Программа проверяет локальные артефакты на Minecraft cheat/client индикаторы, создает TXT и JSON отчеты и может отправить результат в Telegram.

Программа ничего не удаляет, не чистит и не меняет в системе. Исключение: при запуске через команды ниже временный `checker_go.exe`, скачанный в `%TEMP%`, удаляется после завершения проверки.

## Запуск

Скачать готовый exe в `%TEMP%` и сразу запустить focused-проверку:

```powershell
$u="https://raw.githubusercontent.com/Man4osik/-Checker-Minecraft/main/checker_go.exe"; $p="$env:TEMP\checker_go.exe"; curl.exe -L -f --retry 3 -o $p $u; if ($LASTEXITCODE -eq 0) { & $p --self-delete }
```

Скачать и запустить с Telegram:

```powershell
$u="https://raw.githubusercontent.com/Man4osik/-Checker-Minecraft/main/checker_go.exe"; $p="$env:TEMP\checker_go.exe"; curl.exe -L -f --retry 3 -o $p $u; if ($LASTEXITCODE -eq 0) { & $p --telegram --self-delete }
```

Скачать и запустить с Telegram и broad-режимом:

```powershell
$u="https://raw.githubusercontent.com/Man4osik/-Checker-Minecraft/main/checker_go.exe"; $p="$env:TEMP\checker_go.exe"; curl.exe -L -f --retry 3 -o $p $u; if ($LASTEXITCODE -eq 0) { & $p --telegram --broad --self-delete }
```

Скачать и запустить максимальную проверку с Telegram, broad и deep:

```powershell
$u="https://raw.githubusercontent.com/Man4osik/-Checker-Minecraft/main/checker_go.exe"; $p="$env:TEMP\checker_go.exe"; curl.exe -L -f --retry 3 -o $p $u; if ($LASTEXITCODE -eq 0) { & $p --telegram --broad --deep --self-delete }
```

Скачать и сохранить отчеты в выбранную папку:

```powershell
$u="https://raw.githubusercontent.com/Man4osik/-Checker-Minecraft/main/checker_go.exe"; $p="$env:TEMP\checker_go.exe"; curl.exe -L -f --retry 3 -o $p $u; if ($LASTEXITCODE -eq 0) { & $p --out "C:\Reports" --self-delete }
```

Если репозиторий будет называться иначе, замени `Man4osik/-Checker-Minecraft` в ссылках на свой `username/repository`.

Важно: ссылка использует raw-файл из ветки `main`, поэтому `checker_go.exe` должен лежать в корне репозитория.

Файл скачивается сюда:

```text
%TEMP%\checker_go.exe
```

## Если Команда Не Работает

Если ошибка `curl.exe не распознано`, используй PowerShell-вариант без curl:

```powershell
[Net.ServicePointManager]::SecurityProtocol=[Net.SecurityProtocolType]::Tls12; $u="https://raw.githubusercontent.com/Man4osik/-Checker-Minecraft/main/checker_go.exe"; $p="$env:TEMP\checker_go.exe"; (New-Object Net.WebClient).DownloadFile($u,$p); if (Test-Path $p) { & $p --telegram --broad --deep --self-delete }
```

Если ошибка `404`, проверь что файл реально лежит тут:

```text
https://github.com/Man4osik/-Checker-Minecraft/blob/main/checker_go.exe
```

Если Windows или антивирус блокирует запуск, можно скачать файл вручную с GitHub, затем открыть PowerShell в папке с файлом и запустить:

```powershell
.\checker_go.exe --telegram --broad --deep
```

Если PowerShell пишет, что запуск запрещен, запусти PowerShell от имени администратора или используй команду скачивания выше. Для `.exe` обычно `ExecutionPolicy` не нужен.

## Что Проверяет

- Запущенные процессы.
- Важные Windows-службы для screenshare-проверок.
- Автозагрузку в реестре и Startup-папках.
- Запланированные задачи.
- DNS cache.
- Использование данных Windows (`SRUDB.dat`) как read-only binary string scan.
- Файлы в Desktop, Downloads, Documents, AppData, ProgramData и Program Files.
- Minecraft-папки, лаунчеры, логи и конфиги.
- Подозрительные файлы в TEMP.
- Browser artifacts Chrome, Edge и Firefox.
- Recycle Bin.
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
- `--self-delete` - удалить скачанный exe после завершения проверки.

## Отчеты

По умолчанию отчеты сохраняются в `%TEMP%`:

- `scan_summary_go_*.txt` - краткий отчет.
- `scan_report_go_*.txt` - полный TXT-отчет.
- `scan_report_go_*.json` - структурированный JSON-отчет.

Даты в TXT-отчете пишутся в формате `ГГГГ.ММ.ДД ЧЧ:ММ:СС`.

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
