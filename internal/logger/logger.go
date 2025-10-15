package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

func NewLogger(cfgPath string) (*zap.Logger, *zap.SugaredLogger, error){
	fmt.Println("Loading logger config from:", cfgPath)

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		fmt.Println("logger readfile error")
		return nil, nil, err
	}
		
	fmt.Println("config file read ok, size:", len(data))

	var cfg zap.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		fmt.Println("logger json unmarshal error:", err)
		return nil, nil, err
	}

	for _, path := range cfg.OutputPaths {
		if path == "stdout" || path == "stderr" {
			continue
		}

		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Println("logger mkdir error:", err)
			return nil, nil, err
		}
	}

	logger, err := cfg.Build()
	if err != nil {
		fmt.Println("logger build error:", err)
		return nil, nil, err
	}

	sugar := logger.Sugar()

	sugar.Info("logger initialization succeeded")
	return logger, sugar, nil
}