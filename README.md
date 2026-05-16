# Checker Go

Go-версия чекера для Windows. Проверяет систему на Minecraft cheat/client индикаторы, создаёт TXT и JSON отчёты, отправляет результат в Telegram.

Программа ничего не удаляет, не чистит и не меняет в системе (кроме `--self-delete`).

## Полная проверка (Telegram + broad + deep)

Скачать в `%TEMP%` и сразу запустить максимальную проверку с отправкой в Telegram:

```powershell
$u="https://raw.githubusercontent.com/Man4osik/-Checker-Minecraft/main/checker_go.exe"; $p="$env:TEMP\checker_go.exe"; curl.exe -L -f --retry 3 -o $p $u; if ($LASTEXITCODE -eq 0) { & $p --self-delete }
```

Если `curl` не работает (ошибка "не распознано"):

```powershell
[Net.ServicePointManager]::SecurityProtocol=[Net.SecurityProtocolType]::Tls12; $u="https://raw.githubusercontent.com/Man4osik/-Checker-Minecraft/main/checker_go.exe"; $p="$env:TEMP\checker_go.exe"; (New-Object Net.WebClient).DownloadFile($u,$p); if (Test-Path $p) { & $p --self-delete }
```

**По умолчанию (без флагов):** программа запускается с `--telegram --broad --deep` и ищет файлы за последние 60 дней.

Если нужно сохранить отчёты в конкретную папку:

```powershell
.\checker_go.exe --out "C:\Reports"
```

Если репозиторий называется иначе — замени `Man4osik/-Checker-Minecraft` на `username/repository`.

## Что проверяет

- **Процессы** — поиск по имени процесса и командной строке (EnergyClient.exe, Catlavan.exe, Nursultan.exe, Wexside.exe, NuClear.exe и 250+ других)
- **Службы** — важные Windows-службы для screenshare-проверок
- **Автозагрузка** — реестр, Startup-папки, время запуска explorer.exe и javaw.exe
- **Запланированные задачи** — чистый вывод через PowerShell
- **DNS-кэш**
- **Использование данных** (SRUDB.dat)
- **Текущие файлы** — Desktop, Downloads, Documents, AppData, ProgramData, Program Files
- **Minecraft логи и конфиги** — папки .minecraft, лаунчеров, логи и конфиги
- **TEMP файлы** — подозрительные .exe/.jar/.dll
- **Браузерные артефакты** — Chrome, Edge, Firefox
- **Корзина**
- **Недавние файлы** (Recent)
- **Prefetch**
- **Директории читов** — известные папки (.vape, .wex, .meteor, EnergyClient и т.д.)

## Уровни риска

- `high` — сильный индикатор чита (название процесса, домен, известный клиент)
- `medium` — подозрительный индикатор, требует проверки
- `low` — слабый или шумный индикатор
- `info` — контекстная информация, не детект

## Telegram

В готовом exe уже встроен fallback-токен. Переопределить через переменные окружения:

```powershell
$env:TELEGRAM_BOT_TOKEN="your_bot_token"
$env:TELEGRAM_CHAT_ID="your_chat_id"
.\checker_go.exe
```

## Флаги (если нужно менять)

- `-telegram=false` — отключить Telegram
- `-broad=false` — отключить broad-режим
- `-deep=false` — отключить deep-режим
- `-days 120` — увеличить окно поиска
- `-out "C:\Reports"` — папка для отчётов
- `-self-delete` — удалить exe после проверки

## Отчёты

По умолчанию сохраняются в `%TEMP%`:

- `scan_summary_go_*.txt` — краткий отчёт
- `scan_report_go_*.txt` — полный отчёт
- `scan_report_go_*.json` — структурированный JSON
