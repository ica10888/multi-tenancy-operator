package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ControllerSpec defines the desired state of Controller
type ControllerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Tenancies []Tenancy `json:"tenancies"`
}

// ControllerStatus defines the observed state of Controller
type ControllerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	UpdatedTenancies []StatusTenancy `json:"updatedTenancies"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Controller is the Schema for the controllers API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=controllers,scope=Namespaced
type Controller struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ControllerSpec   `json:"spec,omitempty"`
	Status ControllerStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ControllerList contains a list of Controller
type ControllerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Controller `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Controller{}, &ControllerList{})
}

type Tenancy struct {
	Namespace string `json:"namespace"`
	Charts []Chart `json:"charts"`
}

type Chart struct {
	ChartName string `json:"chartName"`
	Settings []Setting `json:"settings"`
}

type Setting struct {
	Key string `json:"key"`
	Value string `json:"value"`
}


type StatusTenancy struct {
	Tenancy
	ReplicationControllerStatusList []ReplicationControllerStatus  `json:"replicationControllerStatus"`
	PodStatusList []PodStatus `json:"podStatus"`
}

type PodStatus struct {
	PodName string `json:"podName"`
	Phase string `json:"phase"`
}

type ReplicationControllerStatus struct {
	ReplicationControllerName string `json:"replicationControllerName"`
	ApiVersion string `json:"apiVersion"`
	Ready string `json:"ready"`
}

func (c *Controller) InitCheck() bool{
	res := false
	if c.Spec.Tenancies == nil {
		c.Spec.Tenancies = []Tenancy{}
		res = true
	}
	if c.Status.UpdatedTenancies == nil {
		c.Status.UpdatedTenancies = []StatusTenancy{}
		res = true
	}
	return res
}


func (ut *ControllerStatus) AppendNamespacedChart(chartName,namespace string){
	var t *StatusTenancy
	namespaceInited := false
	chartInited := false
	for _, tenancy := range ut.UpdatedTenancies {
		if tenancy.Namespace == namespace {
			t = &tenancy
			namespaceInited = true
			break
		}
	}
	if ! namespaceInited {
		t = &StatusTenancy{}
	}
	if t.Charts == nil {
		t.Charts = []Chart{}
	}
	for _, chart := range t.Charts {
		if chart.ChartName == chartName {
			chartInited = true
			break
		}
	}
	if ! chartInited {
		t.Charts = append(t.Charts,Chart{namespace,[]Setting{}})
	}
}

func (ut *ControllerStatus) RemoveNamespacedChart(chartName,namespace string) {
	var t *StatusTenancy
	namespaceInited := false
	for _, tenancy := range ut.UpdatedTenancies {
		if tenancy.Namespace == namespace {
			t = &tenancy
			namespaceInited = true
			break
		}
	}
	if (! namespaceInited) || t.Charts == nil {
		return
	}
	for i, chart := range t.Charts {
		if chart.ChartName == chartName {
			newCharts := append(t.Charts[:i],t.Charts[i+1:]...)
			t.Charts = newCharts
			return
		}
	}
}
