package main

import (
	"ep-terminal/global"
	"ep-terminal/styles"
	"ep-terminal/task"
	"fmt"
	"os"
	"strconv"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/common-nighthawk/go-figure"
)

type model struct {
	y           int
	x           int
	blink       bool
	currentTask *task.Task
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {

		case "down":
			if m.y >= len(m.currentTask.Options)-1 {
				m.y = 0
				break
			}
			m.y += 1
		case "up":
			if m.y <= 0 {
				m.y = len(m.currentTask.Options) - 1
				break
			}
			m.y -= 1
		case "space", "enter":
			if len(m.currentTask.Options) > 0 {
				m.currentTask = m.currentTask.Options[m.y]
				if len(m.currentTask.Options) == 0 {
					m.currentTask.ExecuteCommand()
				}
				m.y = 0
			}
		case "backspace":
			if m.currentTask.Prev != nil {
				m.currentTask.CurrentContentCMD = ""
				m.currentTask = m.currentTask.Prev
			}
		case "ctrl+c", "q":
			if m.currentTask.Processing {
				m.currentTask.Messagers.Global <- "kill-process"
				m.currentTask = m.currentTask.Prev

				return m, nil
			}
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	v := tea.NewView("")

	v.Content += styles.BigRetroStyle.Render(figure.NewFigure("EP TERMINAL", "roman", true).String())
	v.Content += "\n"

	if m.currentTask.Prev != nil {
		text := "Inside: (" + m.currentTask.Title + ")"
		v.Content += styles.MediumRetroStyle.Background(lipgloss.Color("")).Render(text)
		v.Content += "\n\n"

	}

	count := 1
	for i, op := range m.currentTask.Options {
		text := "  " + strconv.Itoa(count) + " " + op.Title
		if i != m.y {
			v.Content += styles.MediumRetroStyle.Background(lipgloss.Color("")).Render(text)
		} else {
			v.Content += styles.MediumRetroStyle.Render(text)
		}
		v.Content += "\n"
		count++
	}

	if m.currentTask.CurrentContentCMD != "" {
		v.Content += styles.OutputCMDStyle.Render("Processing command: \n")
		v.Content += styles.OutputCMDStyle.Render("\n" + m.currentTask.CurrentContentCMD)
		v.Content += styles.OutputCMDStyle.Render("\nPress ctrl+c or q to cancel operation\n")
	}
	v.Content += "\n"
	return v
}

func main() {
	tasks, err := task.ProcessEpFile()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	m := model{
		currentTask: tasks,
		y:           0,
	}

	global.Program = tea.NewProgram(m)
	if _, err := global.Program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
		os.Exit(1)
	}
}
