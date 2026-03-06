package task

import (
	"bufio"
	"context"
	"encoding/json"
	"ep-terminal/global"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type Task struct {
	Title             string
	Options           []*Task
	CMD               string
	Prev              *Task
	CurrentContentCMD string
	Processing        bool
	Messagers         Messagers
	Env               string
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

	ctx := context.WithValue(context.Background(), "env", ".env")

	processParent(&tasks, ctx)
	return &tasks, nil
}

func processParent(t *Task, ctx context.Context) {

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

	if t.Env == "" {
		t.Env = ctx.Value("env").(string)
	} else {
		ctx = context.WithValue(ctx, "env", t.Env)
	}

	for _, op := range t.Options {
		op.Prev = t
		processParent(op, ctx)
	}
}

func (t *Task) formatCommand() []string {

	parts := strings.Split(t.CMD, " ")
	envData, err := LoadEnv(t.Env)
	if err != nil {
		return parts
	}

	for i, part := range parts {

		envDetect := regexp.MustCompile(`\{env\.([^}]+)\}`)
		converted := envDetect.ReplaceAllFunc([]byte(part), func(b []byte) []byte {

			nameEnv := string(b)
			nameEnv = strings.ReplaceAll(nameEnv, "{", "")
			nameEnv = strings.ReplaceAll(nameEnv, "}", "")
			nameEnv = strings.ReplaceAll(nameEnv, "env.", "")
			valueEnv := envData[nameEnv]

			return []byte(valueEnv)
		})
		parts[i] = string(converted)

	}

	return parts

}

func LoadEnv(filename string) (map[string]string, error) {
	env := make(map[string]string)

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("não foi possível abrir o arquivo %s: %w", filename, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("formato inválido na linha %d: %s", lineNum, line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		value = strings.Trim(value, `"'`)

		for strings.HasSuffix(value, "\\") {
			value = strings.TrimSuffix(value, "\\")
			if !scanner.Scan() {
				break
			}
			lineNum++
			nextLine := strings.TrimSpace(scanner.Text())
			value += nextLine
			value = strings.Trim(value, `"'`)
		}

		env[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("erro ao ler o arquivo: %w", err)
	}

	return env, nil
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
	stderr, err := command.StderrPipe()
	if err != nil {
		cancel()
		return err
	}

	if err = command.Start(); err != nil {
		cancel()
		return err
	}

	t.Processing = true
	go func(stdout io.ReadCloser, stderr io.ReadCloser, command *exec.Cmd) {

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

		scanner = bufio.NewScanner(stderr)

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

	}(stdout, stderr, command)

	return nil
}
