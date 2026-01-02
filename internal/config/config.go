package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Reader struct{}

func NewReader() *Reader {
	return &Reader{}
}

func (r *Reader) Read(filename string) (conf *App, err error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var app App
	if err := yaml.Unmarshal(data, &app); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &app, nil
}

type App struct {
	Http           *Http     `yaml:"http"`
	Kafka          *Kafka    `yaml:"kafka"`
	PostgresMaster *Postgres `yaml:"postgresMaster"`
	PostgresSlave  *Postgres `yaml:"postgresSlave"`
	Producer       *Producer `yaml:"producer"`
}

type Http struct {
	Port int `yaml:"port"`
}

type Kafka struct {
	ConnectionString string `yaml:"connectionString"`
	User             string `yaml:"user"`
	Password         string `yaml:"password"`
	Topic            string `yaml:"topic"`
	GroupID          string `yaml:"groupId"`
	RequiredAcks     int    `yaml:"requiredAcks"`
	MaxAttempts      int    `yaml:"maxAttempts"`
}

type Postgres struct {
	ConnectionString string `yaml:"connectionString"`
	User             string `yaml:"user"`
	Password         string `yaml:"password"`
	MaxOpenConns     int    `yaml:"maxOpenConns"`
	MaxIdleConns     int    `yaml:"maxIdleConns"`
	MaxConnLifetime  int    `yaml:"maxConnLifetime"`
}

type Producer struct {
	InitialBatchSize int `yaml:"initialBatchSize"`
	CreationRPS      int `yaml:"creationRPS"`
	DistinctUsers    int `yaml:"distinctUsers"`
	AmountFrom       int `yaml:"amountFrom"`
	AmountTo         int `yaml:"amountTo"`
}
