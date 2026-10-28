// Package config разбирает и проверяет параметры командной строки.
package config

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"regexp"
	"strings"
)

// DefaultRepo — адрес репозитория PyPI по умолчанию.
const DefaultRepo = "https://pypi.org/pypi"

// DefaultDepth — максимальная глубина анализа по умолчанию.
const DefaultDepth = 3

// minDepth — наименьшая допустимая глубина анализа.
const minDepth = 1

// namePattern описывает допустимое имя пакета pip.
var namePattern = regexp.MustCompile(
	`^[A-Za-z0-9]([A-Za-z0-9._-]*[A-Za-z0-9])?$`)

// Config хранит параметры, заданные пользователем.
type Config struct {
	Package   string
	Repo      string
	TestMode  bool
	ASCIITree bool
	MaxDepth  int
}

// Parse разбирает аргументы командной строки и проверяет их значения.
func Parse(args []string) (Config, error) {
	var cfg Config
	flags := newFlagSet(&cfg)
	flags.SetOutput(io.Discard)
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return Config{}, err
		}
		return Config{}, fmt.Errorf("неверные аргументы: %w", err)
	}
	if rest := flags.Args(); len(rest) > 0 {
		return Config{}, fmt.Errorf("лишние аргументы без флага: %s",
			strings.Join(rest, " "))
	}
	if err := validate(cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// PrintUsage выводит справку по всем параметрам.
func PrintUsage(w io.Writer) {
	flags := newFlagSet(&Config{})
	flags.SetOutput(w)
	flags.PrintDefaults()
}

// newFlagSet создаёт набор флагов, связанный с полями cfg.
func newFlagSet(cfg *Config) *flag.FlagSet {
	flags := flag.NewFlagSet("depviz", flag.ContinueOnError)
	flags.StringVar(&cfg.Package, "package", "",
		"имя анализируемого пакета")
	flags.StringVar(&cfg.Repo, "repo", DefaultRepo,
		"URL репозитория или путь к файлу тестового репозитория")
	flags.BoolVar(&cfg.TestMode, "test-mode", false,
		"режим работы с тестовым репозиторием")
	flags.BoolVar(&cfg.ASCIITree, "ascii-tree", false,
		"вывод зависимостей в виде ASCII-дерева")
	flags.IntVar(&cfg.MaxDepth, "max-depth", DefaultDepth,
		"максимальная глубина анализа зависимостей")
	return flags
}

// validate проверяет значения всех параметров по очереди.
func validate(cfg Config) error {
	if err := checkPackage(cfg.Package); err != nil {
		return err
	}
	if err := checkRepo(cfg.Repo, cfg.TestMode); err != nil {
		return err
	}
	return checkDepth(cfg.MaxDepth)
}

// checkPackage проверяет имя пакета.
func checkPackage(name string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("не задано имя пакета (-package)")
	}
	if !namePattern.MatchString(name) {
		return fmt.Errorf("недопустимое имя пакета %q", name)
	}
	return nil
}

// checkRepo проверяет репозиторий с учётом режима работы.
func checkRepo(repo string, testMode bool) error {
	if strings.TrimSpace(repo) == "" {
		return errors.New("не задан репозиторий (-repo)")
	}
	if testMode {
		return checkFile(repo)
	}
	return checkURL(repo)
}

// checkURL проверяет, что строка — адрес http или https.
func checkURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("неверный URL репозитория %q: %w", raw, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("URL репозитория %q должен начинаться "+
			"с http:// или https://", raw)
	}
	if u.Host == "" {
		return fmt.Errorf("в URL репозитория %q нет адреса сервера", raw)
	}
	return nil
}

// checkFile проверяет, что файл тестового репозитория существует.
func checkFile(path string) error {
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("файл тестового репозитория %q не найден", path)
	}
	if err != nil {
		return fmt.Errorf("нет доступа к файлу %q: %w", path, err)
	}
	if info.IsDir() {
		return fmt.Errorf("%q — папка, а не файл тестового репозитория",
			path)
	}
	return nil
}

// checkDepth проверяет максимальную глубину анализа.
func checkDepth(depth int) error {
	if depth < minDepth {
		return fmt.Errorf("глубина %d недопустима: должна быть не меньше %d",
			depth, minDepth)
	}
	return nil
}

// Print выводит параметры в формате ключ=значение.
func (c Config) Print(w io.Writer) {
	fmt.Fprintf(w, "package=%s\n", c.Package)
	fmt.Fprintf(w, "repo=%s\n", c.Repo)
	fmt.Fprintf(w, "test_mode=%t\n", c.TestMode)
	fmt.Fprintf(w, "ascii_tree=%t\n", c.ASCIITree)
	fmt.Fprintf(w, "max_depth=%d\n", c.MaxDepth)
}
