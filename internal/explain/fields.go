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
		Name:    "has_discussions",
		Section: SectionSettings,
		Type:    "bool",
		Short:   "Enable Discussions for community conversations",
		Long: "In the GitHub UI: Settings → General → Features → Discussions.\n\n" +
			"When true, the Discussions tab is shown on the repository page. Discussions\n" +
			"provides a community forum for questions, ideas, and announcements — separate\n" +
			"from issues (which track specific bugs or tasks).",
		DocsURL: "https://docs.github.com/en/discussions",
		Example: "settings:\n  has_discussions: true",
	},
	{
		Name:    "has_pull_requests",
		Section: SectionSettings,
		Type:    "bool",
		Short:   "Enable the Pull Requests tab",
		Long: "In the GitHub UI: Settings → General → Features → Pull Requests.\n\n" +
			"When false, the Pull Requests tab is hidden and no new pull requests can be\n" +
			"opened on the repository. Useful for repositories that accept contributions\n" +
			"exclusively through issues or are purely archival.",
		DocsURL: "https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/managing-repository-settings/managing-pull-request-reviews-in-your-repository",
		Example: "settings:\n  has_pull_requests: true",
	},
	{
		Name:    "pull_request_creation_policy",
		Section: SectionSettings,
		Type:    "string",
		Short:   `Who can open pull requests: "all" | "collaborators_only"`,
		Long: "In the GitHub UI: Settings → General → Pull Requests → Pull request creation.\n\n" +
			"Controls who is allowed to open pull requests against this repository:\n" +
			"  all                — any GitHub user can open a pull request (default for public repos)\n" +
			"  collaborators_only — only repository collaborators can open pull requests\n\n" +
			"Setting this to \"collaborators_only\" reduces noise from external contributors\n" +
			"while still allowing forks and issues.",
		DocsURL: "https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/managing-repository-settings",
		Example: "settings:\n  pull_request_creation_policy: collaborators_only",
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
	{
		Name:    "prevent_self_review",
		Section: SectionEnvironments,
		Type:    "bool",
		Short:   "Prevent deployer from approving their own deployment",
		Long: "In the GitHub UI: Settings → Environments → [environment name]\n" +
			"  → Environment protection rules → Prevent self-review.\n\n" +
			"When true, the user who triggered a deployment cannot be one of the required\n" +
			"reviewers who approve it. Requires at least one reviewer to be configured.",
		DocsURL: "https://docs.github.com/en/actions/deployment/targeting-different-environments/using-environments-for-deployment#required-reviewers",
		Example: "environments:\n  - name: production\n    prevent_self_review: true",
	},
	{
		Name:    "reviewers",
		Section: SectionEnvironments,
		Type:    "[]string",
		Short:   "Users/teams that must approve a deployment (up to 6)",
		Long: "In the GitHub UI: Settings → Environments → [environment name]\n" +
			"  → Environment protection rules → Required reviewers.\n\n" +
			"A list of GitHub usernames (e.g. \"rpothin\") or team references in\n" +
			"\"org/team-slug\" format (e.g. \"my-org/platform-team\") who must approve\n" +
			"before any Actions job targeting this environment can proceed.\n" +
			"GitHub allows up to 6 reviewers per environment.\n\n" +
			"At apply time (create/sync), usernames are resolved to numeric IDs.",
		DocsURL: "https://docs.github.com/en/actions/deployment/targeting-different-environments/using-environments-for-deployment#required-reviewers",
		Example: "environments:\n  - name: production\n    reviewers:\n      - rpothin\n      - my-org/platform-team",
	},
	{
		Name:    "deployment_branch_policy",
		Section: SectionEnvironments,
		Type:    "string",
		Short:   `Branch/tag restriction mode: "all" | "protected" | "custom"`,
		Long: "In the GitHub UI: Settings → Environments → [environment name]\n" +
			"  → Deployment branches and tags.\n\n" +
			"Controls which branches and tags are allowed to deploy to this environment:\n" +
			"  all       — any branch or tag (default; omit this field for the same effect)\n" +
			"  protected — only branches with protection rules applied\n" +
			"  custom    — only branches/tags matching patterns in deployment_branch_patterns",
		DocsURL: "https://docs.github.com/en/actions/deployment/targeting-different-environments/using-environments-for-deployment#deployment-branches-and-tags",
		Example: "environments:\n  - name: production\n    deployment_branch_policy: custom",
	},
	{
		Name:    "deployment_branch_patterns",
		Section: SectionEnvironments,
		Type:    "[]string",
		Short:   "Branch/tag name patterns allowed to deploy (when policy is \"custom\")",
		Long: "Only used when deployment_branch_policy is \"custom\".\n\n" +
			"A list of branch or tag name patterns (glob-style) that are allowed\n" +
			"to deploy to this environment. Patterns support wildcards:\n" +
			"  main        — exact branch name\n" +
			"  release/*   — any branch starting with \"release/\"\n" +
			"  v[0-9]*     — tags matching a version pattern\n\n" +
			"At sync time, extra live patterns not in this list are removed.",
		DocsURL: "https://docs.github.com/en/actions/deployment/targeting-different-environments/using-environments-for-deployment#deployment-branches-and-tags",
		Example: "environments:\n  - name: production\n    deployment_branch_policy: custom\n    deployment_branch_patterns:\n      - main\n      - \"release/*\"",
	},
	{
		Name:    "variables",
		Section: SectionEnvironments,
		Type:    "[]object",
		Short:   "Environment variables available to Actions jobs for this environment",
		Long: "In the GitHub UI: Settings → Environments → [environment name]\n" +
			"  → Environment variables.\n\n" +
			"A list of name/value pairs that are injected as environment variables\n" +
			"into any Actions workflow job that targets this environment. Values are\n" +
			"plaintext (not encrypted). Use secrets for sensitive data instead.\n\n" +
			"At sync time, existing variables are updated; new ones are created.\n" +
			"Variables not in the manifest are left untouched.",
		DocsURL: "https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/store-information-in-variables#creating-configuration-variables-for-an-environment",
		Example: "environments:\n  - name: production\n    variables:\n      - name: DEPLOY_ENV\n        value: production",
	},
	{
		Name:    "secrets",
		Section: SectionEnvironments,
		Type:    "[]object",
		Short:   "Named secrets to initialize in this environment (placeholder value)",
		Long: "In the GitHub UI: Settings → Environments → [environment name]\n" +
			"  → Environment secrets.\n\n" +
			"A list of secret names to ensure exist in the environment. Because GitHub\n" +
			"never returns secret values via the API, the manifest always stores\n" +
			"value: \"PLACEHOLDER\" — this is not the real secret value.\n\n" +
			"Behaviour by command:\n" +
			"  snapshot — captures names only; value is always \"PLACEHOLDER\"\n" +
			"  create   — initializes each missing secret with the literal string\n" +
			"             \"PLACEHOLDER\" so the secret exists; prints a warning per secret\n" +
			"  sync     — same as create for missing secrets; existing secrets are untouched\n" +
			"  audit    — checks that each secret name is present; cannot verify values\n\n" +
			"⚠ Always update secret values manually after repository creation.",
		DocsURL: "https://docs.github.com/en/actions/deployment/targeting-different-environments/using-environments-for-deployment#environment-secrets",
		Example: "environments:\n  - name: production\n    secrets:\n      - name: API_KEY\n        value: \"PLACEHOLDER\"  # replace after creation",
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
