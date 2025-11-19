package dify

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// ErrorResponse 定义服务器返回的错误响应格式
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

type ChatMessageStreamResponse struct {
	Event          string `json:"event"`
	TaskID         string `json:"task_id"`
	ID             string `json:"id"`
	Answer         string `json:"answer"`
	CreatedAt      int64  `json:"created_at"`
	ConversationID string `json:"conversation_id"`
	MessageID      string `json:"message_id"`
	WorkflowRunID  string `json:"workflow_run_id,omitempty"`
	Data           struct {
		ID                        string                     `json:"id,omitempty"`
		NodeID                    string                     `json:"node_id,omitempty"`
		NodeType                  string                     `json:"node_type,omitempty"`
		Title                     string                     `json:"title,omitempty"`
		Label                     string                     `json:"label,omitempty"`
		Index                     int                        `json:"index,omitempty"`
		PredecessorNodeID         string                     `json:"predecessor_node_id,omitempty"`
		Inputs                    map[string]interface{}     `json:"inputs,omitempty"`
		ProcessData               interface{}                `json:"process_data,omitempty"`
		Outputs                   map[string]interface{}     `json:"outputs,omitempty"`
		Status                    string                     `json:"status,omitempty"`
		Error                     interface{}                `json:"error,omitempty"`
		ElapsedTime               float64                    `json:"elapsed_time,omitempty"`
		ExecutionMetadata         map[string]interface{}     `json:"execution_metadata,omitempty"`
		CreatedAt                 int64                      `json:"created_at,omitempty"`
		FinishedAt                int64                      `json:"finished_at,omitempty"`
		Files                     []interface{}              `json:"files,omitempty"`
		ParallelID                string                     `json:"parallel_id,omitempty"`
		ParallelStartNodeID       string                     `json:"parallel_start_node_id,omitempty"`
		ParentParallelID          string                     `json:"parent_parallel_id,omitempty"`
		ParentParallelStartNodeID string                     `json:"parent_parallel_start_node_id,omitempty"`
		IterationID               string                     `json:"iteration_id,omitempty"`
		LoopID                    string                     `json:"loop_id,omitempty"`
		Data                      *ChatMessageStreamResponse `json:"data,omitempty"`
	} `json:"data,omitempty"`
}

type ChatMessageStreamChannelResponse struct {
	ChatMessageStreamResponse
	Err error `json:"-"`
}

func (api *API) ChatMessagesStreamRaw(ctx context.Context, req *ChatMessageRequest) (*http.Response, error) {
	req.ResponseMode = "streaming"

	httpReq, err := api.createBaseRequest(ctx, http.MethodPost, "/v1/chat-messages", req)
	if err != nil {
		return nil, err
	}
	return api.c.sendRequest(httpReq)
}

func (api *API) ChatMessagesStream(ctx context.Context, req *ChatMessageRequest) (chan ChatMessageStreamChannelResponse, error) {
	httpResp, err := api.ChatMessagesStreamRaw(ctx, req)
	if err != nil {
		return nil, err
	}

	streamChannel := make(chan ChatMessageStreamChannelResponse)
	go api.chatMessagesStreamHandle(ctx, httpResp, streamChannel)
	return streamChannel, nil
}

func (api *API) chatMessagesStreamHandle(ctx context.Context, resp *http.Response, streamChannel chan ChatMessageStreamChannelResponse) {
	defer resp.Body.Close()
	defer close(streamChannel)

	reader := bufio.NewReader(resp.Body)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			line, err := reader.ReadBytes('\n')
			if err != nil {
				streamChannel <- ChatMessageStreamChannelResponse{
					Err: fmt.Errorf("error reading line: %w", err),
				}
				return
			}

			if !bytes.HasPrefix(line, []byte("data:")) {
				if bytes.HasPrefix(line, []byte("{")) {
					// 检查是否是错误响应格式
					var errorResp ErrorResponse
					if err := json.Unmarshal(line, &errorResp); err == nil && errorResp.Code != "" {
						streamChannel <- ChatMessageStreamChannelResponse{
							Err: fmt.Errorf("server error: %s (code: %s, status: %d)", errorResp.Message, errorResp.Code, errorResp.Status),
						}
						return
					}
				}
				continue
			}
			line = bytes.TrimPrefix(line, []byte("data:"))

			var streamResp ChatMessageStreamChannelResponse
			if err = json.Unmarshal(line, &streamResp); err != nil {
				streamChannel <- ChatMessageStreamChannelResponse{
					Err: fmt.Errorf("error unmarshalling event: %w", err),
				}
				return
			} else if streamResp.Event == "error" {
				streamChannel <- ChatMessageStreamChannelResponse{
					Err: errors.New("error streaming event: " + string(line)),
				}
				return
			} else if streamResp.Event == "message" && streamResp.Answer == "" {
				continue
			}
			streamChannel <- streamResp
		}
	}
}
