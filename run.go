package console

import (
	"fmt"

	"github.com/kballard/go-shellquote"
	"github.com/spf13/cobra"
)

// Run - Start the console application (readline loop). Blocking.
// The error returned will always be an error that the console
// application does not understand or cannot handle.
func (c *Console) Run() (err error) {
	// Also, if the user specified custom histories to the
	// current menu, they are not bound to the shell yet.
	c.loadActiveHistories()

	// Print the console logo
	if c.printLogo != nil {
		c.printLogo(c)
	}

	for {
		// Current menu setup
		menu := c.activeMenu() // We work with the active menu.
		menu.resetCommands()   // Regenerate the commands for the menu.
		menu.resetCmdOutput()  // Reset or adjust any buffered command output.

		// Console-wide setup.
		c.reloadConfig()         // Rebind the prompt helpers, and similar stuff.
		c.runPreReadHooks()      // Run user-provided pre-loop hooks
		c.ensureNoRootRunner()   // Avoid printing any help when the command line is empty
		c.hideFilteredCommands() // Hide commands that are not available

		// Block and read user input. Provides completion, syntax, hints, etc.
		// Various types of errors might arise from here. We handle them in a
		// special function, where we can specify behavior for certain errors.
		line, err := c.shell.Readline()
		if err != nil {
			menu.handleInterrupt(err)

			continue
		}

		// Split the line into shell words.
		args, err := shellquote.Split(line)
		if err != nil {
			fmt.Printf("Line error: %s\n", err.Error())

			continue
		}

		if len(args) == 0 {
			continue
		}

		// Run user-provided pre-run line hooks,
		// which may modify the input line args.
		args = c.runLineHooks(args)

		// Run all hooks and the command itself
		c.execute(args)
	}
}

// Generally, an empty command entered should just print a new prompt,
// unlike for classic CLI usage when the program will print its usage string.
// We simply remove any RunE from the root command, so that nothing is
// printed/executed by default. Pre/Post runs are still used if any.
func (c *Console) ensureNoRootRunner() {
	if c.activeMenu().Command != nil {
		c.activeMenu().RunE = func(cmd *cobra.Command, args []string) error {
			return nil
		}
	}
}

func (c *Console) loadActiveHistories() {
	c.shell.History.Delete()

	for _, name := range c.activeMenu().historyNames {
		c.shell.History.Add(name, c.activeMenu().histories[name])
	}
}

func (c *Console) runPreReadHooks() {
	for _, hook := range c.PreReadlineHooks {
		hook()
	}
}

func (c *Console) runLineHooks(args []string) []string {
	processed := args

	// Or modify them again
	for _, hook := range c.PreCmdRunLineHooks {
		processed, _ = hook(processed)
	}

	return processed
}

func (c *Console) runPreRunHooks() {
	for _, hook := range c.PreCmdRunHooks {
		hook()
	}
}

func (c *Console) runPostRunHooks() {
	for _, hook := range c.PostCmdRunHooks {
		hook()
	}
}

// execute - The user has entered a command input line, the arguments
// have been processed: we synchronize a few elements of the console,
// then pass these arguments to the command parser for execution and error handling.
func (c *Console) execute(args []string) {
	menu := c.activeMenu()

	// Find the target command: if this command is filtered, don't run it,
	// nor any pre-run hooks. We don't care about any error here: we just
	// want to know if the command is hidden.
	target, _, _ := menu.Find(args)
	if c.isFiltered(target) {
		return
	}

	c.runPreRunHooks()

	// Asynchronous messages do not mess with the prompt from now on,
	// until end of execution. Once we are done executing the command,
	// they can again.
	c.mutex.RLock()
	c.isExecuting = true
	c.mutex.RUnlock()

	defer func() {
		c.mutex.RLock()
		c.isExecuting = false
		c.mutex.RUnlock()
	}()

	// Assign those arguments to our parser
	menu.SetArgs(args)

	if c.LeaveNewline {
		fmt.Println()
	}

	// Execute the command line, with the current menu' parser.
	// Process the errors raised by the parser.
	// A few of them are not really errors, and trigger some stuff.
	menu.Execute()

	c.runPostRunHooks()
}
