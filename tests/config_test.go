// Package tests содержит тесты пакетов проекта.
package tests

import (
	"bytes"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"depviz/src/config"
)

// errorCase описывает неверный набор аргументов и ожидаемую ошибку.
type errorCase struct {
	name    string
	args    []string
	wantErr string
}

// writeRepoFile создаёт временный файл тестового репозитория.
func writeRepoFile(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "repo.txt")
	if err := os.WriteFile(path, []byte("A: B\nB:\n"), 0o644); err != nil {
		t.Fatalf("не удалось создать файл: %v", err)
	}
	return path
}

// TestParseAllFlags проверяет разбор всех параметров сразу.
func TestParseAllFlags(t *testing.T) {
	repo := writeRepoFile(t)
	args := []string{"-package", "flask", "-repo", repo, "-test-mode",
		"-ascii-tree", "-max-depth", "2"}
	want := config.Config{Package: "flask", Repo: repo, TestMode: true,
		ASCIITree: true, MaxDepth: 2}
	got, err := config.Parse(args)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if got != want {
		t.Errorf("получили %+v, ожидали %+v", got, want)
	}
}

// TestParseDefaults проверяет значения параметров по умолчанию.
func TestParseDefaults(t *testing.T) {
	want := config.Config{Package: "flask", Repo: config.DefaultRepo,
		MaxDepth: config.DefaultDepth}
	got, err := config.Parse([]string{"-package", "flask"})
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if got != want {
		t.Errorf("получили %+v, ожидали %+v", got, want)
	}
}

// TestParseValidNames проверяет допустимые имена пакетов и URL.
func TestParseValidNames(t *testing.T) {
	cases := [][]string{
		{"-package", "zope.interface"},
		{"-package", "typing_extensions"},
		{"-package", "Flask-Login", "-repo", "http://localhost:8080/pypi"},
		{"--package=requests", "--max-depth=1"},
	}
	for _, args := range cases {
		if _, err := config.Parse(args); err != nil {
			t.Errorf("%v: неожиданная ошибка: %v", args, err)
		}
	}
}

// argErrorCases возвращает ошибки имени пакета, глубины и аргументов.
func argErrorCases() []errorCase {
	return []errorCase{
		{"нет пакета", []string{}, "не задано имя пакета"},
		{"пустой пакет", []string{"-package", " "}, "не задано имя пакета"},
		{"символы", []string{"-package", "fl@sk"}, "недопустимое имя"},
		{"дефис в начале", []string{"-package", "-x-"}, "недопустимое имя"},
		{"глубина 0", []string{"-package", "a", "-max-depth", "0"},
			"глубина 0 недопустима"},
		{"глубина -1", []string{"-package", "a", "-max-depth", "-1"},
			"глубина -1 недопустима"},
		{"глубина abc", []string{"-package", "a", "-max-depth", "abc"},
			"invalid value"},
		{"нет значения", []string{"-package"}, "needs an argument"},
		{"чужой флаг", []string{"-package", "a", "-depth", "2"},
			"not defined: -depth"},
		{"лишний аргумент", []string{"-package", "a", "extra"},
			"лишние аргументы без флага: extra"},
	}
}

// repoErrorCases возвращает ошибки параметра -repo.
func repoErrorCases(dir string) []errorCase {
	missing := filepath.Join(dir, "missing.txt")
	return []errorCase{
		{"пустой repo", []string{"-package", "a", "-repo", ""},
			"не задан репозиторий"},
		{"не URL", []string{"-package", "a", "-repo", "pypi.org"},
			"должен начинаться с http:// или https://"},
		{"схема ftp", []string{"-package", "a", "-repo", "ftp://pypi.org"},
			"должен начинаться с http:// или https://"},
		{"нет сервера", []string{"-package", "a", "-repo", "https://"},
			"нет адреса сервера"},
		{"нет файла", []string{"-package", "a", "-test-mode",
			"-repo", missing}, "не найден"},
		{"папка", []string{"-package", "a", "-test-mode", "-repo", dir},
			"папка, а не файл"},
		{"URL в тест-режиме", []string{"-package", "a", "-test-mode"},
			"не найден"},
	}
}

// TestParseErrors проверяет текст ошибки для каждого неверного случая.
func TestParseErrors(t *testing.T) {
	cases := append(argErrorCases(), repoErrorCases(t.TempDir())...)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := config.Parse(tc.args)
			if err == nil {
				t.Fatalf("ожидали ошибку %q, получили nil", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("ошибка %q не содержит %q", err, tc.wantErr)
			}
		})
	}
}

// TestParseHelp проверяет, что -h возвращает flag.ErrHelp.
func TestParseHelp(t *testing.T) {
	_, err := config.Parse([]string{"-h"})
	if !errors.Is(err, flag.ErrHelp) {
		t.Errorf("ожидали flag.ErrHelp, получили %v", err)
	}
}

// TestPrint проверяет вывод параметров в формате ключ=значение.
func TestPrint(t *testing.T) {
	cfg := config.Config{Package: "flask", Repo: config.DefaultRepo,
		ASCIITree: true, MaxDepth: 2}
	want := "package=flask\n" +
		"repo=https://pypi.org/pypi\n" +
		"test_mode=false\n" +
		"ascii_tree=true\n" +
		"max_depth=2\n"
	var buf bytes.Buffer
	cfg.Print(&buf)
	if buf.String() != want {
		t.Errorf("получили:\n%s\nожидали:\n%s", buf.String(), want)
	}
}

// TestPrintUsage проверяет, что справка упоминает все параметры.
func TestPrintUsage(t *testing.T) {
	var buf bytes.Buffer
	config.PrintUsage(&buf)
	names := []string{"-package", "-repo", "-test-mode", "-ascii-tree",
		"-max-depth"}
	for _, name := range names {
		if !strings.Contains(buf.String(), name) {
			t.Errorf("в справке нет параметра %s", name)
		}
	}
}
