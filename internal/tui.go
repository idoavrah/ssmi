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
	favoritesEc2       *FavoritesArray
	historyListPanel   *tview.List
	favoritesListPanel *tview.List
	currentTablePanel  *tview.Table
	profile            string
	historyFilename    string
	favoritesFilename  string
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
	ssmiFolder := filepath.Join(homedir, ".ssmi")
	err := os.Mkdir(ssmiFolder, 0755)
	if err != nil {
		if !os.IsExist(err) {
			fmt.Println("Error creating .ssmi directory in home folder:", err)
			os.Exit(1)
		}
	}

	historyFilename = filepath.Join(ssmiFolder, "history.json")
	favoritesFilename = filepath.Join(ssmiFolder, "favorites.json")

	app = tview.NewApplication()
	currentEc2 = []Instance{}
	historyEc2 = LoadHistoryList(historyFilename)
	favoritesEc2 = LoadFavoritesList(favoritesFilename)
	historyListPanel = tview.NewList()
	favoritesListPanel = tview.NewList()
	currentTablePanel = tview.NewTable()
	selectedUsername = ""
}

func exitApp() {
	app.Stop()
	historyEc2.Save(historyFilename)
	favoritesEc2.Save(favoritesFilename)
	os.Exit(0)
}

func runSSM() {

	app.Stop()

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

	if err := cmd.Run(); err == nil {
		addToHistory()
	}

	exitApp()

}

func addToFavorites(position int) {
	favoritesEc2.Add(FavoriteItem{
		ID:       selectEC2ID,
		Name:     selectedEC2Name,
		Username: selectedUsername,
		Profile:  selectedEC2Profile,
	}, position)
	favoritesEc2.Save(favoritesFilename)
	refreshFavorites()
}

func refreshFavorites() {
	favoritesListPanel.Clear()
	for idx, item := range favoritesEc2.Items {

		var primary, secondary string
		if item.ID != "" {
			primary = fmt.Sprintf("Instance: %s@%s (%s)", item.Username, item.Name, item.ID)
			secondary = fmt.Sprintf("Profile:  %s", item.Profile)
		}
		favoritesListPanel.AddItem(primary, secondary, 'a'+rune(idx), nil)
	}
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
		historyListPanel.AddItem(fmt.Sprintf("Instance: %s@%s (%s)", item.Username, name, item.ID), fmt.Sprintf("Profile:  %s", item.Profile), '0'+rune(idx), nil)
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

func StartApplication() {
	profile = os.Getenv("AWS_PROFILE")
	if profile == "" {
		fmt.Println("AWS_PROFILE is not set")
		exitApp()
	}

	var err error
	currentEc2, err = ListInstances(profile)
	if err != nil {
		fmt.Println(err)
		exitApp()
	}
	if len(currentEc2) > 0 {
		selectedEC2Name = currentEc2[0].Name
		selectEC2ID = currentEc2[0].ID
		selectedEC2Profile = profile
	}

	black := tcell.NewRGBColor(0x00, 0x00, 0x00)
	style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(black)

	historyListPanel.SetBorder(true).SetTitle(" History ").SetBorderPadding(1, 1, 1, 1)
	historyListPanel.SetSelectedFocusOnly(true)
	historyListPanel.SetShortcutStyle(style)
	historyListPanel.SetMainTextStyle(style)
	historyListPanel.SetSecondaryTextStyle(style)
	historyListPanel.SetMainTextColor(tcell.ColorWhite)
	historyListPanel.SetBackgroundColor(black)

	favoritesListPanel.SetBorder(true).SetTitle(" Favorites ").SetBorderPadding(1, 1, 1, 1)
	favoritesListPanel.SetSelectedFocusOnly(true)
	favoritesListPanel.SetShortcutStyle(style)
	favoritesListPanel.SetMainTextStyle(style)
	favoritesListPanel.SetSecondaryTextStyle(style)
	favoritesListPanel.SetMainTextColor(tcell.ColorWhite)
	favoritesListPanel.SetBackgroundColor(black)

	quickPanel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(favoritesListPanel, 0, 1, false).
		AddItem(historyListPanel, 0, 1, false)

	currentTablePanel.SetBorders(false).SetSelectable(true, false).SetFixed(1, 0)
	currentTablePanel.SetBorder(true).SetTitle(fmt.Sprintf(" Current Instances (Profile: %s) ", profile)).SetBorderPadding(1, 1, 1, 1)
	currentTablePanel.SetCell(0, 0, tview.NewTableCell("Name").SetAlign(tview.AlignCenter).SetSelectable(false).SetExpansion(1))
	currentTablePanel.SetCell(0, 1, tview.NewTableCell("ID").SetAlign(tview.AlignCenter).SetSelectable(false).SetExpansion(1))
	currentTablePanel.SetCell(0, 2, tview.NewTableCell("Platform").SetAlign(tview.AlignCenter).SetSelectable(false).SetExpansion(1))
	currentTablePanel.SetCell(0, 3, tview.NewTableCell("Type").SetAlign(tview.AlignCenter).SetSelectable(false).SetExpansion(1))
	currentTablePanel.SetBackgroundColor(black)

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
		AddItem(currentTablePanel, 0, 6, true).
		AddItem(quickPanel, 0, 4, true)).
		AddText("Press Enter to SSM into instance, 0-9 to log into recent instance, a-j to log into favorite instance, A-J to save new favorite, Q/ESC to quit", false, tview.AlignCenter, tcell.ColorWhite)

	primaryScreen.SetBackgroundColor(black)

	userForm := tview.NewForm()
	userForm.SetBorderPadding(1, 1, 1, 1)
	userForm.SetBorder(true).SetTitle(" Select username to use ").SetTitleAlign(tview.AlignCenter)

	inputField := userForm.AddInputField("username", "", 20, nil, nil)
	inputField.SetCancelFunc(func() {
		pages.HidePage("modal")
		app.SetFocus(currentTablePanel)
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

		if page == "modal" && event.Key() == tcell.KeyEnter {
			selectedUsername = userForm.GetFormItemByLabel("username").(*tview.InputField).GetText()
			runSSM()
		} else if page == "primary" {
			pressedKey := event.Key()
			pressedRune := event.Rune()
			if pressedKey == tcell.KeyCtrlC || slices.Contains([]rune{'q', 'Q'}, pressedRune) {
				exitApp()
			} else if pressedRune >= 'A' && pressedRune <= 'J' {
				addToFavorites(int(pressedRune - 'A'))
				return nil
			} else if pressedRune >= 'a' && pressedRune <= 'j' {
				position := int(pressedRune - 'a')
				item := favoritesEc2.Items[position]
				if item.ID != "" {
					selectedEC2Name = item.Name
					selectEC2ID = item.ID
					selectedEC2Profile = item.Profile
					selectedUsername = item.Username
					runSSM()
				}
			} else if pressedRune >= '0' && pressedRune <= '9' {
				if int(pressedRune-'0') < len(historyEc2.Items) {
					item := historyEc2.Items[pressedRune-'0']
					selectedEC2Name = item.Name
					selectEC2ID = item.ID
					selectedEC2Profile = item.Profile
					selectedUsername = item.Username
					runSSM()
				}
			}
		}
		return event
	})

	currentTablePanel.SetSelectionChangedFunc(func(row, column int) {
		selectedEC2Name = currentEc2[row-1].Name
		selectEC2ID = currentEc2[row-1].ID
		selectedEC2Profile = profile
	})

	currentTablePanel.SetSelectedFunc(func(row, column int) {
		userForm.GetFormItemByLabel("username").(*tview.InputField).SetText("")
		pages.ShowPage("modal")
	})

	historyListPanel.SetSelectedFunc(func(row int, _ string, _ string, _ rune) {
		selectedEC2Name = historyEc2.Items[row].Name
		selectEC2ID = historyEc2.Items[row].ID
		selectedEC2Profile = historyEc2.Items[row].Profile
		runSSM()
	})

	refreshHistory()
	refreshFavorites()

	app.Run()
}
