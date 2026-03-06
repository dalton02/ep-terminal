package main

import (
	"ep-terminal/global"
	"ep-terminal/styles"
	"ep-terminal/task"
	"fmt"
	"os"
	"strconv"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/common-nighthawk/go-figure"
)

type model struct {
	y           int
	x           int
	blink       bool
	currentTask *task.Task
	spinner     spinner.Model
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
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

			if m.currentTask.Processing {
				m.currentTask.Messagers.Global <- "kill-process"
			}
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
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	}
	return m, nil
}

func (m model) View() tea.View {
	v := tea.NewView("")

	v.Content += styles.BigRetroStyle.Render(figure.NewFigure("EP TERMINAL", "roman", true).String())
	v.Content += "\n"

	if m.currentTask.Prev != nil {
		text := "Context: " + m.currentTask.Title
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

		v.Content += styles.OutputCMDStyle.Foreground(lipgloss.Color("#ececec")).Render("\n --- Command execution below ---")
		v.Content += "\n\n"
		v.Content += styles.OutputCMDStyle.Foreground(lipgloss.Color("#e2e2e2")).Render(m.currentTask.CurrentContentCMD)

		if m.currentTask.Processing {
			v.Content += styles.OutputCMDStyle.Foreground(lipgloss.Color("#e2e2e2")).Render("\nPress ctrl+c or q to cancel operation in execution", m.spinner.View())
		} else {
			v.Content += styles.OutputCMDStyle.Foreground(lipgloss.Color("#ececec")).Render("\n ------ Command finished ------\n")

		}

	}
	v.Content += "\n"
	v.Content += styles.MediumRetroStyle.Background(lipgloss.Color("")).MarginLeft(2).Render("Help: \nUse ↑ ↓ to move between options\nSpace/Enter to select\nBackSpace to go back")

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
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#ececec"))
	m.spinner = s
	global.Program = tea.NewProgram(m)
	if _, err := global.Program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
		os.Exit(1)
	}
}
