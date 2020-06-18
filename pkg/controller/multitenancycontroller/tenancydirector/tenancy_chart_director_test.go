package tenancydirector

import (
	"strings"
	"testing"
)


var service = `
---
# Source: spring-example/templates/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: spring-boot-demo
spec:
  type: ClusterIP
  ports:
  - port: 8761
    targetPort: 8761
    protocol: TCP
    name: http-port
  selector:
    app: spring-example
`

var deployment = `
---
# Source: spring-example/templates/deployment.yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: spring-example
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: spring-example
    spec:
      containers:
      - name: spring-example
        image: springcloud/eureka:latest
`

func TestDeserializer(t *testing.T) {
	type args struct {
		data string
		namespace string
	}
	tests := []struct {
		name     string
		args     args
		wantObjs []Kubeapi
		wantErr  bool
	}{
		{
			name: "single-namespaced-test",
			args:     args{deployment,"dev"},
			wantObjs: []Kubeapi{{"extensions/v1beta1","Deployment","spring-example","dev"}},
			wantErr:  false,
		},
		{
			name: "plural-test",
			args:     args{service + deployment,""},
			wantObjs: []Kubeapi{{"v1","Service","spring-boot-demo",""},{"extensions/v1beta1","Deployment","spring-example",""}},
			wantErr:  false,
		},
		{
			name: "error-type-test",
			args:     args{strings.ReplaceAll(deployment,"name: spring-example","name: true"),""},
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotObjs, err := Deserializer(tt.args.data,tt.args.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("Deserializer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for  i := 0 ;i < len(gotObjs) ;i++ {
				if gotObjs[i].Kubeapi != tt.wantObjs[i] {
					t.Errorf("Deserializer() gotObjs = %v, want %v", gotObjs, tt.wantObjs)
				}
			}
		})
	}
}