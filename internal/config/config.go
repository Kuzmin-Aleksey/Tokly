package config

import (
	"errors"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	Debug            bool             `json:"debug" yaml:"debug" env:"DEBUG" envDefault:"false"`
	Http             *HttpConfig      `json:"http" yaml:"http"`
	MySQL            *MySQLConfig     `json:"mysql" yaml:"mysql"`
	YoloModel        *YoloModelConfig `json:"yolo_model" yaml:"yolo_model"`
	ImagesPath       string           `json:"images_path" yaml:"images_path"`
	DefaultLapConfig map[string]int   `json:"default_lap_config" yaml:"default_lap_config"`
}

type HttpConfig struct {
	Host             string `json:"host" yaml:"host" env:"HTTP_HOST" envDefault:"localhost"`
	ReadTimeoutSec   int    `json:"read_timeout_sec" yaml:"read_timeout_sec" env:"HTTP_READ_TIMEOUT_SEC" envDefault:"10"`
	HandleTimeoutSec int    `json:"handle_timeout_sec" yaml:"handle_timeout_sec" env:"HTTP_HANDE_TIMEOUT_SEC" envDefault:"20"`
	WriteTimeoutSec  int    `json:"write_timeout_sec" yaml:"write_timeout_sec" env:"HTTP_WRITE_TIMEOUT_SEC" envDefault:"10"`
	SSLKeyPath       string `json:"ssl_key_path" yaml:"ssl_key_path" env:"HTTP_SSL_KEY_PATH"`
	SSLCertPath      string `json:"ssl_cert_path" yaml:"ssl_cert_path" env:"HTTP_SSL_CERT_PATH"`
}

type MySQLConfig struct {
	Host              string `json:"host" yaml:"host" env:"MYSQL_HOST" envDefault:"localhost"`
	Port              int    `json:"port" yaml:"port" env:"MYSQL_PORT" envDefault:"3306"`
	User              string `json:"user" yaml:"user" env:"MYSQL_USER" envDefault:"root"`
	Password          string `json:"password" yaml:"password" env:"MYSQL_PASSWORD" envDefault:"pass"`
	Schema            string `json:"schema" yaml:"schema" env:"MYSQL_SCHEMA" envDefault:"app"`
	ConnectTimeoutSec int    `json:"connect_timeout_sec" yaml:"connect_timeout_sec" env:"MYSQL_CONNECT_TIMEOUT_SEC" envDefault:"10"`
}

type YoloModelConfig struct {
	Model          string `json:"model" yaml:"model" env:"YOLO_MODEL"`
	ModelConfig    string `json:"model_config" yaml:"model_config" env:"YOLO_MODEL_CONFIG"`
	ModelSeg       string `json:"model_seg" yaml:"model_seg" env:"YOLO_MODEL_SEG"`
	ModelSegConfig string `json:"model_seg_config" yaml:"model_seg_config" env:"YOLO_MODEL_SEG_CONFIG"`
}

func ReadConfig(path string, dotenv ...string) (*Config, error) {
	if err := godotenv.Load(dotenv...); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}

	cfg := new(Config)
	cfg.DefaultLapConfig = make(map[string]int)

	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
