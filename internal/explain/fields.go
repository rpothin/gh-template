package explain

// Section identifies which top-level manifest section a field belongs to.
type Section string

const (
	SectionSettings     Section = "settings"
	SectionEnvironments Section = "environments"
	SectionTopics       Section = "topics"
)

// FieldDef describes a single manifest field for display purposes.
type FieldDef struct {
	Name    string
	Section Section
	Type    string
	Short   string // one-line description for the table view
	Long    string // multi-line description for the detail view
	DocsURL string
	Example string // YAML snippet
}

// Fields is the ordered registry of all manifest fields.
var Fields = []FieldDef{
	// ── Settings ─────────────────────────────────────────────────────────
	{
		Name:    "has_wiki",
		Section: SectionSettings,
		Type:    "bool",
		Short:   "Enable the repository Wiki tab",
		Long: "When true, the Wiki tab is visible on the repository page and contributors\n" +
			"can create and edit wiki pages for collaborative documentation.",
		DocsURL: "https://docs.github.com/en/communities/documenting-your-project-with-wikis",
		Example: "settings:\n  has_wiki: true",
	},
	{
		Name:    "has_issues",
		Section: SectionSettings,
		Type:    "bool",
		Short:   "Enable the Issues tab for bug/feature tracking",
		Long: "When true, the Issues tab is visible and users can open, comment on,\n" +
			"and close issues to track bugs, feature requests, and tasks.",
		DocsURL: "https://docs.github.com/en/issues",
		Example: "settings:\n  has_issues: true",
	},
	{
		Name:    "has_projects",
		Section: SectionSettings,
		Type:    "bool",
		Short:   "Enable the Projects tab (kanban/task boards)",
		Long: "When true, the Projects tab is available, allowing project boards to be\n" +
			"created directly on the repository for task management.",
		DocsURL: "https://docs.github.com/en/issues/planning-and-tracking-with-projects",
		Example: "settings:\n  has_projects: false",
	},
	{
		Name:    "allow_squash_merge",
		Section: SectionSettings,
		Type:    "bool",
		Short:   "Allow squash-merging pull requests",
		Long: "When true, contributors can merge a PR by squashing all commits into one,\n" +
			"producing a clean, linear history. At least one merge strategy must remain\n" +
			"enabled at all times.",
		DocsURL: "https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/configuring-pull-request-merges/about-merge-methods-on-github",
		Example: "settings:\n  allow_squash_merge: true",
	},
	{
		Name:    "allow_merge_commit",
		Section: SectionSettings,
		Type:    "bool",
		Short:   "Allow standard merge commits on pull requests",
		Long: "When true, contributors can merge a PR using a standard merge commit,\n" +
			"preserving all individual commits and adding an explicit merge commit\n" +
			"to create a non-linear history.",
		DocsURL: "https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/configuring-pull-request-merges/about-merge-methods-on-github",
		Example: "settings:\n  allow_merge_commit: false",
	},
	{
		Name:    "allow_rebase_merge",
		Section: SectionSettings,
		Type:    "bool",
		Short:   "Allow rebase-merging pull requests",
		Long: "When true, contributors can merge a PR by replaying each commit individually\n" +
			"onto the base branch. This produces a linear history without a merge commit.",
		DocsURL: "https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/configuring-pull-request-merges/about-merge-methods-on-github",
		Example: "settings:\n  allow_rebase_merge: true",
	},
	{
		Name:    "allow_auto_merge",
		Section: SectionSettings,
		Type:    "bool",
		Short:   "Let PRs auto-merge once required checks pass",
		Long: "When true, contributors can enable auto-merge on a pull request. The PR\n" +
			"merges automatically as soon as all required status checks and required\n" +
			"reviews are satisfied — no manual merge click needed.",
		DocsURL: "https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/incorporating-changes-from-a-pull-request/automatically-merging-a-pull-request",
		Example: "settings:\n  allow_auto_merge: true",
	},
	{
		Name:    "delete_branch_on_merge",
		Section: SectionSettings,
		Type:    "bool",
		Short:   "Auto-delete source branch after PR merge",
		Long: "When true, GitHub automatically deletes the head branch of a pull request\n" +
			"after it is merged. This keeps the repository tidy and prevents stale\n" +
			"branch accumulation over time.",
		DocsURL: "https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/configuring-pull-request-merges/managing-the-automatic-deletion-of-branches",
		Example: "settings:\n  delete_branch_on_merge: true",
	},
	{
		Name:    "allow_update_branch",
		Section: SectionSettings,
		Type:    "bool",
		Short:   `Show "Update branch" button on PRs behind base`,
		Long: "When true, GitHub surfaces an \"Update branch\" button on open pull requests\n" +
			"whose base branch has advanced. Contributors can click it to sync the latest\n" +
			"base changes into their PR branch directly from the browser, without needing\n" +
			"to run git commands locally.",
		DocsURL: "https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/configuring-pull-request-merges/managing-suggestions-to-update-pull-request-branches",
		Example: "settings:\n  allow_update_branch: true",
	},
	{
		Name:    "visibility",
		Section: SectionSettings,
		Type:    "string",
		Short:   `Repository visibility: "public" or "private"`,
		Long: "Controls who can see the repository.\n" +
			"  public  — visible to everyone on the internet\n" +
			"  private — visible only to you and explicitly added collaborators",
		DocsURL: "https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/managing-repository-settings/setting-repository-visibility",
		Example: "settings:\n  visibility: private",
	},
	{
		Name:    "description",
		Section: SectionSettings,
		Type:    "string",
		Short:   "Short description shown on the repository page",
		Long: "A plain-text string (up to 350 characters) displayed below the repository\n" +
			"name on GitHub and in search results.",
		DocsURL: "https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository",
		Example: "settings:\n  description: \"A lightweight GitHub CLI extension for repo governance.\"",
	},
	// ── Environments ──────────────────────────────────────────────────────
	{
		Name:    "name",
		Section: SectionEnvironments,
		Type:    "string",
		Short:   `Environment name (e.g. "production", "staging")`,
		Long: "The unique identifier for a deployment environment within the repository.\n" +
			"Environment names are referenced in GitHub Actions workflows under the\n" +
			"`environment:` key of a job to apply protection rules.",
		DocsURL: "https://docs.github.com/en/actions/deployment/targeting-different-environments/using-environments-for-deployment",
		Example: "environments:\n  - name: production\n    wait_timer: 0",
	},
	{
		Name:    "wait_timer",
		Section: SectionEnvironments,
		Type:    "int",
		Short:   "Minutes to wait before a deployment can proceed (0–43200)",
		Long: "In the GitHub UI: Settings → Environments → [environment name]\n" +
			"  → Environment protection rules → Wait timer.\n\n" +
			"Adds a mandatory delay (in minutes) before any Actions job targeting this\n" +
			"environment is allowed to run. Set to 0 to disable the wait.\n" +
			"The maximum value is 43200 (equivalent to 30 days).",
		DocsURL: "https://docs.github.com/en/actions/deployment/targeting-different-environments/using-environments-for-deployment#wait-timer",
		Example: "environments:\n  - name: production\n    wait_timer: 5",
	},
}

// FieldsByName returns a map of field name → FieldDef for O(1) lookup.
func FieldsByName() map[string]FieldDef {
	m := make(map[string]FieldDef, len(Fields))
	for _, f := range Fields {
		m[f.Name] = f
	}
	return m
}
