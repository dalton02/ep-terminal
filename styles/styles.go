package styles

import "charm.land/lipgloss/v2"

var (
	titleC  = lipgloss.Color("#7fd1ae")
	optionC = lipgloss.Color("#d7d7d7")
)

var BigRetroStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(titleC).MarginLeft(2).MarginTop(2).MarginBottom(0)

var MediumRetroStyle = lipgloss.NewStyle().
	Bold(true).
	Background(lipgloss.Color("#676767")).
	Foreground(optionC).Width(50)

var OutputCMDStyle = lipgloss.NewStyle().
	Bold(true).
	MarginLeft(2).
	Foreground(optionC).Width(50)
