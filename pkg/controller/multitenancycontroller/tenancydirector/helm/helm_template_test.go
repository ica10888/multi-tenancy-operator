package helm

import (
	"os"
	"path"
	"strings"
	"testing"
)

var deployment = `---
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

var service = `---
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
    app: spring-example`


var notes = `---
# Source: spring-example/templates/NOTES.txt

Get the application URL by running these commands:

fullname: spring-example-spring-example`


type replaces struct {
	old string
	new string
}

func TestTemplate(t *testing.T) {
	dir,_ := os.Getwd()
	dir = path.Join(dir,"testdata")
	type args struct {
		repo        string
		releaseName string
		outputDir   string
		showNotes   bool
		stringValues []string
	}
	tests := []struct {
		name    string
		args    args
		wantRes []string
		replaceAll []replaces
		wantErr bool
	}{
		{
			name: "base-test",
			args: args{
				repo:       path.Join(dir,"spring-example"),
				releaseName: "spring-example",
				outputDir:   "",
				showNotes:   false,
				stringValues: []string{},
			},
			wantRes: []string{deployment,service},
			wantErr: false,
		},
		{
			name: "string-values-test",
			args: args{
				repo:       path.Join(dir,"spring-example"),
				releaseName: "spring-example",
				outputDir:   "",
				showNotes:   false,
				stringValues: []string{"image.pullPolicy=Always","resources.limits.limitscpu=3000m"},
			},
			wantRes: []string{deployment,service},
			replaceAll: []replaces{replaces{"imagePullPolicy: IfNotPresent","imagePullPolicy: Always"},replaces{"cpu: 2000m","cpu: 3000m"}},
			wantErr: false,
		},
		{
			name: "num-and-bool-values-test",
			args: args{
				repo:       path.Join(dir,"spring-example"),
				releaseName: "spring-example",
				outputDir:   "",
				showNotes:   false,
				stringValues: []string{"replicaCount=3","terminationGracePeriodSeconds.enabled=true"},
			},
			wantRes: []string{deployment,service},
			replaceAll: []replaces{replaces{"replicas: 1","replicas: 3"},replaces{"      volumes:","      volumes:\n      terminationGracePeriodSeconds: 60"}},
			wantErr: false,
		},
		{
			name: "notes-test",
			args: args{
				repo:       path.Join(dir,"spring-example"),
				releaseName: "spring-example",
				outputDir:   "",
				showNotes:   true,
				stringValues: []string{},
			},
			wantRes: []string{deployment,service,notes},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRes, err := Template(tt.args.repo, tt.args.releaseName, tt.args.outputDir, tt.args.showNotes, tt.args.stringValues)
			if (err != nil) != tt.wantErr {
				t.Errorf("Template() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if  ! compare(gotRes,tt.wantRes,tt.replaceAll) {
				t.Errorf("Template() gotRes = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}

func compare(gotRes string,wantRes []string,replaceAll []replaces) bool{
	for _, r := range replaceAll {
		for i, re := range wantRes {
			wantRes[i] = strings.ReplaceAll(re,r.old,r.new)
		}
	}
	for _, re := range wantRes {
		if ! strings.Contains(gotRes,re) {
			return false
		}
		gotRes = strings.Replace(gotRes,re,"",1)
	}
	gotRes = strings.ReplaceAll(gotRes,"\n","")
	if gotRes == "" {
		return true
	} else {
		return false
	}
}