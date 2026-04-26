package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/mlarkin00/agentic-minions-tools/internal/mcpclient"
)

func main() {
	log.SetOutput(os.Stderr)

	gatewayURL := os.Getenv("GATEWAY_URL")
	if gatewayURL == "" {
		log.Fatal("GATEWAY_URL environment variable is required")
	}
	// The audience for Cloud Run IAM auth is the service URL itself.
	httpClient, err := mcpclient.NewAuthClient(gatewayURL)
	if err != nil {
		log.Fatalf("Failed to create auth client: %v", err)
	}

	client := mcpclient.NewClient(gatewayURL, httpClient)

	s := server.NewMCPServer(
		"agentic-minions",
		"0.1.0",
		server.WithToolCapabilities(true),
	)

	s.AddTool(
		mcp.NewTool("create_session",
			mcp.WithDescription("Create a new session with an agent. Returns a session ID for use in subsequent send_message calls."),
			mcp.WithString("role",
				mcp.Required(),
				mcp.Description("Agent role to create a session with. Valid roles: 'advising-on-code', 'designing-code', 'generating-code', 'validating-code', 'reviewing-code', 'maintaining-codebase-health', 'pm-mentor', 'pm-assistant', 'authoring-technical-content'."),
			),
			mcp.WithString("user_id",
				mcp.Required(),
				mcp.Description("User identifier for the session"),
			),
			mcp.WithString("agent_name",
				mcp.Description("Optional: The name of the specific agent (e.g. 'ClaudeAgent')"),
			),
		),
		makeCreateSessionHandler(client),
	)

	s.AddTool(
		mcp.NewTool("send_message",
			mcp.WithDescription("Send a message to an agent session. Streams the response via SSE and returns the final answer. Requires an existing session_id from create_session."),
			mcp.WithString("role",
				mcp.Required(),
				mcp.Description("Agent role (e.g. 'designing-code'). See create_session for the full list of valid roles."),
			),
			mcp.WithString("user_id",
				mcp.Required(),
				mcp.Description("User identifier"),
			),
			mcp.WithString("session_id",
				mcp.Required(),
				mcp.Description("Session ID from create_session"),
			),
			mcp.WithString("message",
				mcp.Required(),
				mcp.Description("The message to send to the agent"),
			),
			mcp.WithString("agent_name",
				mcp.Description("Optional: The name of the specific agent (e.g. 'ClaudeAgent')"),
			),
		),
		makeSendMessageHandler(client),
	)

	s.AddTool(
		mcp.NewTool("list_sessions",
			mcp.WithDescription("List active sessions for a user and agent role."),
			mcp.WithString("role",
				mcp.Required(),
				mcp.Description("Agent role (e.g. 'designing-code'). See create_session for the full list of valid roles."),
			),
			mcp.WithString("user_id",
				mcp.Required(),
				mcp.Description("User identifier"),
			),
		),
		makeListSessionsHandler(client),
	)

	s.AddTool(
		mcp.NewTool("delete_session",
			mcp.WithDescription("Delete a session."),
			mcp.WithString("role",
				mcp.Required(),
				mcp.Description("Agent role (e.g. 'designing-code'). See create_session for the full list of valid roles."),
			),
			mcp.WithString("user_id",
				mcp.Required(),
				mcp.Description("User identifier"),
			),
			mcp.WithString("session_id",
				mcp.Required(),
				mcp.Description("Session ID to delete"),
			),
		),
		makeDeleteSessionHandler(client),
	)

	log.Println("MCP server starting on stdio...")
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("MCP server error: %v", err)
	}
}

func makeCreateSessionHandler(client *mcpclient.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		role, err := request.RequireString("role")
		if err != nil {
			return mcp.NewToolResultError("role is required"), nil
		}
		userID, err := request.RequireString("user_id")
		if err != nil {
			return mcp.NewToolResultError("user_id is required"), nil
		}
		agentName := request.GetString("agent_name", "")

		session, err := client.CreateSession(role, userID, agentName)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to create session: %v", err)), nil
		}

		result, _ := json.Marshal(map[string]string{
			"session_id": session.ID,
			"user_id":    session.UserID,
			"agent_name": session.AppName,
		})
		return mcp.NewToolResultText(string(result)), nil
	}
}

func makeSendMessageHandler(client *mcpclient.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		role, err := request.RequireString("role")
		if err != nil {
			return mcp.NewToolResultError("role is required"), nil
		}
		userID, err := request.RequireString("user_id")
		if err != nil {
			return mcp.NewToolResultError("user_id is required"), nil
		}
		sessionID, err := request.RequireString("session_id")
		if err != nil {
			return mcp.NewToolResultError("session_id is required"), nil
		}
		message, err := request.RequireString("message")
		if err != nil {
			return mcp.NewToolResultError("message is required"), nil
		}
		agentName := request.GetString("agent_name", "")

		result, err := client.SendMessage(role, userID, sessionID, message, agentName)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("send message failed: %v", err)), nil
		}

		resp, _ := json.Marshal(result)
		return mcp.NewToolResultText(string(resp)), nil
	}
}

func makeListSessionsHandler(client *mcpclient.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		role, err := request.RequireString("role")
		if err != nil {
			return mcp.NewToolResultError("role is required"), nil
		}
		userID, err := request.RequireString("user_id")
		if err != nil {
			return mcp.NewToolResultError("user_id is required"), nil
		}

		sessions, err := client.ListSessions(role, userID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("list sessions failed: %v", err)), nil
		}

		result, _ := json.Marshal(map[string]any{"sessions": sessions})
		return mcp.NewToolResultText(string(result)), nil
	}
}

func makeDeleteSessionHandler(client *mcpclient.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		role, err := request.RequireString("role")
		if err != nil {
			return mcp.NewToolResultError("role is required"), nil
		}
		userID, err := request.RequireString("user_id")
		if err != nil {
			return mcp.NewToolResultError("user_id is required"), nil
		}
		sessionID, err := request.RequireString("session_id")
		if err != nil {
			return mcp.NewToolResultError("session_id is required"), nil
		}

		if err := client.DeleteSession(role, userID, sessionID); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("delete session failed: %v", err)), nil
		}

		return mcp.NewToolResultText(`{"status":"deleted"}`), nil
	}
}
