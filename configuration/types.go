package configuration

type Config struct {
	MCPServers MCPServers `yaml:"mcp_servers"`
	Models     Models     `yaml:"models"`
	Planner    string     `yaml:"planner"`
	Executor   string     `yaml:"executor"`
}

type MCPServers struct {
	Servers []MCPServer `yaml:"servers"`
}

type MCPServer struct {
	Name      string `yaml:"name"`
	URL       string `yaml:"url"`
	Transport string `yaml:"transport"`
}

type Models struct {
	Models []Model `yaml:"models"`
}

type Model struct {
	Name     string `yaml:"name"`
	Provider string `yaml:"provider"`
	BaseURL  string `yaml:"base_url"`
	APIKey   string `yaml:"api_key"`
	Model    string `yaml:"model"`
}
