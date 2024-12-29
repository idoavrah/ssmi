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
	selectedUsername   string
	pages              *tview.Pages
	primaryScreen      *tview.Frame
	modalScreen        *tview.Flex
)

func init() {
	homedir, _ := os.UserHomeDir()
	historyFilename = filepath.Join(homedir, ".ssmi_history")
	app = tview.NewApplication()
	currentEc2 = []Instance{}
	historyEc2 = LoadHistoryList(historyFilename)
	historyListPanel = tview.NewList()
	currentTablePanel = tview.NewTable()
	focusPanel = currentTablePanel
	selectedUsername = ""
}

func exitApp() {
	app.Stop()
	historyEc2.Save(historyFilename)
	os.Exit(0)
}

func runSSM() {

	app.Stop()
	addToHistory()

	params := []string{"ssm", "start-session", "--target", selectEC2ID, "--profile", selectedEC2Profile}
	if selectedUsername != "" {
		params = append(params,
			"--document-name", "AWS-StartInteractiveCommand", "--parameters",
			fmt.Sprintf("command=\"sudo su - %s\"", selectedUsername))
	}

	fmt.Printf("\nInstance: %s@%s (%s)\nProfile:  %s\n", selectedUsername, selectedEC2Name, selectEC2ID, selectedEC2Profile)

	cmd := exec.Command("aws", params...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	cmd.Run()
	exitApp()
}

func refreshHistory() {
	historyListPanel.Clear()
	var name string
	for idx, item := range historyEc2.Items {
		if item.Name == "" {
			name = item.ID
		} else {
			name = item.Name
		}
		historyListPanel.AddItem(fmt.Sprintf("Instance: %s@%s", item.Username, name), fmt.Sprintf("Profile:  %s", item.Profile), '0'+rune(idx), nil)
	}
}

func addToHistory() {
	historyEc2.Add(HistoryItem{
		ID:       selectEC2ID,
		Name:     selectedEC2Name,
		Username: selectedUsername,
		Profile:  selectedEC2Profile,
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
		fmt.Println("AWS_PROFILE is not set")
		exitApp()
	}

	historyListPanel.SetBorder(true).SetTitle(" History ").SetBorderPadding(1, 1, 1, 1)
	historyListPanel.SetSelectedFocusOnly(true).SetSecondaryTextColor(tcell.ColorDarkGrey)
	refreshHistory()

	currentTablePanel.SetBorders(false).SetSelectable(true, false).SetFixed(1, 0)
	currentTablePanel.SetBorder(true).SetTitle(fmt.Sprintf(" Current Instances (Profile: %s) ", profile)).SetBorderPadding(1, 1, 1, 1)
	currentTablePanel.SetCell(0, 0, tview.NewTableCell("Name").SetAlign(tview.AlignCenter).SetSelectable(false).SetExpansion(1))
	currentTablePanel.SetCell(0, 1, tview.NewTableCell("ID").SetAlign(tview.AlignCenter).SetSelectable(false).SetExpansion(1))
	currentTablePanel.SetCell(0, 2, tview.NewTableCell("Platform").SetAlign(tview.AlignCenter).SetSelectable(false).SetExpansion(1))
	currentTablePanel.SetCell(0, 3, tview.NewTableCell("Type").SetAlign(tview.AlignCenter).SetSelectable(false).SetExpansion(1))

	var err error
	currentEc2, err = ListInstances(profile)
	if err != nil {
		fmt.Println(err)
		exitApp()
	}

	var color tcell.Color
	for i, instance := range currentEc2 {
		if !instance.Supported {
			color = tcell.ColorDarkGrey
		} else {
			color = tcell.ColorWhite
		}
		currentTablePanel.SetCell(i+1, 0, tview.NewTableCell(instance.Name).SetTextColor(color))
		currentTablePanel.SetCell(i+1, 1, tview.NewTableCell(instance.ID).SetTextColor(color))
		currentTablePanel.SetCell(i+1, 2, tview.NewTableCell(instance.Platform).SetTextColor(color))
		currentTablePanel.SetCell(i+1, 3, tview.NewTableCell(instance.Type).SetTextColor(color))
	}

	primaryScreen = tview.NewFrame(tview.NewFlex().
		AddItem(currentTablePanel, 0, 7, true).
		AddItem(historyListPanel, 0, 3, false)).
		AddText("Press Enter to SSM into instance, 0-9 to SSM into recent instance, Tab to switch focus, Q/ESC to quit", false, tview.AlignCenter, tcell.ColorWhite)

	userForm := tview.NewForm()
	userForm.SetBorderPadding(1, 1, 1, 1)
	userForm.SetBorder(true).SetTitle(" Select username to use ").SetTitleAlign(tview.AlignCenter)

	inputField := userForm.AddInputField("username", "", 20, nil, nil)
	inputField.SetCancelFunc(func() {
		pages.HidePage("modal")
		app.SetFocus(currentTablePanel)
		focusPanel = currentTablePanel
	})

	modalScreen = tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(nil, 0, 2, false).
			AddItem(userForm, 7, 1, true).
			AddItem(nil, 0, 2, false), 0, 1, true).
		AddItem(nil, 0, 1, false)

	pages = tview.NewPages().
		AddPage("primary", primaryScreen, true, true).
		AddPage("modal", modalScreen, true, false)

	app.SetRoot(pages, true)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		page, _ := pages.GetFrontPage()
		if event.Key() == tcell.KeyCtrlC || slices.Contains([]rune{'q', 'Q'}, event.Rune()) {
			exitApp()
		} else if event.Key() == tcell.KeyEnter && page == "modal" {
			selectedUsername = userForm.GetFormItemByLabel("username").(*tview.InputField).GetText()
			runSSM()
		} else if event.Rune() >= '0' && event.Rune() <= '9' {
			if int(event.Rune()-'0') < len(historyEc2.Items) {
				item := historyEc2.Items[event.Rune()-'0']
				selectedEC2Name = item.Name
				selectEC2ID = item.ID
				selectedEC2Profile = item.Profile
				selectedUsername = item.Username
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
		userForm.GetFormItemByLabel("username").(*tview.InputField).SetText("")
		pages.ShowPage("modal")
	})

	historyListPanel.SetSelectedFunc(func(row int, _ string, _ string, _ rune) {
		selectedEC2Name = historyEc2.Items[row].Name
		selectEC2ID = historyEc2.Items[row].ID
		selectedEC2Profile = historyEc2.Items[row].Profile
		runSSM()
	})

	app.Run()
}
