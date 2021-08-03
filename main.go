package main

import (
	"log"

	"github.com/subosito/gotenv"
	"github.com/supabase/supabase-admin-api/cmd"
)

func init() {
	gotenv.Load()
}

func main() {
	if err := cmd.RootCommand().Execute(); err != nil {
		log.Fatal(err)
	}
}
