# Инструкции по сборке проекта

## Быстрый старт

### Обычная сборка проекта

```bash
./build.sh
```

Этот скрипт автоматически:
- ✅ Увеличивает номер сборки в файле `VERSION`
- ✅ Генерирует файлы версий для frontend и backend
- ✅ Запускает `docker compose -f docker-compose.yml up --build`
- ✅ Встраивает версию в Docker образы

### Изменение версии вручную

```bash
# Увеличить мажорную версию (1.0.0)
./version-bump.sh major

# Увеличить минорную версию (0.2.0)
./version-bump.sh minor

# Увеличить патч версию (0.1.1)
./version-bump.sh patch
```

После изменения версии запустите сборку:
```bash
./build.sh
```

## Формат версии

Проект использует формат: `MAJOR.MINOR.PATCH+BUILD`

Пример: `0.1.0+42`
- `0` - мажорная версия
- `1` - минорная версия
- `0` - патч версия
- `42` - номер сборки

**Номер сборки** увеличивается автоматически при каждом запуске `./build.sh`

## Где отображается версия

### Frontend
- В боковой панели Dashboard: `v0.1.0+42`
- В коде: `import { APP_VERSION } from './config/version'`

### Backend
- В коде: `import "Recontext.online/internal/version"`
- Можно использовать: `version.GetVersion()`, `version.Build`, etc.

### Docker образы
Образы используют тег `latest`, но версия встраивается в код:
```
sivanov2018/recontext-managing-portal:latest  (версия 0.1.0+42 внутри)
sivanov2018/recontext-user-portal:latest      (версия 0.1.0+42 внутри)
```

## Важно

⚠️ Не редактируйте вручную автогенерируемые файлы:
- `front/user-portal/src/config/version.ts`
- `front/managing-portal/src/config/version.ts`
- `internal/version/version.go`

Они создаются автоматически при запуске `./build.sh`

## Подробная документация

См. [VERSION.md](VERSION.md) для полной документации по системе версионирования.
