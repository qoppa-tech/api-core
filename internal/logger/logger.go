package logger

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type JSONLogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	Status    int    `json:"status"`
	Latency   string `json:"latency"`
	ClientIP  string `json:"client_ip"`
	UserAgent string `json:"user_agent,omitempty"`
	Error     string `json:"error,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

func SetupLogger() gin.HandlerFunc {
	env := os.Getenv("APP_ENV")

	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
		return JSONLogger()
	}

	gin.SetMode(gin.DebugMode)
	return DebugLogger()
}

func JSONLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		entry := JSONLogEntry{
			Timestamp: param.TimeStamp.Format(time.RFC3339),
			Level:     getLogLevel(param.StatusCode),
			Method:    param.Method,
			Path:      param.Path,
			Status:    param.StatusCode,
			Latency:   param.Latency.String(),
			ClientIP:  param.ClientIP,
			UserAgent: param.Request.UserAgent(),
			Error:     param.ErrorMessage,
		}

		if reqID := param.Request.Header.Get("X-Request-ID"); reqID != "" {
			entry.RequestID = reqID
		}

		return formatJSON(entry) + "\n"
	})
}

func DebugLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		statusColor := getStatusColor(param.StatusCode)
		methodColor := getMethodColor(param.Method)
		resetColor := "\033[0m"

		return formatDebug(param, statusColor, methodColor, resetColor)
	})
}

func getLogLevel(status int) string {
	switch {
	case status >= 500:
		return "error"
	case status >= 400:
		return "warn"
	default:
		return "info"
	}
}

func getStatusColor(code int) string {
	switch {
	case code >= 500:
		return "\033[91m" // Red
	case code >= 400:
		return "\033[93m" // Yellow
	case code >= 300:
		return "\033[96m" // Cyan
	case code >= 200:
		return "\033[92m" // Green
	default:
		return "\033[97m" // White
	}
}

func getMethodColor(method string) string {
	switch method {
	case "GET":
		return "\033[94m" // Blue
	case "POST":
		return "\033[96m" // Cyan
	case "PUT":
		return "\033[93m" // Yellow
	case "DELETE":
		return "\033[91m" // Red
	case "PATCH":
		return "\033[95m" // Magenta
	default:
		return "\033[97m" // White
	}
}

func formatJSON(entry JSONLogEntry) string {
	json := `{"timestamp":"` + entry.Timestamp + `"`
	json += `,"level":"` + entry.Level + `"`
	json += `,"method":"` + entry.Method + `"`
	json += `,"path":"` + entry.Path + `"`
	json += `,"status":` + intToString(entry.Status)
	json += `,"latency":"` + entry.Latency + `"`
	json += `,"client_ip":"` + entry.ClientIP + `"`

	if entry.UserAgent != "" {
		json += `,"user_agent":"` + escapeJSON(entry.UserAgent) + `"`
	}
	if entry.Error != "" {
		json += `,"error":"` + escapeJSON(entry.Error) + `"`
	}
	if entry.RequestID != "" {
		json += `,"request_id":"` + entry.RequestID + `"`
	}

	json += "}"
	return json
}

func formatDebug(param gin.LogFormatterParams, statusColor, methodColor, reset string) string {
	latency := param.Latency
	if latency > time.Minute {
		latency = latency.Truncate(time.Second)
	}

	errorMsg := ""
	if param.ErrorMessage != "" {
		errorMsg = " | \033[91mERROR: " + param.ErrorMessage + reset
	}

	return param.TimeStamp.Format("2006/01/02 15:04:05") +
		" | " + statusColor + intToString(param.StatusCode) + reset +
		" | " + latency.String() +
		" | " + param.ClientIP +
		" | " + methodColor + param.Method + reset +
		" " + param.Path +
		errorMsg + "\n"
}

func intToString(n int) string {
	if n == 0 {
		return "0"
	}

	var result []byte
	negative := n < 0
	if negative {
		n = -n
	}

	for n > 0 {
		result = append([]byte{byte('0' + n%10)}, result...)
		n /= 10
	}

	if negative {
		result = append([]byte{'-'}, result...)
	}

	return string(result)
}

func escapeJSON(s string) string {
	result := ""
	for _, c := range s {
		switch c {
		case '"':
			result += `\"`
		case '\\':
			result += `\\`
		case '\n':
			result += `\n`
		case '\r':
			result += `\r`
		case '\t':
			result += `\t`
		default:
			result += string(c)
		}
	}
	return result
}
