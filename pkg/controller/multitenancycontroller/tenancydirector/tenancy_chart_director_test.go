package tenancydirector

import (
	"fmt"
	"github.com/ica10888/multi-tenancy-operator/pkg/controller/multitenancycontroller"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"testing"
)

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
		wantObj multitenancycontroller.Kubeapi
		wantErr  bool
	}{
		{
			name: "single-namespaced-test",
			args:     args{deployment,"dev"},
			wantObj:  multitenancycontroller.Kubeapi{"extensions/v1beta1","Deployment","spring-example","dev"},
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
			config := rest.Config{}
			c,_ := client.New(&config,client.Options{})
			gotObjs, err := Deserializer(c,tt.args.data,tt.args.namespace,false)
			if (err != nil) != tt.wantErr {
				t.Errorf("Deserializer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotObjs.Kubeapi != tt.wantObj {
				t.Errorf("Deserializer() gotObjs = %v, want %v", gotObjs, tt.wantObj)
			}
		})
	}
}

func Test_immutableFieldSolver(t *testing.T) {
	type args struct {
		objJson  string
		struJson string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"service-test",
			args{
				objJson: `{ "apiVersion": "v1", "kind": "Service", "metadata": { "name": "demo-service", "resourceVersion": "12345" }, "spec": { "clusterIP": "127.0.0.1" } }`,
				struJson: `{ "apiVersion": "v1", "kind": "Service", "metadata": { "name": "demo-service", "resourceVersion": "0" } }`,
			},
			"&{map[apiVersion:v1 kind:Service metadata:map[name:demo-service resourceVersion:0] spec:map[clusterIP:127.0.0.1]]}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj , _, _ := unstructured.UnstructuredJSONScheme.Decode([]byte(tt.args.objJson),nil, nil)
			stru , _, _ := unstructured.UnstructuredJSONScheme.Decode([]byte(tt.args.struJson),nil, nil)
			immutableFieldSolver(obj.(*unstructured.Unstructured),stru.(*unstructured.Unstructured))
			if fmt.Sprint(stru) != tt.want {
				t.Errorf("Template() gotRes = %v, want %v", stru, tt.want)
			}

		})
	}
}