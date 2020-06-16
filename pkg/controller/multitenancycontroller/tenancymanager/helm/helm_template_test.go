package helm

import (
	"os"
	"testing"
)

func TestTemplate(t *testing.T) {
	dir, _ := os.Getwd()
	print(dir)
	type args struct {
		repo        string
		releaseName string
		outputDir   string
		showNotes   bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "for-test",
			args: args{
				repo:       "",
				releaseName: "",
				outputDir:   "",
				showNotes:   false,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Template(tt.args.repo, tt.args.releaseName, tt.args.outputDir, tt.args.showNotes); (err != nil) != tt.wantErr {
				t.Errorf("Template() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}