package watcher

import (
	"testing"

	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
)

func TestGetManagedWorkflow(t *testing.T) {
	tests := []struct {
		name    string
		obj     interface{}
		wantOK  bool
		wantNil bool
	}{
		{
			name:    "invalid object type should be ignored",
			obj:     "not-workflow",
			wantOK:  false,
			wantNil: true,
		},
		{
			name:    "nil workflow pointer should be ignored",
			obj:     (*wfv1.Workflow)(nil),
			wantOK:  false,
			wantNil: true,
		},
		{
			name:    "workflow without labels should be ignored",
			obj:     &wfv1.Workflow{},
			wantOK:  false,
			wantNil: true,
		},
		{
			name:    "workflow with non-matching managed-by should be ignored",
			obj:     &wfv1.Workflow{},
			wantOK:  false,
			wantNil: true,
		},
		{
			name:    "workflow with matching managed-by should be accepted",
			obj:     &wfv1.Workflow{},
			wantOK:  true,
			wantNil: false,
		},
	}

	tests[3].obj.(*wfv1.Workflow).Labels = map[string]string{"managed-by": "other-system"}
	tests[4].obj.(*wfv1.Workflow).Labels = map[string]string{"managed-by": "my-devops-system"}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := getManagedWorkflow(tt.obj)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if (got == nil) != tt.wantNil {
				t.Fatalf("got nil = %v, want nil = %v", got == nil, tt.wantNil)
			}
		})
	}
}
