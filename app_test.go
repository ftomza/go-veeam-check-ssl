package main

import (
	"context"
	"reflect"
	"testing"
)

func TestCheckUrl(t *testing.T) {
	ctx := context.Background()
	type args struct {
		checkedUrl string
	}
	tests := []struct {
		name    string
		args    args
		wantRes *CheckResult
	}{
		{
			name: "ok",
			args: args{
				checkedUrl: "github.com",
			},
			wantRes: &CheckResult{
				URL:      "github.com",
				Hostname: "github.com",
				Port:     "443",
				Result:   "OK",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRes := CheckUrl(ctx, tt.args.checkedUrl)
			gotRes.ValidityExpire = tt.wantRes.ValidityExpire
			if !reflect.DeepEqual(gotRes, tt.wantRes) {
				t.Errorf("CheckUrl() gotRes = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}
