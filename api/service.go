package api

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/go-chi/chi"
)

// RestartServices is the endpoint for fetching current goauth email config
func (a *API) RestartServices(w http.ResponseWriter, r *http.Request) error {
	sudo := "sudo"
	app := "systemctl"
	arg0 := "daemon-reload"

	cmd := exec.Command(sudo, app, arg0)
	stdout, err := cmd.Output()

	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return sendJSON(w, http.StatusInternalServerError, err.Error())
	}

	fmt.Fprintf(os.Stdout, string(stdout))

	// need to do command as goroutine because adminapi gets killed and can't respond
	go func() {
		sudo := "sudo"
		app := "systemctl"
		arg0 := "restart"
		var arg1 string

		application := chi.URLParam(r, "application")

		switch application {
		case "all":
			arg1 = "services.slice"
		case "gotrue":
			arg1 = "gotrue.service"
		case "postgrest":
			arg1 = "postgrest.service"
		case "pglisten":
			arg1 = "pglisten.service"
		case "kong":
			arg1 = "kong.service"
		case "realtime":
			arg1 = "supabase.service"
		case "adminapi":
			arg1 = "adminapi.service"
		default:
			arg1 = "services.slice"
		}

		// if admin api is getting restarted give time for http response first
		if application == "adminapi" || application == "all" {
			time.Sleep(2 * time.Second)
		}

		cmd = exec.Command(sudo, app, arg0, arg1)
		stdout, err = cmd.Output()

		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
		}

		fmt.Fprintf(os.Stdout, string(stdout))
	}()

	return sendJSON(w, http.StatusOK, 200)
}

// RebootMachine is the endpoint for fetching current goauth email config
func (a *API) RebootMachine(w http.ResponseWriter, r *http.Request) error {
	// app := "reboot"
	// exec.Command(app)

	return sendJSON(w, http.StatusInternalServerError, "endpoint not yet implemented")
}
