package context

import (
	"github.com/jesseduffield/generics/slices"
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazygit/pkg/gui/keybindings"
	"github.com/jesseduffield/lazygit/pkg/gui/presentation"
	"github.com/jesseduffield/lazygit/pkg/gui/style"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
)

type MenuContext struct {
	*MenuViewModel
	*ListContextTrait
}

var _ types.IListContext = (*MenuContext)(nil)

func NewMenuContext(
	view *gocui.View,

	onFocus func(...types.OnFocusOpts) error,
	onRenderToMain func(...types.OnFocusOpts) error,
	onFocusLost func() error,

	c *types.HelperCommon,
	getOptionsMap func() map[string]string,
) *MenuContext {
	viewModel := NewMenuViewModel()

	return &MenuContext{
		MenuViewModel: viewModel,
		ListContextTrait: &ListContextTrait{
			Context: NewSimpleContext(NewBaseContext(NewBaseContextOpts{
				ViewName:        "menu",
				Key:             "menu",
				Kind:            types.PERSISTENT_POPUP,
				OnGetOptionsMap: getOptionsMap,
				Focusable:       true,
			}), ContextCallbackOpts{
				OnFocus:        onFocus,
				OnFocusLost:    onFocusLost,
				OnRenderToMain: onRenderToMain,
			}),
			getDisplayStrings: viewModel.GetDisplayStrings,
			list:              viewModel,
			viewTrait:         NewViewTrait(view),
			c:                 c,
		},
	}
}

// TODO: remove this thing.
func (self *MenuContext) GetSelectedItemId() string {
	item := self.GetSelected()
	if item == nil {
		return ""
	}

	return item.DisplayString
}

type MenuViewModel struct {
	menuItems []*types.MenuItem
	*BasicViewModel[*types.MenuItem]
}

func NewMenuViewModel() *MenuViewModel {
	self := &MenuViewModel{
		menuItems: nil,
	}

	self.BasicViewModel = NewBasicViewModel(func() []*types.MenuItem { return self.menuItems })

	return self
}

func (self *MenuViewModel) SetMenuItems(items []*types.MenuItem) {
	self.menuItems = items
}

// TODO: move into presentation package
func (self *MenuViewModel) GetDisplayStrings(_startIdx int, _length int) [][]string {
	showKeys := slices.Some(self.menuItems, func(item *types.MenuItem) bool {
		return item.Key != nil
	})

	return slices.Map(self.menuItems, func(item *types.MenuItem) []string {
		displayStrings := getItemDisplayStrings(item)
		if showKeys {
			displayStrings = slices.Prepend(displayStrings, style.FgCyan.Sprint(keybindings.GetKeyDisplay(item.Key)))
		}
		return displayStrings
	})
}

func getItemDisplayStrings(item *types.MenuItem) []string {
	if item.DisplayStrings != nil {
		return item.DisplayStrings
	}

	styledStr := item.DisplayString
	if item.OpensMenu {
		styledStr = presentation.OpensMenuStyle(styledStr)
	}
	return []string{styledStr}
}

func (self *MenuContext) GetKeybindings(opts types.KeybindingsOpts) []*types.Binding {
	basicBindings := self.ListContextTrait.GetKeybindings(opts)
	menuItemsWithKeys := slices.Filter(self.menuItems, func(item *types.MenuItem) bool {
		return item.Key != nil
	})

	menuItemBindings := slices.Map(menuItemsWithKeys, func(item *types.MenuItem) *types.Binding {
		return &types.Binding{
			Key:     item.Key,
			Handler: func() error { return self.OnMenuPress(item) },
		}
	})

	// appending because that means the menu item bindings have lower precedence.
	// So if a basic binding is to escape from the menu, we want that to still be
	// what happens when you press escape. This matters when we're showing the menu
	// for all keybindings of say the files context.
	return append(basicBindings, menuItemBindings...)
}

func (self *MenuContext) OnMenuPress(selectedItem *types.MenuItem) error {
	if err := self.c.PopContext(); err != nil {
		return err
	}

	if err := selectedItem.OnPress(); err != nil {
		return err
	}

	return nil
}
