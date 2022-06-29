package okta

import (
	"reflect"
	"testing"

	"github.com/okta/okta-sdk-golang/okta"
)

func TestOktaService_GetUsersFromGroup(t *testing.T) {
	type args struct {
		groupName string
	}
	tests := []struct {
		name    string
		args    args
		want    []*okta.User
		wantErr bool
	}{
		{
			name: "test",
			args: args{groupName: "engineering"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := NewOktaService(OktaConfig{})
			got, err := o.GetUsersFromGroup(tt.args.groupName)
			if (err != nil) != tt.wantErr {
				t.Errorf("OktaService.GetUsersFromGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OktaService.GetUsersFromGroup() = %v, want %v", got, tt.want)
			}
		})
	}
}
