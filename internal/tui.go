package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	app                *tview.Application
	currentEc2         []Instance
	historyEc2         *HistoryList
	historyListPanel   *tview.List
	currentTablePanel  *tview.Table
	profile            string
	historyFilename    string
	focusPanel         any
	selectedEC2Name    string
	selectEC2ID        string
	selectedEC2Profile string
)

func init() {
	homedir, _ := os.UserHomeDir()
	historyFilename = filepath.Join(homedir, ".ssmi_history")
	app = tview.NewApplication()
	currentEc2 = []Instance{}
	historyEc2 = LoadHistoryList(historyFilename)
	historyListPanel = tview.NewList()
	currentTablePanel = tview.NewTable()
}

func exitApp() {
	app.Stop()
	historyEc2.Save(historyFilename)
	os.Exit(0)
}

func runSSM() {

	app.Stop()
	addToHistory()

	fmt.Printf("Starting session for instance %s (%s) using profile %s", selectedEC2Name, selectEC2ID, selectedEC2Profile)
	cmd := exec.Command("aws", "ssm", "start-session", "--target", selectEC2ID, "--profile", selectedEC2Profile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	cmd.Run()
	exitApp()
}

func refreshHistory() {
	historyListPanel.Clear()
	for idx, item := range historyEc2.Items {
		historyListPanel.AddItem(fmt.Sprintf("%s (%s)", item.Name, item.ID), fmt.Sprintf("Profile: %s", item.Profile), '0'+rune(idx), nil)
	}
}

func addToHistory() {
	historyEc2.Add(HistoryItem{
		ID:      selectEC2ID,
		Name:    selectedEC2Name,
		Profile: selectedEC2Profile,
	})
	refreshHistory()
}

func SwitchFocus() {
	if focusPanel == historyListPanel {
		app.SetFocus(currentTablePanel)
		focusPanel = currentTablePanel
	} else if focusPanel == currentTablePanel {
		historyListPanel.SetCurrentItem(-1)
		app.SetFocus(historyListPanel)
		focusPanel = historyListPanel
	}
}

func StartApplication() {

	profile = os.Getenv("AWS_PROFILE")
	if profile == "" {
		profile = "No Profile"
	}

	historyListPanel.SetBorder(true).SetTitle(" History ").SetBorderPadding(1, 1, 1, 1)
	historyListPanel.SetSelectedFocusOnly(true).SetSecondaryTextColor(tcell.ColorDarkGrey)
	refreshHistory()

	currentTablePanel.SetBorders(false).SetSelectable(true, false).SetFixed(1, 0)
	currentTablePanel.SetBorder(true).SetTitle(fmt.Sprintf(" Current Instances (%s) ", profile)).SetBorderPadding(1, 1, 1, 1)
	currentTablePanel.SetCell(0, 0, tview.NewTableCell("Name").SetAlign(tview.AlignCenter).SetSelectable(false).SetExpansion(1))
	currentTablePanel.SetCell(0, 1, tview.NewTableCell("ID").SetAlign(tview.AlignCenter).SetSelectable(false).SetExpansion(1))
	currentTablePanel.SetCell(0, 2, tview.NewTableCell("Platform").SetAlign(tview.AlignCenter).SetSelectable(false).SetExpansion(1))
	currentTablePanel.SetCell(0, 3, tview.NewTableCell("Type").SetAlign(tview.AlignCenter).SetSelectable(false).SetExpansion(1))

	currentEc2, _ = ListInstances()

	for i, instance := range currentEc2 {
		currentTablePanel.SetCell(i+1, 0, tview.NewTableCell(instance.Name))
		currentTablePanel.SetCell(i+1, 1, tview.NewTableCell(instance.ID))
		currentTablePanel.SetCell(i+1, 2, tview.NewTableCell(instance.Platform))
		currentTablePanel.SetCell(i+1, 3, tview.NewTableCell(instance.Type))
	}

	focusPanel = currentTablePanel
	form := tview.NewFrame(tview.NewFlex().
		AddItem(currentTablePanel, 0, 7, true).
		AddItem(historyListPanel, 0, 3, false)).
		AddText("Press Enter to SSM into instance, '0-9' to SSM into favorite, Tab to switch focus, Q/ESC to quit", false, tview.AlignCenter, tcell.ColorWhite)

	app.SetRoot(form, true)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if slices.Contains([]tcell.Key{tcell.KeyCtrlC, tcell.KeyEsc}, event.Key()) ||
			slices.Contains([]rune{'q', 'Q'}, event.Rune()) {
			exitApp()
		} else if event.Rune() >= '0' && event.Rune() <= '9' {
			if int(event.Rune()-'0') < len(historyEc2.Items) {
				selectedEC2Name = historyEc2.Items[event.Rune()-'0'].Name
				selectEC2ID = historyEc2.Items[event.Rune()-'0'].ID
				selectedEC2Profile = historyEc2.Items[event.Rune()-'0'].Profile
				runSSM()
			}
		} else if event.Key() == tcell.KeyTab {
			SwitchFocus()
		}
		return event
	})

	currentTablePanel.SetSelectedFunc(func(row, column int) {
		selectedEC2Name = currentEc2[row-1].Name
		selectEC2ID = currentEc2[row-1].ID
		selectedEC2Profile = profile
		runSSM()
	})

	historyListPanel.SetSelectedFunc(func(row int, _ string, _ string, _ rune) {
		selectedEC2Name = historyEc2.Items[row].Name
		selectEC2ID = historyEc2.Items[row].ID
		selectedEC2Profile = historyEc2.Items[row].Profile
		runSSM()
	})

	app.Run()
}
