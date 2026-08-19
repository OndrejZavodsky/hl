package hl

import (
	"fmt"
	"hl/ui"
	"os"
)

func main() {
	cfg := ui.ParseFlags()
	//add the logic for just validation and other cfg options
	pipeline, err := ui.LoadSteps(cfg.ConfigFile)
	if err != nil {
		os.Exit(2)
	}
	missing := ui.CheckTools(pipeline)
	if missing != nil {
		fmt.Errorf("could not find thees tools %w", missing)
		os.Exit(2)
	}
}
