package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/Lukas-Les/gator/internal/config"
	"github.com/Lukas-Les/gator/internal/database"
	_ "github.com/lib/pq"
)

const dbURL = "postgres://postgres:postgres@localhost:5432/gator"

type state struct {
	db     *database.Queries
	config *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	cmds map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	fmt.Printf("running: %v\n with params %v\n", cmd.name, cmd.args)
	if _, ok := c.cmds[cmd.name]; !ok {
		return fmt.Errorf("command '%s' is not registered", cmd.name)
	}
	err := c.cmds[cmd.name](s, cmd)
	if err != nil {
		fmt.Printf("command returned an error: %v\n", err)
	}
	return nil
}

func (c *commands) register(name string, f func(*state, command) error) error {
	_, exists := c.cmds[name]
	if exists {
		return fmt.Errorf("command '%s' is being registered two times!", name)
	}
	c.cmds[name] = f
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("not enough parameters!")
		fmt.Println("gator <command> [arguments]")
		os.Exit(1)
	}

	cfgFilePath, err := config.GetConfigFilePath()
	fmt.Printf("looking for cfg at: %v", cfgFilePath)
	if err != nil {
		panic(err)
	}
	cfg, _ := config.Read(cfgFilePath)

	// initializing db
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to connect to db")
	}
	dbQueries := database.New(db)

	s := state{config: &cfg, db: dbQueries}

	cmd := command{name: os.Args[1], args: os.Args[2:]}

	cmds := commands{cmds: map[string]func(*state, command) error{}}
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerUsers)
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", handlerAddFeed)

	err = cmds.run(&s, cmd)
	if err != nil {
		fmt.Printf("%v", err)
	}

	fmt.Printf("[connection string is: %v]\n", cfg.DbUrl)
	fmt.Printf("[username is: %v]\n", cfg.CurrentUserName)

}
