package gologger

import (
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

type ctxKey string

const ReqIDKey ctxKey = "reqID"

func init() {
	l := NewLogger()
	zerolog.DefaultContextLogger = &l
	callerCache := map[uintptr]string{}
	callerCacheMutex := &sync.RWMutex{}
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		callerCacheMutex.RLock()
		c, ok := callerCache[pc]
		callerCacheMutex.RUnlock()
		if ok {
			return c
		}
		callerCacheMutex.Lock()
		defer callerCacheMutex.Unlock()
		function := ""
		fun := runtime.FuncForPC(pc)
		if fun != nil {
			funName := fun.Name()
			slash := strings.LastIndex(funName, "/")
			if slash > 0 {
				funName = funName[slash+1:]
			}
			function = " " + funName + "()"
		}
		c = file + ":" + strconv.Itoa(line) + function
		callerCache[pc] = c
		return c
	}
}

func NewLogger() zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	// zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	zerolog.TimestampFieldName = "time"

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	logger = logger.Hook(CallerHook{})

	if os.Getenv("PRETTY") == "1" {
		logger = logger.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
	if os.Getenv("DEBUG") == "1" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	return logger
}

type CallerHook struct{}

func (h CallerHook) Run(e *zerolog.Event, _ zerolog.Level, _ string) {
	e.Caller(3)
}
