package main

import (
  "io"
  "os"
  "fmt"
  "flag"
  "bufio"
  "strings"
  "io/ioutil"
  "database/sql"
  "github.com/op/go-logging"
  _ "github.com/mattn/go-sqlite3"
  "github.com/mitchellh/go-homedir"
)

var app_version = "0.1.0"
var history_database, _ = homedir.Expand("~/.shist")
var bash_history, _ = homedir.Expand("~/.bash_history")

var log = logging.MustGetLogger("shist")

var (
	printHelp bool
        version bool
        fetch bool
        load bool
)

func init_database(database *sql.DB) {
    statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS history (id INTEGER PRIMARY KEY, entry TEXT, created DATETIME DEFAULT CURRENT_TIMESTAMP)")
    statement.Exec()
}

func save_history(database *sql.DB, entry string) {
	rows, _ := database.Query("SELECT count(*) FROM history where entry = ?", entry)

	var matches int

	defer rows.Close()
	for rows.Next() {
		rows.Scan(&matches)
	}

	if (matches <= 0) {
		statement, _ := database.Prepare("INSERT INTO history (entry) VALUES (?)")
		_, err := statement.Exec(entry)
		if err != nil {
			panic(err)
		}
	}
}

func main() {
	log_backend, err := logging.NewSyslogBackend("shist")

	if err != nil { fmt.Printf("Error: %", err) }

	logging.SetBackend(log_backend)
        var format = logging.MustStringFormatter("%{level} %{message}")
        logging.SetFormatter(format)
        logging.SetLevel(logging.INFO, "shist")

	database, _ := sql.Open("sqlite3", history_database)

	init_database(database)

	flag.BoolVar(&printHelp, "help", false, "Print this help message.")
        flag.BoolVar(&fetch, "fetch", false, "Save out the bash history file")
        flag.BoolVar(&load, "load", false, "Load the current bash history file")
        flag.BoolVar(&version, "version", false, "Show Version")

        flag.Parse()

	if printHelp {
		fmt.Println("NAME:")
		fmt.Println("    shist - A smart bash history utility\n")
		fmt.Println("DESCRIPTION:")
		fmt.Println("    A smart utility for saving and fetching bash command history.\n")
		fmt.Println("VERSION:")
		fmt.Println("    " + app_version + "\n")
		fmt.Println("FLAGS:")
		flag.PrintDefaults()
		os.Exit(0)
        }

	if fetch {
		rows, _ := database.Query("SELECT * FROM history ORDER BY created asc")

		var id int
		var entry string
		var created string
		var history string

		defer rows.Close()
		for rows.Next() {
			rows.Scan(&id, &entry, &created)
			history += entry + "\n"
		}

		os.Remove(bash_history)
		err := ioutil.WriteFile(bash_history, []byte(history), 0644)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		os.Exit(0)
	}

	if load {
		var processed = 0

		fmt.Println("Importing bash history...")
		if _, err := os.Stat(bash_history); err == nil {
			f, _ := os.Open(bash_history)
			scanner := bufio.NewScanner(f)

			for scanner.Scan() {
				line := scanner.Text()
				save_history(database, line)
				processed += 1
			}
		}

		fmt.Printf("Imported %v history entries.\n", processed)
		os.Exit(0)
	}

        if version {
                fmt.Println(app_version)
                os.Exit(0)
        }

	reader := bufio.NewReader(os.Stdin)

	for {
		input, err := reader.ReadString('\n')
		if err != nil && err == io.EOF {
			break
		}

		input = strings.TrimSpace(input)

		save_history(database, input)
	}
}


