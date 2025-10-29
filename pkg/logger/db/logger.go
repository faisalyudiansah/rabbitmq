package dbLogger

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/fatih/color"
	"github.com/jackc/pgx/v5/tracelog"
)

type PgxLogger struct{}

func (l *PgxLogger) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]any) {
	var lvlColored string
	switch level {
	case tracelog.LogLevelTrace:
		lvlColored = color.New(color.FgHiBlack).Sprint("[TRACE]")
	case tracelog.LogLevelDebug:
		lvlColored = color.New(color.FgBlue).Sprint("[DEBUG]")
	case tracelog.LogLevelInfo:
		lvlColored = color.New(color.FgCyan).Sprint("[INFO]")
	case tracelog.LogLevelWarn:
		lvlColored = color.New(color.FgYellow).Sprint("[WARN]")
	case tracelog.LogLevelError:
		lvlColored = color.New(color.FgRed).Sprint("[ERROR]")
	default:
		lvlColored = color.New(color.FgWhite).Sprint("[LOG]")
	}

	sqlText := fmt.Sprintf("%v", data["sql"])
	duration := fmt.Sprintf("%v", data["time"])
	pid := fmt.Sprintf("%v", data["pid"])

	argsReadable := []string{}
	switch v := data["args"].(type) {
	case []any:
		for _, a := range v {
			strVal := fmt.Sprintf("%v", a)
			if strings.HasPrefix(strVal, "map[") && strings.Contains(strVal, ":1") {
				continue
			}
			if strings.Contains(strVal, "pgtype") {
				continue
			}
			argsReadable = append(argsReadable, fmt.Sprintf("'%v'", a))
		}
	case map[uint32]any:
		for _, a := range v {
			argsReadable = append(argsReadable, fmt.Sprintf("'%v'", a))
		}
	}

	for i, arg := range argsReadable {
		placeholder := fmt.Sprintf("$%d", i+1)
		sqlText = strings.Replace(sqlText, placeholder, arg, 1)
	}

	sqlText = strings.TrimSpace(strings.ReplaceAll(sqlText, "\n", " "))

	timeColored := color.New(color.FgHiGreen).Sprint(duration)

	log.Printf("%s %s (%s) [PID: %s] %s\n",
		lvlColored,
		strings.ToUpper(msg),
		timeColored,
		color.New(color.Bold).Sprint(pid),
		color.New(color.FgCyan).Sprint(sqlText),
	)
}
