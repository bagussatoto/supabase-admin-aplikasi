package api

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"

	"github.com/go-chi/chi"
)

const postgrestLogPath string = "/var/log/postgrest.stdout"

const pgListenLogPath string = "/var/log/pg_listen.stdout"

const gotrueLogPath string = "/var/log/gotrue.stdout"

const adminAPILogPath string = "/var/log/admin-api.stdout"

const realtimeLogPath string = "/var/log/supabase.stdout"

const kongLogPath string = "/usr/local/kong/logs/access.log"
const kongErrorLogPath string = "/usr/local/kong/logs/error.log"

const syslogPath string = "/var/log/syslog"

// GetLogContents is the method for returning the contents of a given log file
func (a *API) GetLogContents(w http.ResponseWriter, r *http.Request) error {
	application := chi.URLParam(r, "application")

	// fetchType is head, tail
	fetchType := chi.URLParam(r, "type")

	// number of lines if head or tail
	n := chi.URLParam(r, "n")

	// default is concatenate entire file
	app := "tail"
	arg0 := "-n"
	arg1 := "100"
	logFilePath := syslogPath

	switch application {
	case "test":
		logFilePath = "./README.md"
	case "gotrue":
		logFilePath = gotrueLogPath
	case "postgrest":
		logFilePath = postgrestLogPath
	case "pglisten":
		logFilePath = pgListenLogPath
	case "kong":
		logFilePath = kongLogPath
	case "kong-error":
		logFilePath = kongErrorLogPath
	case "realtime":
		logFilePath = realtimeLogPath
	case "admin":
		logFilePath = adminAPILogPath
	case "syslog":
		logFilePath = syslogPath
	}

	switch fetchType {
	case "head":
		app = "head"
		arg0 = "-n"
		arg1 = n
	case "tail":
		app = "tail"
		arg0 = "-n"
		arg1 = n
	}

	cmd := exec.Command(app, arg0, arg1, logFilePath)
	stdout, err := cmd.Output()

	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return sendJSON(w, http.StatusInternalServerError, err.Error())
	}

	return sendJSON(w, http.StatusOK, string(stdout))
}
