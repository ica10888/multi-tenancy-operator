package helm

import (
	"os"
	"path"
	"testing"
)

var baseTestWantRes =
`---
# Source: spring-example/templates/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: spring-boot-demo
  labels:
    chart: spring-example-dev
spec:
  type: ClusterIP
  ports:
  - port: 8761
    targetPort: 8761
    protocol: TCP
    name: http-port
  selector:
    app: spring-example
---
# Source: spring-example/templates/deployment.yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: spring-example
  labels:
    app: spring-app
    chart: spring-example-dev
spec:
  replicas: 1
  strategy:
    rollingUpdate:
      maxSurge: 15%
      maxUnavailable: 15%
  template:
    metadata:
      labels:
        app: spring-example
      annotations:
        prometheus.io/path: /prometheus
        prometheus.io/scrape: "true"
        
    spec:
      containers:
      - name: spring-example
        image: springcloud/eureka:latest
        imagePullPolicy: IfNotPresent
        env:
        - name: MY_NODE_NAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        - name: MY_POD_IP
          valueFrom:
            fieldRef:
              fieldPath: status.podIP
        ports:
        - containerPort: 8761
        resources:
          limits:
            cpu: 2000m
            memory: 512Mi
          requests:
            cpu: 50m
            memory: 512Mi
        volumeMounts:
      volumes:`



func TestTemplate(t *testing.T) {
	dir,_ := os.Getwd()
	dir = path.Join(dir,"testdata")
	type args struct {
		repo        string
		releaseName string
		outputDir   string
		showNotes   bool
	}
	tests := []struct {
		name    string
		args    args
		wantRes string
		wantErr bool
	}{
		{
			name: "base-test",
			args: args{
				repo:       path.Join(dir,"spring-example"),
				releaseName: "spring-example",
				outputDir:   "",
				showNotes:   false,
			},
			wantRes: baseTestWantRes,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRes, err := Template(tt.args.repo, tt.args.releaseName, tt.args.outputDir, tt.args.showNotes)
			if (err != nil) != tt.wantErr {
				t.Errorf("Template() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotRes != tt.wantRes {
				t.Errorf("Template() gotRes = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}