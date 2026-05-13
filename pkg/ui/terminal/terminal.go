/*
 *
 * Copyright 2026 perrault authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package terminal

import (
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	key "github.com/charmbracelet/bubbles/key"
	list "github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss"
	ui "github.com/dvaumoron/perrault/pkg/ui"
)

const (
	defaultWidth      = 20
	defaultListHeight = 14
	selectedColorCode = "170"
)

const (
	noAction actionKind = iota
	quitAction
	selectAction
)

var _ ui.UI = &UI{}

type UI struct {
	Width             int
	ListHeight        int
	TitleStyle        lipgloss.Style
	SelectedItemStyle lipgloss.Style
	PaginationStyle   lipgloss.Style
	HelpStyle         lipgloss.Style
}

func NewUI() *UI {
	defaultStyles := list.DefaultStyles()
	return &UI{
		Width:             defaultWidth,
		ListHeight:        defaultListHeight,
		TitleStyle:        lipgloss.NewStyle(),
		SelectedItemStyle: lipgloss.NewStyle().Foreground(lipgloss.Color(selectedColorCode)),
		PaginationStyle:   defaultStyles.PaginationStyle,
		HelpStyle:         defaultStyles.HelpStyle,
	}
}

func (ui *UI) AskUserChoice(title string, choices []string) int {
	wrappedChoices := make([]list.Item, len(choices))
	for i, choice := range choices {
		wrappedChoices[i] = item(choice)
	}

	delegate := itemDelegate{
		SelectedItemStyle: ui.SelectedItemStyle,
	}

	displayList := list.New(wrappedChoices, delegate, ui.Width, ui.ListHeight)
	displayList.Title = title
	displayList.SetShowStatusBar(false)
	displayList.SetFilteringEnabled(false)
	displayList.Styles.Title = ui.TitleStyle
	displayList.Styles.PaginationStyle = ui.PaginationStyle
	displayList.Styles.HelpStyle = ui.HelpStyle

	displayList.AdditionalFullHelpKeys = additionalFullHelpKeys
	displayList.AdditionalShortHelpKeys = additionalShortHelpKeys

	selector := itemSelector{
		list: displayList,
	}

	_, err := tea.NewProgram(&selector).Run()
	if err != nil || selector.action == quitAction {
		return -1
	}

	choice := selector.list.SelectedItem().FilterValue()
	fmt.Println(choice)
	return slices.Index(choices, choice)
}

type actionKind uint8

type item string

func (i item) FilterValue() string {
	return string(i)
}

type itemDelegate struct {
	SelectedItemStyle lipgloss.Style
}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(writer io.Writer, displayList list.Model, index int, listItem list.Item) {
	var builder strings.Builder
	builder.WriteString(strconv.Itoa(index + 1))
	builder.WriteByte('.')
	builder.WriteByte(' ')
	builder.WriteString(listItem.FilterValue())
	line := builder.String()
	if index == displayList.Index() {
		line = d.SelectedItemStyle.Render(line)
	}

	io.WriteString(writer, line)
}

type itemSelector struct {
	list   list.Model
	action actionKind
}

func (m *itemSelector) Init() tea.Cmd {
	return nil
}

func (m *itemSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)

		return m, nil
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "esc", "q":
			m.action = quitAction
			return m, tea.Quit
		case "enter", " ":
			m.action = selectAction
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)

	return m, cmd
}

func (m itemSelector) View() string {
	if m.action == noAction {
		return "\n" + m.list.View()
	}

	return ""
}

func additionalFullHelpKeys() []key.Binding {
	return []key.Binding{
		key.NewBinding(
			key.WithKeys("space"),
			key.WithHelp("space", "select item"),
		),
		key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select item"),
		),
	}
}

func additionalShortHelpKeys() []key.Binding {
	return []key.Binding{
		key.NewBinding(
			key.WithKeys("space"),
			key.WithHelp("space", "select"),
		),
		key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
	}
}
