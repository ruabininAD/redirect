package config

import (
	"github.com/joho/godotenv"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"strconv"
	"strings"
)

var Log *zap.Logger
var Cfg Config

type Config struct {
	Service       string `yaml:"service"`
	Port          string `yaml:"port"`
	DspToDatabase string `yaml:"dsp_to_database"`
	MetricPort    string `yaml:"metricPort"`
	Logger        struct {
		Pretty     bool   `yaml:"prettyLog"`
		Format     string `yaml:"format"`
		Level      string `yaml:"level"`
		Output     string `yaml:"output"`
		StackTrace bool   `yaml:"stacktrace"`

		OutputPaths []string `yaml:"outputPaths"`
	} `yaml:"logger"`
}

func Init() {
	//mustInitConfigFile()
	mustInitEnvFile()
	initLog()
	tracerInit()
}

func mustInitConfigFile() {
	// Чтение файла конфигурации
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		Log.Panic("Ошибка чтения файла конфигурации:" + err.Error())
	}

	// Разбор YAML и запись в структуру
	err = yaml.Unmarshal(data, &Cfg)
	if err != nil {
		Log.Panic("Ошибка разбора YAML: " + err.Error())
	}
}

func mustInitEnvFile() {
	// Загрузка переменных окружения из файла .env
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Ошибка при загрузке файла .env: %s", err)
	}

	Cfg.Port = os.Getenv("APP_LISTEN_PORT")
	Cfg.DspToDatabase = os.Getenv("DSP_TO_DATABASE")
	Cfg.Logger.Level = os.Getenv("LOGGER_LEVEL")
	Cfg.Logger.OutputPaths = strings.Split(os.Getenv("LOGGER_OUTPUTPATHS"), ",")
	Cfg.Logger.Output = os.Getenv("LOGGER_OUTPUT")
	Cfg.Logger.Format = os.Getenv("LOGGER_FORMAT")
	Cfg.Service = os.Getenv("SERVICE")
	Cfg.MetricPort = os.Getenv("METRICS_PORT")
	Cfg.Logger.Pretty, err = strconv.ParseBool(os.Getenv("LOGGER_PRETTYLOG"))
	if err != nil {
		log.Fatalf("error parse LOGGER_PRETTYLOG .env: %s", err)
	}
	Cfg.Logger.StackTrace, err = strconv.ParseBool(os.Getenv("LOGGER_STACKTRACE"))
	if err != nil {
		log.Fatalf("error parse LOGGER_STACKTRACE .env: %s", err)
	}

}

func initLog() {

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.EpochMillisTimeEncoder //ISO8601TimeEncoder
	encoderCfg.EncodeLevel = zapcore.LowercaseLevelEncoder //zapcore.CapitalLevelEncoder()
	encoderCfg.MessageKey = "msg"
	encoderCfg.TimeKey = "ts"

	var logLevel zapcore.Level
	switch Cfg.Logger.Level {
	case "info":
		logLevel = zap.InfoLevel
	case "debug":
		logLevel = zap.DebugLevel
	case "error":
		logLevel = zap.ErrorLevel
	case "warn":
		logLevel = zap.WarnLevel
	case "panic":
		logLevel = zap.PanicLevel
	case "fatal":
		logLevel = zap.FatalLevel
	default:
		logLevel = zap.InfoLevel

	}

	encoding := Cfg.Logger.Format
	if Cfg.Logger.Pretty {
		encoding = "console"
		encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	logConfig := zap.Config{
		Level:             zap.NewAtomicLevelAt(logLevel),
		Development:       true,
		DisableCaller:     true,
		DisableStacktrace: Cfg.Logger.StackTrace,
		Sampling:          nil,
		Encoding:          encoding,
		EncoderConfig:     encoderCfg,
		OutputPaths:       Cfg.Logger.OutputPaths,
		ErrorOutputPaths:  Cfg.Logger.OutputPaths,
		InitialFields:     map[string]interface{}{},
	}
	Log = zap.Must(logConfig.Build())
	Log = Log.With(zap.Field{
		Key:    "service",
		Type:   zapcore.StringType,
		String: Cfg.Service,
	})
	Log.Debug("log init")
}

func tracerInit() {

	// Настройка провайдера трассировки с генератором идентификаторов
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)

}
