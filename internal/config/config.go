package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Reader struct {
	viper *viper.Viper
}

func NewReader() *Reader {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	return &Reader{
		viper: v,
	}
}

func (r *Reader) Read(filename string) (conf *App, err error) {
	r.viper.SetConfigFile(filename)

	if err := r.viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var app App
	if err := r.viper.Unmarshal(&app); err != nil {
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
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	Database        string `yaml:"database"`
	User            string `yaml:"user"`
	Password        string `yaml:"password"`
	SSLMode         string `yaml:"sslmode"`
	MaxOpenConns    int    `yaml:"maxOpenConns"`
	MaxIdleConns    int    `yaml:"maxIdleConns"`
	MaxConnLifetime int    `yaml:"maxConnLifetime"`
}

type Producer struct {
	InitialBatchSize int `yaml:"initialBatchSize"`
	CreationRPS      int `yaml:"creationRPS"`
	DistinctUsers    int `yaml:"distinctUsers"`
	AmountFrom       int `yaml:"amountFrom"`
	AmountTo         int `yaml:"amountTo"`
}
