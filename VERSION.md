# Система управления версиями

Этот проект использует автоматическую систему версионирования с форматом версии: `MAJOR.MINOR.PATCH+BUILD`

## Формат версии

- **MAJOR** - мажорная версия (несовместимые изменения API)
- **MINOR** - минорная версия (новая функциональность с обратной совместимостью)
- **PATCH** - патч версия (исправления ошибок)
- **BUILD** - номер сборки (автоматически увеличивается при каждой сборке)

Пример: `0.1.0+42` означает версию 0.1.0, сборка номер 42

## Файлы системы версионирования

- **VERSION** - основной файл с текущей версией проекта
- **build.sh** - главный скрипт сборки (увеличивает BUILD и запускает Docker Compose)
- **version-bump.sh** - утилита для ручного изменения версии
- **front/user-portal/src/config/version.ts** - автогенерируемый файл версии для frontend
- **front/managing-portal/src/config/version.ts** - автогенерируемый файл версии для managing portal
- **internal/version/version.go** - автогенерируемый файл версии для backend

## Использование

### Обычная сборка (автоматическое увеличение BUILD)

```bash
./build.sh
```

Этот скрипт:
1. Читает текущую версию из файла `VERSION`
2. Увеличивает номер сборки (BUILD) на 1
3. Генерирует файлы версий для frontend и backend
4. Запускает `docker compose -f docker-compose.yml up --build`

### Ручное изменение версии

Для увеличения MAJOR версии (например, 0.1.0+5 → 1.0.0+0):
```bash
./version-bump.sh major
```

Для увеличения MINOR версии (например, 0.1.0+5 → 0.2.0+0):
```bash
./version-bump.sh minor
```

Для увеличения PATCH версии (например, 0.1.0+5 → 0.1.1+0):
```bash
./version-bump.sh patch
```

Для увеличения только BUILD версии (например, 0.1.0+5 → 0.1.0+6):
```bash
./version-bump.sh build
# или просто
./version-bump.sh
```

После изменения версии вручную запустите сборку:
```bash
./build.sh
```

## Использование версии в коде

### Frontend (TypeScript/React)

```typescript
import { APP_VERSION, BUILD_DATE } from './config/version';

console.log(`Version: ${APP_VERSION}`);
console.log(`Build date: ${BUILD_DATE}`);
```

В компонентах:
```tsx
import { APP_VERSION } from '../config/version';

<div className="version">v{APP_VERSION}</div>
```

### Backend (Go)

```go
import "Recontext.online/internal/version"

func main() {
    fmt.Printf("Version: %s\n", version.GetVersion())
    fmt.Printf("Build date: %s\n", version.GetBuildDate())

    // Или используйте отдельные компоненты
    fmt.Printf("Major: %s, Minor: %s, Patch: %s, Build: %s\n",
        version.Major, version.Minor, version.Patch, version.Build)
}
```

## Docker образы и версия

Docker образы остаются с тегом `latest`, но версия приложения:
- Передается как переменная окружения `APP_VERSION` в контейнеры
- Жестко встраивается в код frontend и backend через автогенерируемые файлы
- Доступна в runtime через `version.GetVersion()` (Go) или `APP_VERSION` (TypeScript)

## Примеры использования

### Релиз новой минорной версии

```bash
# Увеличиваем минорную версию
./version-bump.sh minor
# VERSION теперь содержит 0.2.0+0

# Собираем и деплоим
./build.sh
```

### Исправление ошибки (hotfix)

```bash
# Увеличиваем патч версию
./version-bump.sh patch
# VERSION теперь содержит 0.1.1+0

# Собираем и деплоим
./build.sh
```

### Обычная разработка

```bash
# Просто запускаем сборку - BUILD увеличится автоматически
./build.sh
```

## Важные замечания

⚠️ **НЕ редактируйте вручную** следующие файлы - они генерируются автоматически:
- `front/user-portal/src/config/version.ts`
- `front/managing-portal/src/config/version.ts`
- `internal/version/version.go`

✅ **Редактировать можно** только файл `VERSION` (или используйте `version-bump.sh`)

## Continuous Integration

Для CI/CD систем можно использовать:

```bash
# Установить конкретную версию
echo "1.0.0+0" > VERSION

# Или увеличить версию программно
./version-bump.sh patch

# Запустить сборку
./build.sh
```

## Workflow для команды

1. **Фича-разработка**: Просто используйте `./build.sh` - BUILD будет увеличиваться
2. **Перед релизом**: Используйте `./version-bump.sh minor` или `./version-bump.sh major`
3. **Hotfix**: Используйте `./version-bump.sh patch`
4. **Коммит**: Закоммитьте файл `VERSION` после изменения версии

```bash
# Пример workflow для релиза
./version-bump.sh minor
git add VERSION
git commit -m "chore: bump version to $(cat VERSION)"
git tag -a "v$(cat VERSION)" -m "Release version $(cat VERSION)"
./build.sh
```
