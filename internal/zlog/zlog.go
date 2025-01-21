package zlog

import (
	"fmt"
	"sync"

	"go.uber.org/zap"
)

var logger *zap.SugaredLogger
var initOnce sync.Once

func Logger() *zap.SugaredLogger {
	initOnce.Do(func() {
		base, err := zap.NewDevelopment()
		if err != nil {
			panic(fmt.Errorf("can't initialize zap logger: %s", err))
		}

		logger = base.Sugar()
	})

	return logger
}
