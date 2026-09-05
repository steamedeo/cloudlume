package tui

import (
	"context"
	"sort"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/steamedeo/cloudlume/internal/model"
	"github.com/steamedeo/cloudlume/internal/provider"
)

const appVersion = "v0.1.0"

// sidebarAll is the pseudo-account name for the aggregated "all accounts"
// view, which is always the first sidebar row and the default selection —
// per the product requirement that resources from every account show up
// together, not behind a switcher.
const sidebarAll = "All Accounts"

// Model is cloudlume's root Bubble Tea model.
type Model struct {
	providers []provider.Provider
	refresh   time.Duration

	accounts []model.Account
	statuses map[string]*model.AccountStatus // key: account.Name
	byAccount map[string][]model.Resource     // key: account.Name

	sidebarIndex int // 0 = All Accounts, 1..N = accounts (sorted)
	tabIndex     int
	cursor       int
	paused       bool

	width, height int
	ready         bool
}

// New builds the initial model. providers should already be registered
// (their init() functions ran via import side effects); refresh is the
// polling interval for re-fetching resources.
func New(providers []provider.Provider, refresh time.Duration) Model {
	return Model{
		providers: providers,
		refresh:   refresh,
		statuses:  map[string]*model.AccountStatus{},
		byAccount: map[string][]model.Resource{},
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(discoverAccountsCmd(m.providers), tick(m.refresh))
}

func discoverAccountsCmd(providers []provider.Provider) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		var accounts []model.Account
		for _, p := range providers {
			found, err := p.Accounts(ctx)
			if err != nil {
				continue
			}
			accounts = append(accounts, found...)
		}
		return accountsMsg{accounts: accounts}
	}
}

func fetchAccountCmd(providers []provider.Provider, account model.Account) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()

		for _, p := range providers {
			if p.Name() != account.Provider {
				continue
			}
			res, err := p.FetchResources(ctx, account)
			return resourcesMsg{account: account, resources: res, err: err, fetchedAt: time.Now()}
		}
		return resourcesMsg{account: account, fetchedAt: time.Now()}
	}
}

func tick(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg { return tickMsg{} })
}

func (m Model) sortedAccountNames() []string {
	names := make([]string, 0, len(m.accounts))
	for _, a := range m.accounts {
		names = append(names, a.Name)
	}
	sort.Strings(names)
	return names
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.ready = true
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case accountsMsg:
		m.accounts = msg.accounts
		var cmds []tea.Cmd
		for _, a := range m.accounts {
			m.statuses[a.Name] = &model.AccountStatus{Account: a, Health: model.HealthWarn}
			cmds = append(cmds, fetchAccountCmd(m.providers, a))
		}
		return m, tea.Batch(cmds...)

	case resourcesMsg:
		st := m.statuses[msg.account.Name]
		if st == nil {
			st = &model.AccountStatus{Account: msg.account}
			m.statuses[msg.account.Name] = st
		}
		st.LastRefresh = msg.fetchedAt
		if msg.err != nil {
			st.Health = model.HealthDown
			st.Error = msg.err.Error()
		} else {
			st.Health = model.HealthOK
			st.Error = ""
			st.ResourceCount = len(msg.resources)
			m.byAccount[msg.account.Name] = msg.resources
		}
		m.clampCursor()
		return m, nil

	case tickMsg:
		if m.paused {
			return m, tick(m.refresh)
		}
		var cmds []tea.Cmd
		cmds = append(cmds, tick(m.refresh))
		for _, a := range m.accounts {
			cmds = append(cmds, fetchAccountCmd(m.providers, a))
		}
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		m.cursor++
		m.clampCursor()
	case "left", "h":
		if m.sidebarIndex > 0 {
			m.sidebarIndex--
			m.cursor = 0
		}
	case "right", "l":
		if m.sidebarIndex < len(m.accounts) {
			m.sidebarIndex++
			m.cursor = 0
		}
	case "tab":
		m.tabIndex = (m.tabIndex + 1) % len(model.AllCategories)
		m.cursor = 0
	case "shift+tab":
		m.tabIndex--
		if m.tabIndex < 0 {
			m.tabIndex = len(model.AllCategories) - 1
		}
		m.cursor = 0
	case "1", "2", "3", "4":
		idx := int(msg.String()[0] - '1')
		if idx < len(model.AllCategories) {
			m.tabIndex = idx
			m.cursor = 0
		}
	case "p":
		m.paused = !m.paused
	case "r":
		var cmds []tea.Cmd
		for _, a := range m.accounts {
			cmds = append(cmds, fetchAccountCmd(m.providers, a))
		}
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

// selectedAccountName returns "" for the aggregated view, else the chosen
// account's name.
func (m Model) selectedAccountName() string {
	if m.sidebarIndex == 0 {
		return ""
	}
	names := m.sortedAccountNames()
	idx := m.sidebarIndex - 1
	if idx < 0 || idx >= len(names) {
		return ""
	}
	return names[idx]
}

// visibleResources returns resources for the current account selection and
// active category tab, sorted for stable display.
func (m Model) visibleResources() []model.Resource {
	category := model.AllCategories[m.tabIndex]
	selected := m.selectedAccountName()

	var out []model.Resource
	if selected == "" {
		for _, name := range m.sortedAccountNames() {
			for _, r := range m.byAccount[name] {
				if r.Category == category {
					out = append(out, r)
				}
			}
		}
	} else {
		for _, r := range m.byAccount[selected] {
			if r.Category == category {
				out = append(out, r)
			}
		}
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Account != out[j].Account {
			return out[i].Account < out[j].Account
		}
		if out[i].Region != out[j].Region {
			return out[i].Region < out[j].Region
		}
		return out[i].Name < out[j].Name
	})
	return out
}

func (m *Model) clampCursor() {
	rows := len(m.visibleResources())
	if m.cursor >= rows {
		m.cursor = rows - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}
