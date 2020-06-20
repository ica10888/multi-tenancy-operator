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
	Namespace string `json:"namespace"`
	ChartNames []string `json:"chartNames"`
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
	newUpdatedTenancies := []StatusTenancy{}
	chartNames := []string{}
	rcList := []ReplicationControllerStatus{}
	podList := []PodStatus{}
	index := -1
	for i, tenancy := range ut.UpdatedTenancies {
		if tenancy.Namespace == namespace {
			chartNames = tenancy.ChartNames
			rcList = tenancy.ReplicationControllerStatusList
			podList = tenancy.PodStatusList
			index = i
			break
		}
	}
	newChartNames := append(chartNames,chartName)
	st := StatusTenancy{
		namespace,
		newChartNames,
		rcList,
		podList,
	}
	if index == -1 {
		newUpdatedTenancies = append(ut.UpdatedTenancies, st)
	} else {
		newUpdatedTenancies = append(append(ut.UpdatedTenancies[:index], st), ut.UpdatedTenancies[index+1:]...)
	}

	ut.UpdatedTenancies = newUpdatedTenancies

}




func (ut *ControllerStatus) RemoveNamespacedChart(chartName,namespace string) {
	newUpdatedTenancies := []StatusTenancy{}
	rcList := []ReplicationControllerStatus{}
	podList := []PodStatus{}
	st := StatusTenancy{}
	index := -1
	for i, tenancy := range ut.UpdatedTenancies {
		if tenancy.Namespace == namespace {
			for j, name := range tenancy.ChartNames {
				if name == chartName {
					rcList = tenancy.ReplicationControllerStatusList
					podList = tenancy.PodStatusList
					newChartNames := append(tenancy.ChartNames[:j],tenancy.ChartNames[j+1:]...)
					st = StatusTenancy{
						namespace,
						newChartNames,
						rcList,
						podList,
					}
					break
				}
			}
			index = i
			break
		}
	}
	if len(st.ChartNames) == 0 {
		newUpdatedTenancies = append(ut.UpdatedTenancies[:index],ut.UpdatedTenancies[index+1:]...)
	} else {
		newUpdatedTenancies = append(append(ut.UpdatedTenancies[:index], st), ut.UpdatedTenancies[index+1:]...)
	}
	ut.UpdatedTenancies = newUpdatedTenancies
}
