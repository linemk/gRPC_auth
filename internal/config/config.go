package config

import (
	"flag"                               // Модуль для парсинга флагов командной строки
	"github.com/ilyakaznacheev/cleanenv" // Библиотека для загрузки конфигурации из файла
	"os"                                 // Модуль для работы с операционной системой
	"time"                               // Модуль для работы с временем
)

// Config содержит основные настройки приложения
type Config struct {
	Env         string        `yaml:"env" env-default:"local"`          // Среда выполнения приложения, по умолчанию "local"
	StoragePath string        `yaml:"storage_path" env-required:"true"` // Путь к файлу хранилища, обязателен для заполнения
	TokenTTL    time.Duration `yaml:"token_ttl" env-required:"true"`    // Время жизни токена, обязателен для заполнения
	GRPC        GRPCConfig    `yaml:"grpc"`                             // Настройки gRPC сервиса
}

// GRPCConfig содержит настройки для gRPC сервера
type GRPCConfig struct {
	Port    int           `yaml:"port"`    // Порт, на котором запускается gRPC сервер
	Timeout time.Duration `yaml:"timeout"` // Таймаут для gRPC соединений
}

// MustLoad загружает конфигурацию и завершает приложение при ошибке
func MustLoad() *Config {
	path := fetchConfigPath() // Получаем путь к файлу конфигурации
	if path == "" {
		panic("config path not found")
	}
	return MustLoadByPath(path)
}

// fetchConfigPath получает путь к конфигурационному файлу
func fetchConfigPath() string {
	var res string // Переменная для хранения пути

	flag.StringVar(&res, "config", "", "path to config file") // Читаем путь из аргументов командной строки
	flag.Parse()                                              // Парсим аргументы командной строки

	if res == "" { // Если путь не найден
		res = os.Getenv("CONFIG_PATH") // Пробуем получить его из переменной окружения
	}

	return res // Возвращаем путь
}

func MustLoadByPath(configPath string) *Config {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file not found: " + configPath)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}

	return &cfg
}
