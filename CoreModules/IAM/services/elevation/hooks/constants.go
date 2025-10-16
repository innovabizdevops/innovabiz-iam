// Package hooks implementa a integração entre o serviço de elevação
// e os diferentes hooks MCP (Docker, Desktop Commander, GitHub, Figma)
// da plataforma INNOVABIZ IAM.
package hooks

// MCPHookType representa o tipo de hook MCP
type MCPHookType string

// Tipos de hook MCP suportados
const (
	Docker           MCPHookType = "docker"
	DesktopCommander MCPHookType = "desktop-commander"
	GitHub           MCPHookType = "github"
	Figma            MCPHookType = "figma"
)

// Constantes de escopo para Docker
const (
	ScopeDockerAdmin   = "docker:admin"
	ScopeDockerExec    = "docker:exec"
	ScopeDockerBuild   = "docker:build"
	ScopeDockerPull    = "docker:pull"
	ScopeDockerPush    = "docker:push"
	ScopeDockerRun     = "docker:run"
	ScopeDockerNetwork = "docker:network"
	ScopeDockerVolume  = "docker:volume"
)

// Constantes de escopo para GitHub
const (
	ScopeGitHubAdmin    = "github:admin"
	ScopeGitHubRead     = "github:read"
	ScopeGitHubWrite    = "github:write"
	ScopeGitHubDelete   = "github:delete"
	ScopeGitHubPR       = "github:pr"
	ScopeGitHubIssues   = "github:issues"
	ScopeGitHubSecurity = "github:security"
)

// Constantes de escopo para Desktop Commander
const (
	ScopeDesktopAdmin   = "desktop:admin"
	ScopeDesktopFS      = "desktop:fs"
	ScopeDesktopCmd     = "desktop:cmd"
	ScopeDesktopProcess = "desktop:process"
	ScopeDesktopConfig  = "desktop:config"
	ScopeDesktopSearch  = "desktop:search"
	ScopeDesktopEdit    = "desktop:edit"
)

// Constantes de escopo para Figma
const (
	ScopeFigmaAdmin   = "figma:admin"
	ScopeFigmaView    = "figma:view"
	ScopeFigmaEdit    = "figma:edit"
	ScopeFigmaComment = "figma:comment"
	ScopeFigmaExport  = "figma:export"
	ScopeFigmaLibrary = "figma:library"
	ScopeFigmaTeam    = "figma:team"
)