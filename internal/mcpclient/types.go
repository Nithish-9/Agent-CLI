package client

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type MCPClient struct {
	Client                *mcp.Client
	ServerInfo            map[string]*MCPServerInfo // key is server name
	MCPServerPlannerInput *MCPServerPlannerInput
}

type MCPServerInfo struct {
	Session       *mcp.ClientSession
	Metadata      *MCPServerMetadata
	Tools         map[string]*ToolSpecData // key is tool category
	ToolConfigMap map[string]*mcp.Tool     // key is a tool
	URL           string
	Transport     string
}

type ToolSpecData struct {
	ToolMap  map[string]*mcp.Tool // key is tool name
	ToolList []*PlainToolData
}

type MCPServerMetadata struct {
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	ToolCategories []*ToolCategory `json:"tool_categories"`
}

type ToolCategory struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Tools       []*MCPTool `json:"tools"`
}

type MCPTool struct {
	Tool    *mcp.Tool                                                                          `json:"tool"`
	Handler func(context.Context, *mcp.CallToolRequest, any) (*mcp.CallToolResult, any, error) `json:"-"`
}

type MCPServerPlannerInput struct {
	MCPServer []*MCPServerPlanner `json:"mcp_servers" jsonschema:"description=List of MCP servers available for the LLM to choose from"`
}

type MCPServerPlanner struct {
	Name            string              `json:"name"            jsonschema:"description=Unique name of the MCP server"`
	Description     string              `json:"description"     jsonschema:"description=What this MCP server does and what domain it covers"`
	ToolCategoryLst []*PlainToolCatData `json:"tool_categories" jsonschema:"description=List of tool categories available under this MCP server"`
}

type PlainToolCatData struct {
	Name        string `json:"name"        jsonschema:"description=Name of the tool category"`
	Description string `json:"description" jsonschema:"description=What kinds of tools this category contains"`
}

type MCPServerPlannerOutput struct {
	MCPServer []*MCPServerPlannerOutputItem `json:"mcp_servers" jsonschema:"description=List of relevant MCP servers selected by the LLM. Return empty array if no tools are needed"`
}

type MCPServerPlannerOutputItem struct {
	Name            string                    `json:"name"            jsonschema:"description=Exact server name from the catalogue — do not rename or modify"`
	ToolCategoryLst []*PlainToolCatOutputData `json:"tool_categories" jsonschema:"description=List of relevant tool category names from this server"`
}

type PlainToolCatOutputData struct {
	Name string `json:"name" jsonschema:"description=Exact tool category name from the catalogue — do not rename or modify"`
}

type MCPToolPlannerInput struct {
	MCPToolPlanner []*MCPToolPlanner `json:"mcp_tools" jsonschema:"description=List of MCP servers with their tools available for selection"`
}

type MCPToolPlanner struct {
	Name        string           `json:"name"        jsonschema:"description=Name of the MCP server"`
	Description string           `json:"description" jsonschema:"description=What this MCP server does and what domain it covers"`
	ToolLst     []*PlainToolData `json:"tool_list"   jsonschema:"description=List of available tools under this server"`
}

type PlainToolData struct {
	Name        string `json:"name"        jsonschema:"description=Exact tool name as registered in the MCP server"`
	Description string `json:"description" jsonschema:"description=What this tool does and when it should be invoked"`
}

type ToolPlannerOutput struct {
	MCPToolPlanner []*MCPToolPlannerOutputItem `json:"mcp_tools" jsonschema:"description=Final list of servers and tools selected to fulfil the user request. Return empty array if no tools are needed"`
}

type MCPToolPlannerOutputItem struct {
	Name    string                 `json:"name"      jsonschema:"description=Exact server name from the catalogue — do not rename or modify"`
	ToolLst []*PlainToolOutputData `json:"tool_list" jsonschema:"description=List of exact tools selected from this server to fulfil the request"`
}

type PlainToolOutputData struct {
	Name string `json:"name" jsonschema:"description=Exact tool name from the catalogue — do not rename or modify"`
}
