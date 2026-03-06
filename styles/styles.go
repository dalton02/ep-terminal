package styles

import "charm.land/lipgloss/v2"

var (
	titleC  = lipgloss.Color("#FF3E9B")
	optionC = lipgloss.Color("#66D0BC")
)

var BigRetroStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(titleC).MarginLeft(2).MarginTop(2).MarginBottom(0)

var MediumRetroStyle = lipgloss.NewStyle().
	Bold(true).
	Background(lipgloss.Color("#494949")).
	Foreground(optionC).Width(50)

var OutputCMDStyle = lipgloss.NewStyle().
	Bold(true).
	MarginLeft(2)
