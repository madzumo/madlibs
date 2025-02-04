package menulist

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	digitalColorFront   = "39"
	awsColorFront       = "220"
	headerColorFront    = "46"
	manifestColorFront  = "39"
	batchTagColor       = "197"
	lipTitleStyle       = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color("205"))
	itemStyle           = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle   = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle     = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle           = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	textPromptColor     = "141" //"100" //nice: 141
	textInputColor      = "193" //"40" //nice: 193
	textErrorColorBack  = "1"
	textErrorColorFront = "15"
	textResultJob       = "141" //PINK"205"
	textJobOutcomeFront = "216"

	menuTOP = []string{
		"TAG deployment",
		"DEPLOY Boxes",
		"CREATE Post Launch Scripts",
		"RUN Post Launch URLs",
		"VERIFY Boxes (TightVNC)",
		"DELETE Boxes",
		"Toggle Provider",
		"Enter API Token",
		"Enter AWS Key",
		"Enter AWS Secret",
		"Set Region to deploy",
		"Set # of Boxes to deploy",
		"Set URL Post Launch",
		"Save Settings",
	}
)

// App States
type MenuState int

const (
	StateMainMenu MenuState = iota
	StateSettingsMenu
	StateResultDisplay
	StateSpinner
	StateTextInput
)

// Messsage returend when the background job finishes
type backgroundJobMsg struct {
	result string
}

// // message returned when you have to continue the prompting of data
//
//	type continueJobs struct {
//		result string
//	}
type JobList int

type MenuList struct {
	list                list.Model
	choice              string
	header              string
	state               MenuState
	prevState           MenuState
	prevMenuState       MenuState
	spinner             spinner.Model
	spinnerMsg          string
	backgroundJobResult string
	textInput           textinput.Model
	inputPrompt         string
	textInputError      bool
	jobOutcome          string
	app                 *applicationMain
}

func (m MenuList) Init() tea.Cmd {
	return nil
}

func (m MenuList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case StateMainMenu:
		return m.updateMainMenu(msg)
	case StateSpinner:
		return m.updateSpinner(msg)
	case StateTextInput:
		return m.updateTextInput(msg)
	case StateResultDisplay:
		return m.updateResultDisplay(msg)
	default:
		return m, nil
	}
}

func (m *MenuList) updateMainMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// case tea.MouseMsg:
	// 	if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
	// 		err := clipboard.WriteAll(m.headerIP)
	// 		if err != nil {
	// 			fmt.Println("Failed to copy to clipboard:", err)
	// 		}
	// 	}
	// 	return m, nil
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c", "Q":
			return m, tea.Quit
		case "r", "R":
			m.header = m.app.getAppHeader()
			return m, nil
		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i)
				switch m.choice {
				case menuTOP[6]:
					if m.app.Provider == "digital" {
						m.app.Provider = "aws"
						manifestColorFront = awsColorFront
					} else if m.app.Provider == "aws" {
						m.app.Provider = "digital"
						manifestColorFront = digitalColorFront
					}

					m.header = m.app.getAppHeader()
					return m, nil
				case menuTOP[7]:
					m.prevMenuState = m.state
					m.prevState = m.state
					m.state = StateTextInput
					m.inputPrompt = menuTOP[7]
					m.textInput = textinput.New()
					m.textInput.Placeholder = "e.g., dop_v1_a0xx"
					m.textInput.Focus()
					m.textInput.CharLimit = 200
					m.textInput.Width = 200
					m.textInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor))
					m.textInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textInputColor))
					return m, nil
				case menuTOP[8]:
					m.prevMenuState = m.state
					m.prevState = m.state
					m.state = StateTextInput
					m.inputPrompt = menuTOP[8]
					m.textInput = textinput.New()
					m.textInput.Placeholder = "e.g., AKIAYxx"
					m.textInput.Focus()
					m.textInput.CharLimit = 200
					m.textInput.Width = 200
					m.textInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor))
					m.textInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textInputColor))
					return m, nil
				case menuTOP[9]:
					m.prevMenuState = m.state
					m.prevState = m.state
					m.state = StateTextInput
					m.inputPrompt = menuTOP[9]
					m.textInput = textinput.New()
					m.textInput.Placeholder = "e.g., it's a secret.."
					m.textInput.Focus()
					m.textInput.CharLimit = 200
					m.textInput.Width = 200
					m.textInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor))
					m.textInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textInputColor))
					return m, nil
				case menuTOP[12]:
					m.prevMenuState = m.state
					m.prevState = m.state
					m.state = StateTextInput
					m.inputPrompt = menuTOP[12]
					m.textInput = textinput.New()
					m.textInput.Placeholder = "e.g., https://www.whatever.com"
					m.textInput.Focus()
					m.textInput.CharLimit = 200
					m.textInput.Width = 200
					m.textInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor))
					m.textInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textInputColor))
					return m, nil
				case menuTOP[10]:
					m.prevMenuState = m.state
					m.prevState = m.state
					m.state = StateTextInput
					m.inputPrompt = menuTOP[10]
					m.textInput = textinput.New()
					m.textInput.Placeholder = "e.g., us-east-1"
					m.textInput.Focus()
					m.textInput.CharLimit = 200
					m.textInput.Width = 200
					m.textInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor))
					m.textInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textInputColor))
					return m, nil
				case menuTOP[11]:
					m.prevMenuState = m.state
					m.prevState = m.state
					m.state = StateTextInput
					m.inputPrompt = menuTOP[11]
					m.textInput = textinput.New()
					m.textInput.Placeholder = "e.g., 5"
					m.textInput.Focus()
					m.textInput.CharLimit = 10
					m.textInput.Width = 10
					m.textInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor))
					m.textInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textInputColor))
					return m, nil
				case menuTOP[0]:
					m.prevMenuState = m.state
					m.prevState = m.state
					m.state = StateTextInput
					m.inputPrompt = menuTOP[0]
					m.textInput = textinput.New()
					m.textInput.Placeholder = "e.g., TAG you're it"
					m.textInput.Focus()
					m.textInput.CharLimit = 10
					m.textInput.Width = 10
					m.textInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor))
					m.textInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textInputColor))
					return m, nil
				case menuTOP[1]:
					m.prevState = m.state
					m.prevMenuState = m.state
					m.state = StateSpinner
					return m, tea.Batch(m.spinner.Tick, m.backgroundJobCreateBox())
				case menuTOP[2]:
					m.prevState = m.state
					m.prevMenuState = m.state
					m.state = StateSpinner
					return m, tea.Batch(m.spinner.Tick, m.backgroundJobPS1scripts())
				case menuTOP[3]:
					m.prevState = m.state
					m.prevMenuState = m.state
					m.state = StateSpinner
					return m, tea.Batch(m.spinner.Tick, m.backgroundJobRunPostURL())
					// m.prevMenuState = m.state
					// m.prevState = m.state
					// m.state = StateTextInput
					// m.inputPrompt = menuTOP[6]
					// m.textInput = textinput.New()
					// m.textInput.Placeholder = "e.g., Type All or box #"
					// m.textInput.Focus()
					// m.textInput.CharLimit = 10
					// m.textInput.Width = 10
					// m.textInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor))
					// m.textInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textInputColor))
					// return m, nil
				case menuTOP[4]:
					m.prevState = m.state
					m.prevMenuState = m.state
					m.state = StateSpinner
					return m, tea.Batch(m.spinner.Tick, m.backgroundJobVerifyVNC())
				case menuTOP[5]:
					m.prevState = m.state
					m.prevMenuState = m.state
					m.state = StateSpinner
					return m, tea.Batch(m.spinner.Tick, m.backgroundJobDeleteBox())
				case menuTOP[13]:
					m.prevState = m.state
					m.prevMenuState = m.state
					m.state = StateSpinner
					return m, tea.Batch(m.spinner.Tick, m.backgroundSaveSettings())
				}
			}
			return m, nil
		}
		// case jobListMsg:

		// 	// m.state = StateResultDisplay
		// 	// return m, nil
		// 	m.prevState = m.state
		// 	m.state = StateSpinner
		// 	return m, tea.Batch(m.spinner.Tick, m.startBackgroundJob())
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *MenuList) updateTextInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	m.textInputError = false
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			inputValue := m.textInput.Value() // User pressed enter, save the input

			switch m.inputPrompt {
			case menuTOP[7]:
				m.app.Digital.ApiToken = inputValue
				m.backgroundJobResult = fmt.Sprintf("Saved API: %s", inputValue)
				m.header = m.app.getAppHeader()
			case menuTOP[8]:
				m.app.Aws.Key = inputValue
				m.backgroundJobResult = fmt.Sprintf("Saved AWS Key: %s", inputValue)
				m.header = m.app.getAppHeader()
			case menuTOP[9]:
				m.app.Aws.Secret = inputValue
				m.backgroundJobResult = fmt.Sprintf("Saved AWS Secret: %s", inputValue)
				m.header = m.app.getAppHeader()
			case menuTOP[12]:
				m.app.URL = inputValue
				m.backgroundJobResult = fmt.Sprintf("Saved URL: %s", inputValue)
				m.header = m.app.getAppHeader()
			case menuTOP[11]:
				boxes, err := strconv.Atoi(inputValue)
				if err != nil {
					m.backgroundJobResult = "Data inputed is not a valid Number"
				} else {
					m.app.NumberBoxes = boxes
					m.backgroundJobResult = fmt.Sprintf("Number of Boxes = %s", inputValue)
					m.header = m.app.getAppHeader()
				}
			case menuTOP[10]:
				if m.app.Provider == "digital" {
					m.app.Digital.Region = inputValue
				} else { //AWS
					m.app.Aws.Region = inputValue
				}
				m.backgroundJobResult = fmt.Sprintf("Saved Region: %s", inputValue)
				m.header = m.app.getAppHeader()
			case menuTOP[0]:
				m.app.BatchTag = inputValue
				m.backgroundJobResult = fmt.Sprintf("Saved Batch Tag: %s", inputValue)
				m.header = m.app.getAppHeader()
			}
			m.prevState = m.state
			m.state = StateResultDisplay
			return m, nil

		case tea.KeyEsc:
			// m.state = StateSettingsMenu
			m.state = m.prevState
			return m, nil
		}
	}

	return m, cmd
}

func (m *MenuList) updateSpinner(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// case "q", "esc":
		// 	m.backgroundJobResult = "Job Cancelled"
		// 	m.state = StateResultDisplay
		// 	return m, nil
		default:
			// For other key presses, update the spinner
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	case backgroundJobMsg:
		m.backgroundJobResult = m.jobOutcome + "\n\n" + msg.result + "\n"
		m.state = StateResultDisplay
		return m, nil
	// case continueJobs:
	// 	return m, tea.Batch(m.spinner.Tick, m.startBackgroundJob())
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m *MenuList) updateResultDisplay(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			if m.textInputError {
				m.state = m.prevState
			} else {
				m.state = m.prevMenuState
			}
			m.updateListItems()
			return m, nil
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m MenuList) viewResultDisplay() string {
	outro := "Press 'esc' to return."
	outroRender := lipgloss.NewStyle().Foreground(lipgloss.Color("231")).Bold(true).Render(outro)
	lipgloss.NewStyle().Foreground(lipgloss.Color("231")).Bold(true)
	if m.textInputError {
		m.backgroundJobResult = lipgloss.NewStyle().Foreground(lipgloss.Color(textErrorColorFront)).Background(lipgloss.Color(textErrorColorBack)).Bold(true).Render(m.backgroundJobResult)
	} else {
		m.backgroundJobResult = lipgloss.NewStyle().Foreground(lipgloss.Color(textResultJob)).Render(m.backgroundJobResult)
	}
	return fmt.Sprintf("\n\n%s\n\n%s", m.backgroundJobResult, outroRender)

	// //repeat interval
	// if m.configSettings.Interval > 0 {

	// }
}

func (m MenuList) View() string {
	switch m.state {
	case StateMainMenu, StateSettingsMenu:
		return m.header + "\n" + m.list.View()
	case StateSpinner:
		return m.viewSpinner()
	case StateTextInput:
		return m.viewTextInput()
	case StateResultDisplay:
		return m.viewResultDisplay()
	default:
		return "Unknown state"
	}
}

func (m MenuList) viewSpinner() string {
	// tea.ClearScreen()
	spinnerBase := fmt.Sprintf("\n\n   %s %s\n\n", m.spinner.View(), m.spinnerMsg)

	// return spinnerBase + m.jobOutcome
	return spinnerBase + lipgloss.NewStyle().Foreground(lipgloss.Color(textJobOutcomeFront)).Bold(true).Render(m.jobOutcome)
}

func (m MenuList) viewTextInput() string {
	promptStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor)).Bold(true)
	return fmt.Sprintf("\n\n%s\n\n%s", promptStyle.Render(m.inputPrompt), m.textInput.View())

}

func (m *MenuList) updateListItems() {
	switch m.state {
	case StateMainMenu:
		items := []list.Item{}
		for _, value := range menuTOP {
			items = append(items, item(value))
		}
		m.list.SetItems(items)
		// case StateSettingsMenu:
		// 	items := []list.Item{}
		// 	for _, value := range menuSettings {
		// 		items = append(items, item(value[0]))
		// 	}
		// 	m.list.SetItems(items)
	}

	m.list.ResetSelected()
}

func (m *MenuList) backgroundSaveSettings() tea.Cmd {
	return func() tea.Msg {
		m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("13")) //white = 231

		m.spinnerMsg = "Saving Settings"
		// m.spinner.Tick()
		time.Sleep(1 * time.Second)
		saveSettings(m.app)

		return backgroundJobMsg{result: "Settings Saved"}
	}
}

func (m *MenuList) backgroundJobCreateBox() tea.Cmd {
	return func() tea.Msg {
		m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("82")) //white = 231
		m.spinnerMsg = "Creating Boxes..."
		resultX := fmt.Sprintf("%d - Boxes created!", m.app.NumberBoxes)

		if m.app.Provider == "digital" {
			err1 := m.app.Digital.createFirewall()
			if err1 != nil {
				resultX = fmt.Sprintf("Error creating firewall:\n%s", err1)
			}
			for i := 1; i <= m.app.NumberBoxes; i++ {
				err := m.app.Digital.createBox()
				if err != nil {
					resultX = fmt.Sprintf("error creating box:\n%s", err)
				}
				// time.Sleep(1 * time.Second)
			}
		} else { //aws
			pepa, err := m.app.Aws.createEc2Client()
			if err != nil {
				resultX = fmt.Sprintf("error getting AWS credentials:\n%s", err)
			} else {
				err = m.app.Aws.createPEMFile(pepa)
				if err != nil {
					resultX = fmt.Sprintf("error creating PEM:\n%s", err)
				} else {
					sgAuto, err2 := m.app.Aws.createSecurityGroup("sgAutoBox", "pepita stuff", pepa)
					if err2 != nil {
						resultX = fmt.Sprintf("error creating Security Group:\n%s", err2)
					} else {
						for i := 1; i <= m.app.NumberBoxes; i++ {
							m.app.Aws.createEC2Instance(sgAuto, pepa, m.app.BatchTag)
						}
					}
				}

			}
		}

		return backgroundJobMsg{result: resultX}
	}
}

func (m *MenuList) backgroundJobRunPostURL() tea.Cmd {
	return func() tea.Msg {
		var wg sync.WaitGroup
		m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("82")) //white = 231
		m.spinnerMsg = "Running Post Launch Scripts"
		result := "Finished Executing Post Launch Scripts"
		scriptsFolder := fmt.Sprintf("./%s", m.app.Digital.Region)
		if m.app.Provider == "aws" {
			scriptsFolder = fmt.Sprintf("./%s", m.app.Aws.Region)
		}

		files, err := os.ReadDir(scriptsFolder)
		if err != nil {
			result = fmt.Sprintf("Error executing scripts:\n%s", err)
		}

		// Loop through each .ps1 file and execute it
		batchMatch := true
		for _, file := range files {
			if m.app.BatchTag != "" {
				batchMatch = strings.Contains(file.Name(), m.app.BatchTag)
			}
			delta := 0
			if filepath.Ext(file.Name()) == ".ps1" && batchMatch {
				delta++
				wg.Add(delta)
				go func() {
					defer wg.Done()
					scriptPath, _ := filepath.Abs(filepath.Join(scriptsFolder, file.Name()))
					m.app.runPS1file(scriptPath, file.Name())
				}()
			}
		}

		wg.Wait()
		return backgroundJobMsg{result: result}
	}
}

func (m *MenuList) backgroundJobPS1scripts() tea.Cmd {
	return func() tea.Msg {
		m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("82")) //white = 231
		m.spinnerMsg = "Creating Post Launch scripts..."
		result := "Created Post Launch scripts"

		if m.app.Provider == "digital" {
			ips, err := m.app.Digital.compileIPaddressesDigital()
			if err != nil {
				result = fmt.Sprintf("Error compiling IP addresses:\n%s", err)
			} else {
				for _, ip := range ips {
					err := m.app.createPostSCRIPT(ip, "")
					if err != nil {
						result = fmt.Sprintf("Error creating post script\n%s", err)
					}
				}
			}
		} else { //AWS
			pepa, _ := m.app.Aws.createEc2Client()
			ips, _, err := m.app.Aws.compileIPaddressesAws(pepa, m.app.BatchTag)
			if err != nil {
				result = fmt.Sprintf("Error compiling IP addresses:\n%s", err)
			} else {
				for _, ip := range ips {
					err := m.app.createPostSCRIPT(ip, m.app.Aws.PemKeyFileName)
					if err != nil {
						result = fmt.Sprintf("Error creating post script\n%s", err)
					}
				}
			}
		}

		return backgroundJobMsg{result: result}
	}
}

func (m *MenuList) backgroundJobDeleteBox() tea.Cmd {
	return func() tea.Msg {
		m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("82")) //white = 231
		m.spinnerMsg = "Deleting Boxes"
		resultX := "Boxes & Related Resources Deleted!"

		if m.app.Provider == "digital" {
			err := m.app.Digital.deleteBox()
			if err != nil {
				resultX = fmt.Sprintf("Error deleting droplets\n%s", err)
			}
			err = m.app.Digital.deleteFirewall()
			if err != nil {
				resultX = fmt.Sprintf("Error deleting firewall\n%s", err)
			}
		} else { //aws
			pepa, err := m.app.Aws.createEc2Client()
			if err != nil {
				resultX = fmt.Sprintf("error getting AWS credentials:\n%s", err)
			} else {
				err = m.app.Aws.deleteEC2Instances(pepa, m.app.BatchTag)
				if err != nil {
					resultX = fmt.Sprintf("%s\n%s", err, resultX)
				}
				if m.app.BatchTag == "" {
					// err = m.app.Aws.deleteSecurityGroups(pepa)
					// if err != nil {
					// 	resultX = fmt.Sprintf("%s\n%s", err, resultX)
					// }
					err = m.app.Aws.deletePEMFile(pepa)
					if err != nil {
						resultX = fmt.Sprintf("%s\n%s", err, resultX)
					}
				}
			}
		}

		scriptsFolder := fmt.Sprintf("./%s", m.app.Digital.Region)
		if m.app.Provider == "aws" {
			scriptsFolder = fmt.Sprintf("./%s", m.app.Aws.Region)
		}
		entries, err := os.ReadDir(scriptsFolder)
		if err != nil {
			resultX = fmt.Sprintf("Failed to clear scripts folder\n%s", err)
		}
		for _, entry := range entries {
			if m.app.BatchTag == "" {
				if !entry.IsDir() && (filepath.Ext(entry.Name()) == ".ps1" || filepath.Ext(entry.Name()) == ".pem") {
					err = os.Remove(filepath.Join(scriptsFolder, entry.Name()))
					if err != nil {
						resultX = fmt.Sprintf("Failed to clear scripts folder\n%s", err)
					}
				}
			} else {
				if !entry.IsDir() && filepath.Ext(entry.Name()) == ".ps1" && strings.Contains(entry.Name(), m.app.BatchTag) {
					err = os.Remove(filepath.Join(scriptsFolder, entry.Name()))
					if err != nil {
						resultX = fmt.Sprintf("Failed to clear scripts folder\n%s", err)
					}
				}
			}
		}

		return backgroundJobMsg{result: resultX}
	}
}

func (m *MenuList) backgroundJobVerifyVNC() tea.Cmd {
	return func() tea.Msg {
		m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("82")) //white = 231
		m.spinnerMsg = "Verify with TightVNC"
		// fmt.Println("started job")
		result := "Verified mofo!"

		if m.app.Provider == "digital" {
			scriptsFolder := fmt.Sprintf("./%s", m.app.Digital.Region)
			files, _ := os.ReadDir(scriptsFolder)
			ips, err := m.app.Digital.compileIPaddressesDigital()
			if err != nil {
				result = fmt.Sprintf("Error compiling IP addresses:\n%s", err)
			} else {
				for _, ip := range ips {
					for _, file := range files {
						if strings.Contains(file.Name(), m.app.BatchTag) &&
							strings.Contains(file.Name(), ip) {
							err := m.app.runVNC(ip)
							if err != nil {
								result = fmt.Sprintf("Error running TightVNC\n%s", err)
							}
						}
					}
				}
			}
		} else { //aws
			scriptsFolder := fmt.Sprintf("./%s", m.app.Aws.Region)
			files, _ := os.ReadDir(scriptsFolder)
			// fmt.Println(m.app.Aws.Region) //troubleshoot
			// time.Sleep(100 * time.Second) //troubleshoot
			pepa, _ := m.app.Aws.createEc2Client()
			ips, _, err := m.app.Aws.compileIPaddressesAws(pepa, m.app.BatchTag)
			if err != nil {
				result = fmt.Sprintf("Error compiling IP addresses:\n%s", err)
			} else {
				for _, ip := range ips {
					for _, file := range files {
						if strings.Contains(file.Name(), m.app.BatchTag) &&
							strings.Contains(file.Name(), ip) {
							err := m.app.runVNC(ip)
							if err != nil {
								result = fmt.Sprintf("Error running TightVNC\n%s", err)
							}
						}
					}
				}
			}
		}

		return backgroundJobMsg{result: result}
	}
}
func ShowMenu(app *applicationMain) {

	const listWidth = 90
	const listHeight = 14

	// Initialize the list with empty items; items will be set in updateListItems
	l := list.New([]list.Item{}, itemDelegate{}, listWidth, listHeight)
	l.Title = ""
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowTitle(true)
	l.Styles.Title = lipTitleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle
	l.KeyMap.ShowFullHelp = key.NewBinding() // remove '?' help option

	s := spinner.New()
	s.Spinner = spinner.Pulse

	m := MenuList{
		list:       l,
		header:     app.getAppHeader(),
		state:      StateMainMenu,
		spinner:    s,
		spinnerMsg: "Action Performing",
		app:        app,
	}

	m.updateListItems()

	m.list.KeyMap.Quit = key.NewBinding(
		key.WithKeys("esc", "ctrl+c"),
		key.WithHelp("esc", "quit"),
	)

	//show Menu
	_, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}
