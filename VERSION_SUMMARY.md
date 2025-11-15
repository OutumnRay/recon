# Сводка системы версионирования

## ✅ Что было реализовано

### 1. Автоматическое управление версиями
- Формат версии: `MAJOR.MINOR.PATCH+BUILD` (например, `0.1.0+42`)
- Файл `VERSION` содержит текущую версию проекта
- Номер сборки (`BUILD`) увеличивается автоматически при каждом запуске `./build.sh`

### 2. Скрипты
- **`build.sh`** - главный скрипт сборки:
  - Увеличивает номер сборки
  - Генерирует файлы версий для frontend и backend
  - Запускает `docker compose up --build`

- **`version-bump.sh`** - утилита для изменения версии:
  - Поддержка: `major`, `minor`, `patch`, `build`

### 3. Интеграция версии

#### Frontend (TypeScript)
Автогенерируемые файлы:
- `front/user-portal/src/config/version.ts`
- `front/managing-portal/src/config/version.ts`

Содержат:
```typescript
export const APP_VERSION = '0.1.0+42';
export const BUILD_DATE = '2025-11-15 12:00:00 UTC';
```

Использование в коде:
```typescript
import { APP_VERSION } from './config/version';
<div>v{APP_VERSION}</div>
```

#### Backend (Go)
Автогенерируемый файл:
- `internal/version/version.go`

Содержит:
```go
const Version = "0.1.0+42"
const BuildDate = "..."
const Major = "0"
const Minor = "1"
const Patch = "0"
const Build = "42"

func GetVersion() string { return Version }
func GetBuildDate() string { return BuildDate }
```

Использование в коде:
```go
import "Recontext.online/internal/version"
fmt.Printf("Version: %s\n", version.GetVersion())
```

#### Docker
- Docker образы используют тег `latest`
- Версия передается в контейнеры через переменную окружения `APP_VERSION`
- Версия жестко встроена в код через автогенерируемые файлы

### 4. Git интеграция
Добавлено в `.gitignore`:
```
# Auto-generated version files
front/user-portal/src/config/version.ts
front/managing-portal/src/config/version.ts
internal/version/version.go
```

**Важно:** Коммитить нужно только файл `VERSION`!

## 🎯 Быстрый старт

```bash
# Обычная сборка
./build.sh

# Релиз новой версии
./version-bump.sh minor
./build.sh
```

## 📚 Документация

- **[QUICKSTART.md](QUICKSTART.md)** - быстрый старт
- **[BUILD_INSTRUCTIONS.md](BUILD_INSTRUCTIONS.md)** - инструкции по сборке
- **[VERSION.md](VERSION.md)** - полная документация
- **[VERSION_CHEATSHEET.md](VERSION_CHEATSHEET.md)** - шпаргалка команд

## 🔍 Где отображается версия

1. **Frontend UI** - в боковой панели Dashboard: `v0.1.0+42`
2. **Backend логи** - можно добавить вывод версии при старте
3. **Переменные окружения** - `APP_VERSION` в контейнерах
4. **Код приложения** - доступно через импорт модулей

## ⚠️ Важные моменты

✅ **Редактировать можно:**
- `VERSION` - текущая версия проекта
- `build.sh` - при необходимости изменения процесса сборки
- `version-bump.sh` - при необходимости

❌ **Никогда не редактируйте вручную:**
- `front/*/src/config/version.ts`
- `internal/version/version.go`

Эти файлы генерируются автоматически!

## 🔄 Workflow

### Ежедневная разработка
```bash
./build.sh
```
Версия: `0.1.0+1` → `0.1.0+2` → `0.1.0+3` ...

### Релиз новой фичи
```bash
./version-bump.sh minor
git add VERSION
git commit -m "chore: bump version to $(cat VERSION)"
./build.sh
```
Версия: `0.1.0+42` → `0.2.0+0`

### Исправление бага
```bash
./version-bump.sh patch
./build.sh
```
Версия: `0.1.0+42` → `0.1.1+0`

### Мажорный релиз
```bash
./version-bump.sh major
./build.sh
```
Версия: `0.1.0+42` → `1.0.0+0`

## 📦 Структура файлов

```
.
├── VERSION                           # Текущая версия (редактируется)
├── build.sh                          # Скрипт сборки
├── version-bump.sh                   # Утилита изменения версии
├── docker-compose.yml                # Обновлен для передачи VERSION
│
├── front/
│   ├── user-portal/src/config/
│   │   └── version.ts               # ❌ Автогенерация
│   └── managing-portal/src/config/
│       └── version.ts               # ❌ Автогенерация
│
├── internal/
│   └── version/
│       └── version.go               # ❌ Автогенерация
│
└── docs/
    ├── QUICKSTART.md
    ├── BUILD_INSTRUCTIONS.md
    ├── VERSION.md
    ├── VERSION_CHEATSHEET.md
    └── VERSION_SUMMARY.md           # Этот файл
```

## ✨ Преимущества

1. **Автоматизация** - версия увеличивается автоматически
2. **Согласованность** - одна версия для frontend и backend
3. **Отслеживаемость** - легко понять, какая версия запущена
4. **Простота** - один скрипт для всего процесса
5. **Git-friendly** - минимум файлов для коммита

---

**Текущая версия проекта:** `0.1.0+0`

Для начала работы просто запустите: `./build.sh`
