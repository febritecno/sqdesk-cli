package config

import "runtime"

// KeyMap defines keybindings for the editor
type KeyMap struct {
	Undo              []string `yaml:"undo" mapstructure:"undo"`
	Redo              []string `yaml:"redo" mapstructure:"redo"`
	Copy              []string `yaml:"copy" mapstructure:"copy"`
	Paste             []string `yaml:"paste" mapstructure:"paste"`
	Cut               []string `yaml:"cut" mapstructure:"cut"`
	Word              []string `yaml:"word" mapstructure:"word"`
	SelectAll         []string `yaml:"select_all" mapstructure:"select_all"`
	ExecuteSelection  []string `yaml:"execute_selection" mapstructure:"execute_selection"`
	AIPromptSelection []string `yaml:"ai_prompt_selection" mapstructure:"ai_prompt_selection"`
}

// DefaultKeyMap returns keybindings based on OS
func DefaultKeyMap() KeyMap {
	km := KeyMap{
		Undo:              []string{"ctrl+z"},
		Redo:              []string{"ctrl+y"},
		Copy:              []string{"alt+c"},
		Paste:             []string{"alt+v"},
		Cut:               []string{"alt+x"},
		Word:              []string{"ctrl+w", "alt+backspace"},
		SelectAll:         []string{"ctrl+a"},
		ExecuteSelection:  []string{"ctrl+shift+e"},
		AIPromptSelection: []string{"ctrl+shift+k"},
	}

	if runtime.GOOS != "darwin" {
		// On non-Mac, try to support standard Ctrl keys
		// Note: Ctrl+C might still be intercepted by global quit handler
		// unless we handle it specifically in Update
		km.Copy = append(km.Copy, "ctrl+c")
		km.Paste = append(km.Paste, "ctrl+v")
		km.Cut = append(km.Cut, "ctrl+x")
	}
	
	return km
}
