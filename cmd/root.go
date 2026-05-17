package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "template",
	Short: "Create configured repositories and keep them in sync with a template",
	Long: `GitHub template repositories copy files, but not settings, topics,
environments, or other repository configuration. gh-template bridges that gap
with a template-metadata.yml manifest.

Capture supported settings from an existing repository, then:
  - create repositories with configuration applied from day one
  - audit derived repositories for drift against the template manifest
  - sync template configuration updates into derived repositories

Works best when one template repository is the source of truth for a family of
related repositories.`,
}

// Execute runs the root cobra command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Force Cobra to create the built-in completion command so we can patch
	// its Long description with usage guidance specific to this extension.
	rootCmd.InitDefaultCompletionCmd()

	if completionCmd, _, err := rootCmd.Find([]string{"completion"}); err == nil {
		completionCmd.Long = `Generate shell autocompletion scripts for gh-template.

⚠  Important — gh extension dispatch limitation:
   Completions register for the binary name "template", not "gh template".
   Tab completion therefore works when you invoke the binary directly, not
   through "gh template <TAB>".  This is a constraint of how "gh" extensions
   are dispatched and cannot be changed inside the extension itself.

── Windows / PowerShell ────────────────────────────────────────────────────

  1. Find the extension binary (run once to locate it):

       Get-Command gh-template

  2. Create a short alias so you can invoke it directly:

       Set-Alias -Name template -Value gh-template

  3. Load completions into your current session:

       gh template completion powershell | Out-String | Invoke-Expression

  4. Try it — type "template" then press Tab:

       template <TAB>                # lists subcommands
       template create --<TAB>       # lists --manifest, --private
       template audit --r<TAB>       # completes --repo

  5. To make the alias and completions permanent, add both lines to your
     PowerShell profile (open it with: notepad $PROFILE):

       Set-Alias -Name template -Value gh-template
       gh template completion powershell | Out-String | Invoke-Expression

── macOS / Linux ───────────────────────────────────────────────────────────

  Bash  — add to ~/.bashrc:
    source <(gh template completion bash)

  Zsh   — add to ~/.zshrc:
    source <(gh template completion zsh)

  Fish  — add to ~/.config/fish/config.fish:
    gh template completion fish | source

Run "gh template completion [shell] --help" for shell-specific details.`
	}
}
