package dify

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

type ConversationsRequest struct {
	LastID string `json:"last_id,omitempty"`
	Limit  int    `json:"limit"`
	User   string `json:"user"`
}

type ConversationsResponse struct {
	Limit   int                         `json:"limit"`
	HasMore bool                        `json:"has_more"`
	Data    []ConversationsDataResponse `json:"data"`
}

type ConversationsDataResponse struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Inputs    map[string]string `json:"inputs"`
	Status    string            `json:"status"`
	CreatedAt int64             `json:"created_at"`
}

// ConversationsVariablesRequest defines the request for getting conversation variables
type ConversationsVariablesRequest struct {
	ConversationID string `json:"conversation_id,omitempty"`
	User           string `json:"user"`
	LastID         string `json:"last_id,omitempty"`
	Limit          int    `json:"limit,omitempty"`
	VariableName   string `json:"variable_name,omitempty"`
}

// ConversationsVariablesResponse defines the response for getting conversation variables
type ConversationsVariablesResponse struct {
	Limit   int                         `json:"limit"`
	HasMore bool                        `json:"has_more"`
	Data    []ConversationsVariableData `json:"data"`
}

// ConversationsVariableData defines the conversation variable data structure
type ConversationsVariableData struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ValueType   string `json:"value_type"`
	Value       string `json:"value"`
	Description string `json:"description"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

type ConversationsRenamingRequest struct {
	ConversationID string `json:"conversation_id,omitempty"`
	Name           string `json:"name"`
	User           string `json:"user"`
}

type ConversationsRenamingResponse struct {
	Result string `json:"result"`
}

/* Get conversation list
 * Gets the session list of the current user. By default, the last 20 sessions are returned.
 */
func (api *API) Conversations(ctx context.Context, req *ConversationsRequest) (resp *ConversationsResponse, err error) {
	if req.User == "" {
		err = errors.New("ConversationsRequest.User Illegal")
		return
	}
	if req.Limit == 0 {
		req.Limit = 20
	}

	httpReq, err := api.createBaseRequest(ctx, http.MethodGet, "/v1/conversations", nil)
	if err != nil {
		return
	}

	query := httpReq.URL.Query()
	query.Set("last_id", req.LastID)
	query.Set("user", req.User)
	query.Set("limit", strconv.FormatInt(int64(req.Limit), 10))
	httpReq.URL.RawQuery = query.Encode()

	err = api.c.sendJSONRequest(httpReq, &resp)
	return
}

/* Get conversation variables
 * Gets the variables of a specific conversation.
 */
func (api *API) ConversationsVariables(ctx context.Context, req *ConversationsVariablesRequest) (resp *ConversationsVariablesResponse, err error) {
	if req.ConversationID == "" {
		err = errors.New("ConversationsVariablesRequest.ConversationID is required")
		return
	}
	if req.User == "" {
		err = errors.New("ConversationsVariablesRequest.User is required")
		return
	}
	if req.Limit == 0 {
		req.Limit = 20
	}
	if req.Limit < 1 || req.Limit > 100 {
		err = errors.New("ConversationsVariablesRequest.Limit must be between 1 and 100")
		return
	}

	httpReq, err := api.createBaseRequest(ctx, http.MethodGet, fmt.Sprintf("/v1/conversations/%s/variables", req.ConversationID), nil)
	if err != nil {
		return
	}

	query := httpReq.URL.Query()
	query.Set("user", req.User)
	if req.LastID != "" {
		query.Set("last_id", req.LastID)
	}
	if req.Limit > 0 {
		query.Set("limit", strconv.FormatInt(int64(req.Limit), 10))
	}
	if req.VariableName != "" {
		query.Set("variable_name", req.VariableName)
	}
	httpReq.URL.RawQuery = query.Encode()

	err = api.c.sendJSONRequest(httpReq, &resp)
	return
}

/* Conversation renaming
 * Rename conversations; the name is displayed in multi-session client interfaces.
 */
func (api *API) ConversationsRenaming(ctx context.Context, req *ConversationsRenamingRequest) (resp *ConversationsRenamingResponse, err error) {
	url := fmt.Sprintf("/v1/conversations/%s/name", req.ConversationID)
	req.ConversationID = ""

	httpReq, err := api.createBaseRequest(ctx, http.MethodPost, url, req)
	if err != nil {
		return
	}
	err = api.c.sendJSONRequest(httpReq, &resp)
	return
}
