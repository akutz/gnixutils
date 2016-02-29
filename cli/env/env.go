package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
)

var (
	envVarRx      *regexp.Regexp
	specifiedOnly bool
)

func init() {
	envVarRx = regexp.MustCompile(`(.+?)=(.+)`)

	flag.BoolVar(&specifiedOnly, "i", false,
		"Execute the utility with only those environment variables specified.")
}

func main() {
	flag.Parse()

	envMap := map[string]string{}

	if !specifiedOnly {
		for _, v := range os.Environ() {
			m := envVarRx.FindStringSubmatch(v)
			if len(m) == 3 {
				envMap[m[1]] = m[2]
			}
		}
	}

	var cmd string
	var args []string
	flagArgs := flag.Args()

	for x, p := range flagArgs {
		m := envVarRx.FindStringSubmatch(p)
		if len(m) == 3 {
			envMap[m[1]] = m[2]
		} else {
			cmd = p
			args = flagArgs[x+1:]
			break
		}
	}

	if cmd == "" {
		os.Exit(1)
	}

	env := []string{}
	for k, v := range envMap {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	c := exec.Command(cmd, args...)
	c.Env = env
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
