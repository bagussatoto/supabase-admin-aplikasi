package cmd

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/supabase/supabase-admin-api/api"
	"os"
)

var serveCmd = cobra.Command{
	Use:  "serve",
	Long: "Start API server",
	Run: func(cmd *cobra.Command, args []string) {
		serve()
	},
}

func serve() {
	config := api.Config{}
	err := envconfig.Process("", &config); if err != nil {
		logrus.Errorf("Could not read in config. %+v", err)
		os.Exit(1)
	}

	createdApiInstance := api.NewAPIWithVersion(&config, Version)
	l := fmt.Sprintf("%v:%v", config.Host, config.Port)
	logrus.Infof("Supabase Admin API started on: %s", l)
	createdApiInstance.ListenAndServe(l)
}
