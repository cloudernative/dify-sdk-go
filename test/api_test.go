package test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"sync"

	"github.com/cloudernative/dify-sdk-go"
)

var (
	host         = "https://api.xxx.com"
	apiSecretKey = "app-xxx"
)

func TestApi3(t *testing.T) {
	// 初始化测试客户端
	client := dify.NewClientWithConfig(&dify.ClientConfig{
		Host:             host,
		DefaultAPISecret: apiSecretKey,
	})

	// 测试正常流
	t.Run("正常流式对话", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()

		// 发起流式请求
		ch, err := client.Api().ChatMessagesStream(ctx, &dify.ChatMessageRequest{
			Query: "IF([销售额] > 1000, CONVERT([订单日期], \"YYYY-MM\"), LEFT([产品编号], 3))",
			User:  "system",
		})
		if err != nil {
			t.Fatalf("请求失败: %v", err)
		}

		var (
			strBuilder strings.Builder
			received   int
		)

		// 处理流式响应
		for r := range ch {
			if r.Err != nil {
				t.Errorf(r.Err.Error())
			}
			if r.Event != "message" {
				t.Logf("忽略非消息事件: %s", r.Event)
				continue
			}
			// 验证必要字段
			if r.ConversationID == "" {
				t.Error("ConversationID 不能为空")
			}

			strBuilder.WriteString(r.Answer)
			received++
			fmt.Print(r.Answer)
		}

		// 结果断言
		if received == 0 {
			t.Error("未收到任何消息片段")
		}
		if strBuilder.Len() < 10 {
			t.Errorf("响应内容过短: %s", strBuilder.String())
		}
		t.Logf(strBuilder.String())
	})
}

func TestMessages(t *testing.T) {
	var cId = "ec373942-2d17-4f11-89bb-f9bbf863ebcc"
	var err error
	ctx := context.Background()

	// messages
	var messageReq = &dify.MessagesRequest{
		ConversationID: cId,
		User:           "jiuquan AI",
	}

	var client = dify.NewClient(host, apiSecretKey)

	var msg *dify.MessagesResponse
	if msg, err = client.Api().Messages(ctx, messageReq); err != nil {
		t.Fatal(err.Error())
		return
	}
	j, _ := json.Marshal(msg)
	t.Log(string(j))
}

func TestMessagesFeedbacks(t *testing.T) {
	var client = dify.NewClient(host, apiSecretKey)
	var err error
	ctx := context.Background()

	var id = "72d3dc0f-a6d5-4b5e-8510-bec0611a6048"

	var res *dify.MessagesFeedbacksResponse
	if res, err = client.Api().MessagesFeedbacks(ctx, &dify.MessagesFeedbacksRequest{
		MessageID: id,
		Rating:    dify.FeedbackLike,
		User:      "jiuquan AI",
	}); err != nil {
		t.Fatal(err.Error())
	}

	j, _ := json.Marshal(res)

	log.Println(string(j))
}

func TestConversations(t *testing.T) {
	var client = dify.NewClient(host, apiSecretKey)
	var err error
	ctx := context.Background()

	var res *dify.ConversationsResponse
	if res, err = client.Api().Conversations(ctx, &dify.ConversationsRequest{
		User: "jiuquan AI",
	}); err != nil {
		t.Fatal(err.Error())
	}

	j, _ := json.Marshal(res)

	log.Println(string(j))
}

func TestConversationsRename(t *testing.T) {
	var client = dify.NewClient(host, apiSecretKey)
	var err error
	ctx := context.Background()

	var res *dify.ConversationsRenamingResponse
	if res, err = client.Api().ConversationsRenaming(ctx, &dify.ConversationsRenamingRequest{
		ConversationID: "ec373942-2d17-4f11-89bb-f9bbf863ebcc",
		Name:           "rename!!!",
		User:           "jiuquan AI",
	}); err != nil {
		t.Fatal(err.Error())
	}

	j, _ := json.Marshal(res)

	log.Println(string(j))
}

func TestParameters(t *testing.T) {
	var client = dify.NewClient(host, apiSecretKey)
	var err error
	ctx := context.Background()

	var res *dify.ParametersResponse
	if res, err = client.Api().Parameters(ctx, &dify.ParametersRequest{
		User: "jiuquan AI",
	}); err != nil {
		t.Fatal(err.Error())
	}

	j, _ := json.Marshal(res)

	log.Println(string(j))
}

func TestRunWorkflow(t *testing.T) {
	client := dify.NewClient(host, apiSecretKey)
	//client := dify.NewClient("https://dify.zhaokm.org", "app-")

	// 测试带图片的工作流请求
	workflowReq := dify.WorkflowRequest{
		Inputs: map[string]interface{}{
			"image_url_new": map[string]string{
				"type":            "image",
				"transfer_method": "remote_url",
				"url":             "https://localhost/1-1.jpg",
			},
		},
		ResponseMode: "blocking",
		User:         "Zhaokm@AWS",
	}

	resp, err := client.API().RunWorkflow(context.Background(), workflowReq)

	if err != nil {
		t.Fatalf("RunWorkflow encountered an error: %v", err)
	}

	// 基本字段验证
	if resp.WorkflowRunID == "" {
		t.Errorf("Expected non-empty WorkflowRunID, got empty")
	}
	if resp.TaskID == "" {
		t.Errorf("Expected non-empty TaskID, got empty")
	}

	// 验证工作流执行状态
	if resp.Data.Status != "succeeded" {
		t.Errorf("Expected workflow status 'succeeded', got: %v", resp.Data.Status)
	}

	// 验证输出和元数据
	if len(resp.Data.Outputs) == 0 {
		t.Errorf("Expected outputs, but got none")
	}
	if resp.Data.ElapsedTime <= 0 {
		t.Errorf("Expected positive ElapsedTime, but got: %v", resp.Data.ElapsedTime)
	}
	if resp.Data.TotalSteps <= 0 {
		t.Errorf("Expected positive TotalSteps, but got: %v", resp.Data.TotalSteps)
	}

	t.Logf("Received workflow response: %+v", resp)
}

func TestRunWorkflowStreaming(t *testing.T) {
	client := dify.NewClient(host, apiSecretKey)

	workflowReq := dify.WorkflowRequest{
		Inputs: map[string]interface{}{
			"image_url_new": map[string]string{
				"type":            "image",
				"transfer_method": "remote_url",
				"url":             "https://localhost/1-1.jpg",
			},
		},
		ResponseMode: "streaming",
		User:         "Zhaokm@AWS",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var (
		mu               sync.Mutex
		workflowStarted  bool
		nodeStarted      bool
		nodeFinished     bool
		workflowFinished bool
		ttsReceived      bool
	)

	// 创建一个实现 EventHandler 接口的处理器
	handler := &testEventHandler{
		t:  t,
		mu: &mu,
		onStreamingResponse: func(resp dify.StreamingResponse) {
			mu.Lock()
			defer mu.Unlock()

			switch resp.Event {
			case dify.EventWorkflowStarted:
				workflowStarted = true
			case dify.EventNodeStarted:
				nodeStarted = true
			case dify.EventNodeFinished:
				nodeFinished = true
				if resp.Data.ExecutionMetadata.TotalTokens > 0 {
					t.Logf("Node used %d tokens", resp.Data.ExecutionMetadata.TotalTokens)
				}
			case dify.EventWorkflowFinished:
				workflowFinished = true
				if resp.Data.Status != "succeeded" {
					t.Errorf("Expected workflow status 'succeeded', got: %v", resp.Data.Status)
				}
			}
		},
		onTTSMessage: func(msg dify.TTSMessage) {
			mu.Lock()
			defer mu.Unlock()

			ttsReceived = true
			if msg.Audio == "" {
				t.Error("Expected non-empty audio data in TTS message")
			}
		},
	}

	err := client.API().RunStreamWorkflowWithHandler(ctx, workflowReq, handler)

	if err != nil {
		t.Fatalf("RunStreamWorkflow encountered an error: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	// 验证是否收到所有预期的事件
	if !workflowStarted {
		t.Error("Expected workflow_started event, but didn't receive it")
	}
	if !nodeStarted {
		t.Error("Expected node_started event, but didn't receive it")
	}
	if !nodeFinished {
		t.Error("Expected node_finished event, but didn't receive it")
	}
	if !workflowFinished {
		t.Error("Expected workflow_finished event, but didn't receive it")
	}
	if !ttsReceived {
		t.Error("Expected TTS message, but didn't receive it")
	}

	t.Log("Streaming workflow test completed successfully")
}

// testEventHandler 实现 EventHandler 接口
type testEventHandler struct {
	t                   *testing.T
	mu                  *sync.Mutex
	onStreamingResponse func(dify.StreamingResponse)
	onTTSMessage        func(dify.TTSMessage)
}

func (h *testEventHandler) HandleStreamingResponse(resp dify.StreamingResponse) {
	if h.onStreamingResponse != nil {
		h.onStreamingResponse(resp)
	}
}

func (h *testEventHandler) HandleTTSMessage(msg dify.TTSMessage) {
	if h.onTTSMessage != nil {
		h.onTTSMessage(msg)
	}
}
