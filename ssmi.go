package main

import (
	"fmt"
	"os"

	"github.com/idoavrah/ssmi/internal/aws"
	"github.com/idoavrah/ssmi/internal/tui"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func main() {
	instances, err := aws.ListInstances()
	if err != nil {
		panic(err)
	}

	columns := []table.Column{
		{Title: "Name", Width: 60},
		{Title: "ID", Width: 20},
		{Title: "Type", Width: 20},
		{Title: "State", Width: 20},
	}

	rows := make([]table.Row, len(instances))
	for i, instance := range instances {
		rows[i] = table.Row{instance.Name, instance.ID, instance.Type, instance.State}

	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	m := tui.Model{t}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

}
