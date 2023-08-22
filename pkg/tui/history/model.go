package history

import (
	"strings"

	"github.com/DomBlack/bubble-shell/internal/config"
	. "github.com/DomBlack/bubble-shell/pkg/modelid"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	// Limit is the maximum number of history items to store
	Limit = 1000
)

// Model represents the main history model
type Model struct {
	id            ID             // The ID of this instance of the history UI
	cfg           *config.Config // The config for this shell
	width, height int            // The width and height of the space we're given to render in

	Scrollback int    // The number of lines to scroll back
	Items      []Item // The history items we're currently displaying
}

// New creates a new history model
func New(cfg *config.Config) Model {
	return Model{
		id:  Next(),
		cfg: cfg,
	}
}

// Init implements tea.Model init function
func (m Model) Init() tea.Cmd {
	return m.ReadHistory()
}

// Update implements tea.Model update function
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {

	switch msg := msg.(type) {

	// Update our width and height
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	// When the history file has been read we can update the model
	case readCompletedMsg:
		if m.id.Matches(msg) {
			m.Items = msg.Items

			// Initialise all the items
			var cmds []tea.Cmd
			for _, item := range m.Items {
				cmds = append(cmds, item.Init())
			}

			return m, tea.Batch(cmds...)
		}

	// When a new item is added we can update the model
	// and then save the history
	case addItemMsg:
		if m.id.Matches(msg) {
			m.Items = append(m.Items, msg.Item)

			return m, tea.Batch(
				msg.Item.Init(),
				m.SaveHistory(m.Items),
			)
		}

	// Update an item in the history updates the given
	// item and then saves the history
	case updateItemMsg:
		if m.id.Matches(msg) {
			var cmds []tea.Cmd

			// If we're an inline shell and the item is no longer running
			// then we need to print it outside the bubbletea managed area
			// so the normal shell scrollbar will work
			if m.cfg.InlineShell && msg.Item.Status > RunningStatus {
				// Mark it as loaded history so we wont render it within the bubble tea program now
				msg.Item.LoadedHistory = true
				cmds = append(cmds, tea.Println(msg.Item.View(m.cfg, m.width)))
			}

			// Start searching from the end of the slice as
			// 99 times out of 100 we'll be updating the most
			// recent item
			found := false
			if len(m.Items) > 0 {
				for i := len(m.Items) - 1; i >= 0; i-- {
					if m.Items[i].ID == msg.Item.ID {
						m.Items[i] = msg.Item
						found = true
						break
					}
				}
			}

			// If we didn't find it then we need to add it
			if !found {
				cmds = append(cmds, m.AppendItem(msg.Item))
			} else {
				cmds = append(cmds, m.SaveHistory(m.Items))
			}

			return m, tea.Batch(cmds...)
		}
	}

	return m, nil
}

// View implements tea.Model view function
func (m Model) View() string {
	lineRender := lipgloss.NewStyle().Width(m.width)

	// If we're inline, just render everything all at once - except for loaded history
	if m.cfg.InlineShell {
		lines := make([]string, 0, len(m.Items))
		for _, item := range m.Items {
			if !item.LoadedHistory {
				lines = append(lines, lineRender.Render(item.View(m.cfg, m.width)))
			}
		}

		return lipgloss.JoinVertical(lipgloss.Left, lines...)
	}

	// Render the lines in reverse order
	// But we only want a maximum of m.height lines
	lineCount := 0
	getLines := m.height + m.Scrollback
	lines := make([]string, getLines)
renderLoop:
	for i := len(m.Items) - 1; i >= 0; i-- {
		item := lineRender.Render(m.Items[i].View(m.cfg, m.width))
		itemLines := strings.Split(item, "\n")

		for j := len(itemLines) - 1; j >= 0; j-- {
			lines[lineCount] = itemLines[j]
			lineCount++
			if lineCount >= getLines {
				break renderLoop
			}
		}
	}

	// Reverse the lines so they're in the correct order
	for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
		lines[i], lines[j] = lines[j], lines[i]
	}

	// Only keep the last m.height lines
	if len(lines) > m.height {
		lines = lines[len(lines)-m.height:]
	}

	// Join all the lines together
	output := lipgloss.JoinVertical(lipgloss.Left, lines...)

	// Render the output with the correct height
	return lipgloss.NewStyle().Height(m.height).Render(output)
}

// Lookback returns the item that is lookback items back in the history
//
// Valid range for lookback is 1 to len(history.Items)
func (m Model) Lookback(lookback int) Item {
	if len(m.Items) == 0 || lookback <= 0 || lookback > len(m.Items) {
		return Item{}
	}

	// If the offset is negative then we're looking back
	// from the end of the slice
	return m.Items[len(m.Items)-lookback]
}

// Search searches through the history for the given string starting
// from startIdx moving by delta, it returns the next index that matches the search string,
// or the startIdx if no more matches are found.
//
// Expected call patterns are:
// - `Search("foo", 0, 1)` - search from the start of the history going backwards in time
// - `Search("foo", len(history.Items), -1)` - search from the end of the history going forwards in time
func (m Model) Search(search string, startIdx int, delta int) (foundIdx int, found bool) {
	search = strings.ToLower(strings.TrimSpace(search))

	if len(m.Items) == 0 || search == "" {
		return 0, false
	}

	currentResult := m.Lookback(startIdx).Line

	// Start searching from the startIdx
	foundIdx = startIdx
	for {
		foundIdx += delta

		// If we've gone past the end of the slice then we're done
		if foundIdx <= 0 || foundIdx > len(m.Items) {
			return startIdx, false
		}

		// If we've found a match then we're done
		item := m.Lookback(foundIdx)
		if item.ItemType == Command && strings.Contains(strings.ToLower(item.Line), search) && item.Line != currentResult {
			return foundIdx, true
		}
	}
}

// AppendItem adds a new item to the history
func (m Model) AppendItem(item Item) tea.Cmd {
	return func() tea.Msg {
		return addItemMsg{
			ID:   m.id,
			Item: item,
		}
	}
}

// UpdateItem updates an item in the history
//
// If the item doesn't exist in the history it will be added
//
// DO NOT rely on this to add items to the history. Use [AppendItem]
// instead as that will be far more efficient on large histories
func (m Model) UpdateItem(item Item) tea.Cmd {
	return func() tea.Msg {
		return updateItemMsg{
			ID:   m.id,
			Item: item,
		}
	}
}

type addItemMsg struct {
	ID   ID
	Item Item
}

func (msg addItemMsg) ForModelID() ID {
	return msg.ID
}

type updateItemMsg struct {
	ID   ID
	Item Item
}

func (msg updateItemMsg) ForModelID() ID {
	return msg.ID
}
