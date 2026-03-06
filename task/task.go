package task

import (
	"bufio"
	"context"
	"encoding/json"
	"ep-terminal/global"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
)

type Task struct {
	Title             string
	Options           []*Task
	CMD               string
	Prev              *Task
	CurrentContentCMD string
	Processing        bool
	Messagers         Messagers
}

type Messagers struct {
	Global chan string
	CMD    chan string
}

func ProcessEpFile() (*Task, error) {
	r, err := os.ReadFile("epActions.json")
	if err != nil {
		return nil, err
	}
	tasks := Task{}
	err = json.Unmarshal(r, &tasks)
	if err != nil {
		return nil, err
	}

	processParent(&tasks)

	return &tasks, nil
}

func processParent(t *Task) {

	t.Messagers.Global = make(chan string)
	t.Messagers.CMD = make(chan string)

	go func() {
		for {
			select {
			case rcv := <-t.Messagers.Global:
				t.Messagers.CMD <- rcv
			}
		}
	}()

	for _, op := range t.Options {
		op.Prev = t
		processParent(op)
	}
}

func (t *Task) formatCommand() []string {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	parts := strings.Split(t.CMD, " ")

	for i, part := range parts {

		envDetect := regexp.MustCompile(`\{env\.([^}]+)\}`)
		converted := envDetect.ReplaceAllFunc([]byte(part), func(b []byte) []byte {
			nameEnv := string(b)
			nameEnv = strings.ReplaceAll(nameEnv, "{", "")
			nameEnv = strings.ReplaceAll(nameEnv, "}", "")
			nameEnv = strings.ReplaceAll(nameEnv, "env.", "")
			valueEnv := os.Getenv(nameEnv)

			return []byte(valueEnv)
		})
		parts[i] = string(converted)

	}

	return parts

}

func (t *Task) ExecuteCommand() error {

	ctx, cancel := context.WithCancel(context.Background())

	parts := t.formatCommand()

	command := exec.CommandContext(ctx, parts[0], parts[1:]...)

	stdout, err := command.StdoutPipe()
	if err != nil {
		cancel()
		return err
	}

	if err = command.Start(); err != nil {
		cancel()
		return err
	}

	t.Processing = true
	go func(stdout io.ReadCloser, command *exec.Cmd) {

		go func(command *exec.Cmd) {
			select {
			case rcv := <-t.Messagers.CMD:
				switch rcv {
				case "kill-process":
					if command != nil {
						command.Cancel()
						cancel()
						t.Processing = false
						t.CurrentContentCMD = ""
					}
				}
			}
		}(command)

		scanner := bufio.NewScanner(stdout)

		for scanner.Scan() {
			line := scanner.Text()
			t.CurrentContentCMD += line + "\n"
			type force struct{}
			global.Program.Send(force{})
		}
		if err := scanner.Err(); err != nil {

			t.CurrentContentCMD += err.Error() + "\n"
			type force struct{}
			global.Program.Send(force{})
		}
		t.Processing = false

	}(stdout, command)

	return nil
}
