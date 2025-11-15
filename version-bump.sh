#!/bin/bash

# Скрипт для ручного изменения версии
# Использование: ./version-bump.sh [major|minor|patch]

set -e

VERSION_FILE="VERSION"
COLOR_GREEN='\033[0;32m'
COLOR_YELLOW='\033[1;33m'
COLOR_RED='\033[0;31m'
COLOR_RESET='\033[0m'

if [ ! -f "$VERSION_FILE" ]; then
    echo -e "${COLOR_RED}Ошибка: файл VERSION не найден!${COLOR_RESET}"
    exit 1
fi

CURRENT_VERSION=$(cat "$VERSION_FILE" | tr -d '\n\r')

if [[ $CURRENT_VERSION =~ ^([0-9]+)\.([0-9]+)\.([0-9]+)\+([0-9]+)$ ]]; then
    MAJOR="${BASH_REMATCH[1]}"
    MINOR="${BASH_REMATCH[2]}"
    PATCH="${BASH_REMATCH[3]}"
    BUILD="${BASH_REMATCH[4]}"
else
    echo -e "${COLOR_RED}Ошибка: неверный формат версии${COLOR_RESET}"
    exit 1
fi

BUMP_TYPE="${1:-build}"

case "$BUMP_TYPE" in
    major)
        MAJOR=$((MAJOR + 1))
        MINOR=0
        PATCH=0
        BUILD=0
        echo -e "${COLOR_GREEN}Увеличение мажорной версии${COLOR_RESET}"
        ;;
    minor)
        MINOR=$((MINOR + 1))
        PATCH=0
        BUILD=0
        echo -e "${COLOR_GREEN}Увеличение минорной версии${COLOR_RESET}"
        ;;
    patch)
        PATCH=$((PATCH + 1))
        BUILD=0
        echo -e "${COLOR_GREEN}Увеличение патч версии${COLOR_RESET}"
        ;;
    build)
        BUILD=$((BUILD + 1))
        echo -e "${COLOR_GREEN}Увеличение номера сборки${COLOR_RESET}"
        ;;
    *)
        echo -e "${COLOR_RED}Ошибка: неизвестный тип версии '$BUMP_TYPE'${COLOR_RESET}"
        echo "Использование: $0 [major|minor|patch|build]"
        exit 1
        ;;
esac

NEW_VERSION="${MAJOR}.${MINOR}.${PATCH}+${BUILD}"

echo -e "Старая версия: ${COLOR_YELLOW}${CURRENT_VERSION}${COLOR_RESET}"
echo -e "Новая версия:  ${COLOR_YELLOW}${NEW_VERSION}${COLOR_RESET}"

echo "$NEW_VERSION" > "$VERSION_FILE"

echo -e "${COLOR_GREEN}Версия обновлена успешно!${COLOR_RESET}"
echo "Запустите ./build.sh для сборки с новой версией"
