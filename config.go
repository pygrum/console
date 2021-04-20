package gonsole

import (
	"fmt"

	"github.com/maxlandon/readline"
)

// var (
//         highlightingItemsComps = map[string]string{
//                 "{command}":          "highlight the command words",
//                 "{command-argument}": "highlight the command arguments",
//                 "{option}":           "highlight the option name",
//                 "{option-argument}":  "highlight the option arguments",
//                 // We will dynamically add all <$-env> items as well.
//         }
// )

// ConsoleConfig - The console configuration (prompts, hints, modes, etc)
type ConsoleConfig struct {
	InputMode           readline.InputMode       `json:"input_mode"`
	Prompts             map[string]*PromptConfig `json:"prompts"`
	Hints               bool                     `json:"hints"`
	MaxTabCompleterRows int                      `json:"max_tab_completer_rows"`
	Highlighting        map[string]string        `json:"highlighting"`
}

// NewDefaultConfig - Users wishing to setup a special console configuration should
// use this function in order to ensure there are no nil maps anywhere, and with defaults.
func NewDefaultConfig() *ConsoleConfig {
	return &ConsoleConfig{
		InputMode:           readline.Vim,
		Prompts:             map[string]*PromptConfig{},
		Hints:               true,
		MaxTabCompleterRows: 50,
		Highlighting: map[string]string{
			"{command}":          readline.BOLD,
			"{command-argument}": readline.FOREWHITE,
			"{option}":           readline.BOLD,
			"{option-argument}":  readline.FOREWHITE,
		},
	}
}

// PromptConfig - Contains all the information needed for the PromptConfig of a given context.
type PromptConfig struct {
	Left            string `json:"left"`
	Right           string `json:"right"`
	Newline         bool   `json:"newline"`
	Multiline       bool   `json:"multiline"`
	MultilinePrompt string `json:"multiline_prompt"`
}

// newDefaultPromptConfig - Newly created contexts have a default prompt configuration
func newDefaultPromptConfig(context string) *PromptConfig {
	return &PromptConfig{
		Left:            fmt.Sprintf("gonsole (%s)", context),
		Right:           "",
		Newline:         true,
		Multiline:       true,
		MultilinePrompt: " > ",
	}
}

// loadDefaultConfig - Sane defaults for the gonsole Console.
func (c *Console) loadDefaultConfig() {
	c.config = NewDefaultConfig()
	// Make a default prompt for this application
	c.config.Prompts[""] = newDefaultPromptConfig("")
}

func (c *Console) reloadConfig() {

	// Setup the prompt, and input mode
	c.current.Prompt.loadFromConfig(c.config.Prompts[c.current.Name])
	c.Shell.MultilinePrompt = c.config.Prompts[c.current.Name].MultilinePrompt
	c.Shell.Multiline = c.config.Prompts[c.current.Name].Multiline
	c.Shell.InputMode = c.config.InputMode
	c.PreOutputNewline = c.config.Prompts[c.current.Name].Newline
}

// ExportConfig - The console exports its configuration in a JSON struct.
func (c *Console) ExportConfig() (conf *ConsoleConfig) {
	return c.config
}

// LoadConfig - Loads a config struct, but does immediately refresh the prompt.
// Settings will apply as they are needed by the console.
func (c *Console) LoadConfig(conf *ConsoleConfig) {
	if conf == nil {
		return
	}

	// Ensure no fields are nil
	if conf.Prompts == nil {
		p := &PromptConfig{
			Left:            "gonsole",
			Right:           "",
			Newline:         true,
			Multiline:       true,
			MultilinePrompt: " > ",
		}
		conf.Prompts = map[string]*PromptConfig{"": p}
	}

	// Users might forget to load default highlighting maps.
	if conf.Highlighting == nil {
		conf.Highlighting = map[string]string{
			"{command}":          readline.BOLD,
			"{command-argument}": readline.FOREWHITE,
			"{option}":           readline.BOLD,
			"{option-argument}":  readline.FOREWHITE,
		}
	}
	// Then load and apply all componenets that need a refresh now
	c.config = conf

	// Setup the prompt, and input mode
	c.reloadConfig()

	return
}
