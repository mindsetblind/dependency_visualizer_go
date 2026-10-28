// Command depviz визуализирует граф зависимостей пакетов Python.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"depviz/src/config"
)

// exitError — код выхода программы при ошибке.
const exitError = 1

// main разбирает параметры и выводит их в формате ключ=значение.
func main() {
	cfg, err := config.Parse(os.Args[1:])
	if errors.Is(err, flag.ErrHelp) {
		fmt.Println("Использование: ./run.sh -package <имя> [параметры]")
		config.PrintUsage(os.Stdout)
		return
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "ошибка:", err)
		os.Exit(exitError)
	}
	cfg.Print(os.Stdout)
}
