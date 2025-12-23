Gator

This is a blog aggregator is written in Go and uses Postgres for storage.

Prerequisites:
- Go installed in your machine for build
- Go added to your $PATH

Installation:
1. Clone the repo to your machine
2. Run Go build to build
3. Run Go install to install this on your machine. Now you can run gator command from anywhere on your machine.

Usage:
Run gator <command> [arguments]

First, you'll need to register. To do it, run gator with a register command.
```bash
gator register username
```

To log in, run login command. 

To add feed for your current user, use the addfeed command.
For browsing, use browse.
