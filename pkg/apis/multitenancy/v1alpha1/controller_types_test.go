package v1alpha1

import (
	"fmt"
	"testing"
)

func TestControllerStatus(t *testing.T) {
	sets := make(map[string]string)
	sets["spec.username"] = "root"
	type fields struct {
		UpdatedTenancies []StatusTenancy
	}
	type args struct {
		chartName string
		namespace string
	}
	tests := []struct {
		name   string
		fields fields
		appendArgs   []args
		removeArgs   []args
		want  string
	}{
		{
			"base-test",
			fields{[]StatusTenancy{}},
			[]args{{"kafka","dev"}},
			[]args{{"kafka","dev"}},
			"[]",
		},
		{
			"remove-one-test",
			fields{[]StatusTenancy{}},
			[]args{{"kafka","dev"},{"mysql","dev"}},
			[]args{{"kafka","dev"}},
			"[{dev [{mysql map[] }] [] []}]",
		},
		{
			"with-lists-test",
			fields{[]StatusTenancy{
				{
					"dev",
					[]ChartMessage{{"mysql",sets,"mysqlErr"}},
					[]ReplicationControllerStatus{{"mysql","Deployment","1/1"}},
					[]PodStatus{{"mysql-0","Running"}},
				},
			},
			},
			[]args{{"kafka","dev"}},
			[]args{{"kafka","dev"}},
			"[{dev [{mysql map[spec.username:root] mysqlErr}] [{mysql Deployment 1/1}] [{mysql-0 Running}]}]",
		},
		{
			"more-namespaces-test",
			fields{[]StatusTenancy{}},
			[]args{{"kafka","dev"},{"mysql","dev"},{"redis","test"}},
			[]args{{"kafka","dev"}},
			"[{dev [{mysql map[] }] [] []} {test [{redis map[] }] [] []}]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ut := &ControllerStatus{
				UpdatedTenancies: tt.fields.UpdatedTenancies,
			}
			for _, arg := range tt.appendArgs {
				ut.AppendNamespacedChart(arg.chartName,arg.namespace)
			}
			for _, arg := range tt.removeArgs {
				ut.RemoveNamespacedChart(arg.chartName,arg.namespace)
			}
			if fmt.Sprint(ut.UpdatedTenancies) != tt.want {
				t.Errorf("Template() gotRes = %v, want %v", ut.UpdatedTenancies, tt.want)
			}
		})
	}
}

func TestControllerStatus_UpdateNamespacedChartSettings(t *testing.T) {
	sets := make(map[string]string)
	sets["spec.username"] = "root"
	type fields struct {
		UpdatedTenancies []StatusTenancy
	}
	type args struct {
		chartName string
		namespace string
		sets      map[string]string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			"base-test",
			fields{[]StatusTenancy{
					{
						"dev",
						[]ChartMessage{{"mysql",sets,"mysqlErr"}},
						[]ReplicationControllerStatus{{"mysql","Deployment","1/1"}},
						[]PodStatus{{"mysql-0","Running"}},

					},
				},
			},
			args{"mysql", "dev", sets},
			"[{dev [{mysql map[spec.username:root] mysqlErr}] [{mysql Deployment 1/1}] [{mysql-0 Running}]}]",

		},

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := &ControllerStatus{
				UpdatedTenancies: tt.fields.UpdatedTenancies,
			}
			cs.UpdateNamespacedChartSettings(tt.args.chartName,tt.args.namespace,tt.args.sets)
			if fmt.Sprint(cs.UpdatedTenancies) != tt.want {
				t.Errorf("Template() gotRes = %v, want %v", cs.UpdatedTenancies, tt.want)
			}

		})
	}
}