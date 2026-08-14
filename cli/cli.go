package cli

import "flag"

type Config struct {
	config_file string
	backup_file string
	do_backup   bool
	validate    bool
}

func Parse_flags() Config {
	cf := flag.String("config_file", "", "determines the path to a file that holds config")
	bf := flag.String("backup_file", "", "determines the path at which backup_file should be created")
	backup := flag.Bool("backup", false, "switches functionality to only create a backup")
	validate := flag.Bool("validate", false, "switches functionality to only validate the config")
	flag.Parse()
	return Config{config_file: *cf, backup_file: *bf, do_backup: *backup, validate: *validate}
}
