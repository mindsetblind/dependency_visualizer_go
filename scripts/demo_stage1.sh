#!/bin/bash
# Демонстрация этапа 1: вывод параметров и обработка ошибок.

cd "$(dirname "$0")/.." || exit 1

tmp_dir=$(mktemp -d demo.XXXX)
repo_file="$tmp_dir/repo.txt"
printf 'A: B\nB:\n' > "$repo_file"
trap 'rm -rf "$tmp_dir"' EXIT

demo() {
    local title=$1
    shift
    echo "=== $title"
    echo "\$ ./run.sh $*"
    ./run.sh "$@"
    echo "[код выхода: $?]"
    echo
}

echo "##### Корректные запуски #####"
echo
demo "Только имя пакета (остальное по умолчанию)" -package flask
demo "Все параметры" -package requests -repo https://pypi.org/pypi \
    -ascii-tree -max-depth 2
demo "Тестовый репозиторий" -package A -test-mode -repo "$repo_file"
demo "Справка" -h

echo "##### Ошибки: имя пакета #####"
echo
demo "Не задано имя пакета"
demo "Недопустимые символы" -package "fl@sk"

echo "##### Ошибки: репозиторий #####"
echo
demo "Не URL" -package flask -repo pypi.org
demo "Неподдерживаемая схема" -package flask -repo ftp://pypi.org
demo "Нет адреса сервера" -package flask -repo https://
demo "Тест-режим: файла нет" -package A -test-mode \
    -repo "$tmp_dir/missing.txt"
demo "Тест-режим: папка вместо файла" -package A -test-mode -repo "$tmp_dir"

echo "##### Ошибки: глубина #####"
echo
demo "Глубина 0" -package flask -max-depth 0
demo "Отрицательная глубина" -package flask -max-depth -1
demo "Глубина не число" -package flask -max-depth abc

echo "##### Ошибки: аргументы #####"
echo
demo "Неизвестный флаг" -package flask -depth 2
demo "Нет значения у флага" -package
demo "Лишний аргумент без флага" -package flask extra
