package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gouline/blaster/internal/pkg/server"
	zaplogfmt "github.com/sykesm/zap-logfmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	debug := os.Getenv("DEBUG") == "1"

	var logger *zap.Logger
	if debug {
		logger = zap.Must(zap.NewDevelopment())
	} else {
		loggerConfig := zap.NewProductionEncoderConfig()
		loggerConfig.EncodeTime = func(ts time.Time, encoder zapcore.PrimitiveArrayEncoder) {
			encoder.AppendString(ts.UTC().Format(time.RFC3339))
		}
		logger = zap.New(zapcore.NewCore(
			zaplogfmt.NewEncoder(loggerConfig),
			os.Stdout,
			zapcore.DebugLevel,
		))
	}
	defer logger.Sync()

	s, err := server.New(server.Config{
		Logger:            logger,
		Debug:             debug,
		Host:              os.Getenv("HOST"),
		Port:              os.Getenv("PORT"),
		CertFile:          os.Getenv("CERT_FILE"),
		KeyFile:           os.Getenv("KEY_FILE"),
		StaticRoot:        "static",
		TemplatesRoot:     "templates",
		SlackClientID:     os.Getenv("SLACK_CLIENT_ID"),
		SlackClientSecret: os.Getenv("SLACK_CLIENT_SECRET"),
	})
	if err != nil {
		panic(fmt.Sprintf("failed to create server: %s", err))
	}
	err = s.Start()
	logger.Fatal("server stopped", zap.Error(err))
}
