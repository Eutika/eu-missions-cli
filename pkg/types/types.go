package types

// Command represents a remote command to be executed
type Command struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Description string `json:"description"`
	Commands []string `json:"commands"`
}

// CommandResult represents the result of a command execution
type CommandResult struct {
	ID     string `json:"id"`
	Results []string `json:"results"`
}
