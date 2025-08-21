package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type buildTarget struct {
	GOOS       string
	GOARCH     string
	OutputDir  string
	OutputName string
}

const sourcePath = "./cmd/main.go"

func main() {
	targets := []buildTarget{
		{
			GOOS:       "windows",
			GOARCH:     "amd64",
			OutputDir:  "./build",
			OutputName: "windows_mtc.exe",
		},
		{
			GOOS:       "linux",
			GOARCH:     "amd64",
			OutputDir:  "./build",
			OutputName: "linux_mtc",
		},
		{
			GOOS:       "darwin",
			GOARCH:     "amd64",
			OutputDir:  "./build",
			OutputName: "macos_mtc",
		},
	}

	log.Println("Starting build process...")

	for _, target := range targets {
		log.Printf("Building for %s/%s...", target.GOOS, target.GOARCH)

		// 1. Создаем директорию для сборки, если она не существует
		if err := os.MkdirAll(target.OutputDir, os.ModePerm); err != nil {
			log.Fatalf("Failed to create directory %s: %v", target.OutputDir, err)
		}

		// 2. Формируем полный путь к выходному файлу
		outputPath := filepath.Join(target.OutputDir, target.OutputName)

		// 3. Создаем команду для сборки
		cmd := exec.Command("go", "build", "-o", outputPath, sourcePath)

		// 4. Устанавливаем переменные окружения для кросс-компиляции
		cmd.Env = append(os.Environ(),
			fmt.Sprintf("GOOS=%s", target.GOOS),
			fmt.Sprintf("GOARCH=%s", target.GOARCH),
		)

		// 5. Выполняем борку
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("ERROR building for %s/%s.", target.GOOS, target.GOARCH)
			log.Fatalf("Command failed with error: %v\nOutput:\n%s", err, string(output))
		}

		log.Printf("Successfully built: %s", outputPath)
	}

	log.Println("All builds completed successfully!")
}
