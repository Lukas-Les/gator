package main

import (
	"fmt"
	"os"

	"github.com/Lukas-Les/gator/internal/config"
)

type state struct {
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
		return fmt.Errorf("Command '%s' is not registered", cmd.name)
	}
	c.cmds[cmd.name](s, cmd)
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
	if len(os.Args) < 3 {
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
	s := state{config: &cfg}

	cmd := command{name: os.Args[1], args: os.Args[2:]}

	cmds := commands{cmds: map[string]func(*state, command) error{}}
	cmds.register("login", handlerLogin)

	err = cmds.run(&s, cmd)
	if err != nil {
		fmt.Printf("%v", err)
	}

	fmt.Printf("connection string is: %v\n", cfg.DbUrl)
	fmt.Printf("username is: %v\n", cfg.CurrentUserName)
}
