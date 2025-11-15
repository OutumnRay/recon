# Шпаргалка по версионированию

## 🚀 Основные команды

| Команда | Описание | Результат |
|---------|----------|-----------|
| `./build.sh` | Сборка с автоувеличением BUILD | `0.1.0+5` → `0.1.0+6` |
| `./version-bump.sh major` | Увеличить мажорную версию | `0.1.0+5` → `1.0.0+0` |
| `./version-bump.sh minor` | Увеличить минорную версию | `0.1.0+5` → `0.2.0+0` |
| `./version-bump.sh patch` | Увеличить патч версию | `0.1.0+5` → `0.1.1+0` |

## 📁 Ключевые файлы

| Файл | Назначение | Редактировать? |
|------|------------|----------------|
| `VERSION` | Текущая версия проекта | ✅ Да (или через скрипт) |
| `build.sh` | Скрипт сборки | ⚙️ При необходимости |
| `version-bump.sh` | Утилита изменения версии | ⚙️ При необходимости |
| `front/*/src/config/version.ts` | Версия для frontend | ❌ Автогенерация |
| `internal/version/version.go` | Версия для backend | ❌ Автогенерация |

## 🔄 Типичные сценарии

### Обычная разработка
```bash
./build.sh
```

### Релиз новой фичи (minor)
```bash
./version-bump.sh minor
git add VERSION
git commit -m "chore: bump version to $(cat VERSION)"
git tag -a "v$(cat VERSION)" -m "Release $(cat VERSION)"
./build.sh
```

### Исправление бага (patch)
```bash
./version-bump.sh patch
./build.sh
```

### Мажорный релиз
```bash
./version-bump.sh major
./build.sh
```

## 💡 Быстрые факты

- ✅ `build.sh` всегда увеличивает только номер сборки (+1)
- ✅ Версия встраивается в Docker образы автоматически
- ✅ Версия отображается в UI обоих порталов
- ❌ Никогда не редактируйте `version.ts` и `version.go` вручную
- 📝 Закоммитьте `VERSION` после изменения

## 🎯 Формат версии

```
MAJOR.MINOR.PATCH+BUILD
  │     │     │     └─── Автоувеличивается при ./build.sh
  │     │     └───────── Исправления ошибок (./version-bump.sh patch)
  │     └─────────────── Новые функции (./version-bump.sh minor)
  └───────────────────── Несовместимые изменения (./version-bump.sh major)
```

## 🔍 Проверка версии

### В файле
```bash
cat VERSION
```

### В коде (Frontend)
```typescript
import { APP_VERSION } from './config/version';
console.log(APP_VERSION);
```

### В коде (Backend)
```go
import "Recontext.online/internal/version"
fmt.Println(version.GetVersion())
```
