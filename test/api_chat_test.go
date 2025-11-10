package test

import (
	"context"
	"reflect"
	"testing"

	"github.com/cloudernative/dify-sdk-go"
)

func TestAPI_ChatMessages(t *testing.T) {
	var c = &dify.ClientConfig{
		Host:             host,
		DefaultAPISecret: apiSecretKey,
	}
	var client = dify.NewClientWithConfig(c)
	api := client.Api()
	type args struct {
		ctx context.Context
		req *dify.ChatMessageRequest
	}
	tests := []struct {
		name     string
		args     args
		wantResp *dify.ChatMessageResponse
		wantErr  bool
	}{
		{
			name: "test1",
			args: args{
				ctx: context.Background(),
				req: &dify.ChatMessageRequest{
					Inputs: nil,
					Query:  "IF([销售额] > 1000, CONVERT([订单日期], \"YYYY-MM\"), LEFT([产品编号], 3))\n\n",
					User:   "system",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResp, err := api.ChatMessages(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ChatMessages() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResp, tt.wantResp) {
				t.Errorf("ChatMessages() gotResp = %v, want %v", gotResp, tt.wantResp)
			}
		})
	}
}
