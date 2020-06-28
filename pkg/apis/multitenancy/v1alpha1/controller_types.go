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
	ReleaseName *string `json:"releaseName,omitempty"`
	Settings []Setting `json:"settings"`
}

type Setting struct {
	Key string `json:"key"`
	Value string `json:"value"`
}


type StatusTenancy struct {
	Namespace string `json:"namespace"`
	ChartMessages []ChartMessage `json:"chartMessages"`
	ReplicationControllerStatusList []ReplicationControllerStatus  `json:"replicationControllerStatus"`
	PodStatusList []PodStatus `json:"podStatus"`
}

type ChartMessage struct {
	ChartName string `json:"chartName"`
	SettingMap map[string]string `json:"settingMap"`
	ErrorMessage *string `json:"errorMessage,omitempty"`
}

type PodStatus struct {
	PodName string `json:"podName"`
	Phase string `json:"phase"`
}

type ReplicationControllerStatus struct {
	ReplicationControllerName string `json:"replicationControllerName"`
	ApiVersion string `json:"apiVersion"`
	Kind string `json:"kind"`
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


func (cs *ControllerStatus) AppendNamespacedChart(chartName,namespace string){
	newUpdatedTenancies := []StatusTenancy{}
	chartMessages := []ChartMessage{}
	rcList := []ReplicationControllerStatus{}
	podList := []PodStatus{}
	index := -1
	for i, tenancy := range cs.UpdatedTenancies {
		if tenancy.Namespace == namespace {
			chartMessages = tenancy.ChartMessages
			rcList = tenancy.ReplicationControllerStatusList
			podList = tenancy.PodStatusList
			index = i
			break
		}
	}
	newChartMessages := append(chartMessages,ChartMessage{chartName,make(map[string]string),nil})
	st := StatusTenancy{
		namespace,
		newChartMessages,
		rcList,
		podList,
	}
	if index == -1 {
		newUpdatedTenancies = append(cs.UpdatedTenancies, st)
	} else {
		newUpdatedTenancies = append(append(cs.UpdatedTenancies[:index], st), cs.UpdatedTenancies[index+1:]...)
	}

	cs.UpdatedTenancies = newUpdatedTenancies

}

func (cs *ControllerStatus) RemoveNamespacedChart(chartName,namespace string) {
	newUpdatedTenancies := []StatusTenancy{}
	rcList := []ReplicationControllerStatus{}
	podList := []PodStatus{}
	st := StatusTenancy{}
	index := -1
	for i, tenancy := range cs.UpdatedTenancies {
		if tenancy.Namespace == namespace {
			for j, chartMessage := range tenancy.ChartMessages {
				if chartMessage.ChartName == chartName {
					rcList = tenancy.ReplicationControllerStatusList
					podList = tenancy.PodStatusList
					newChartNames := append(tenancy.ChartMessages[:j],tenancy.ChartMessages[j+1:]...)
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
	if len(st.ChartMessages) == 0 {
		newUpdatedTenancies = append(cs.UpdatedTenancies[:index],cs.UpdatedTenancies[index+1:]...)
	} else {
		newUpdatedTenancies = append(append(cs.UpdatedTenancies[:index], st), cs.UpdatedTenancies[index+1:]...)
	}
	cs.UpdatedTenancies = newUpdatedTenancies
}

func (cs *ControllerStatus) UpdateNamespacedChartReplicationControllerStatusReady(namespace,rCName,apiVersion,kind,ready string) bool {
	for _, tenancy := range cs.UpdatedTenancies {
		if tenancy.Namespace == namespace {
			list := tenancy.ReplicationControllerStatusList
			for i, status := range list {
				if status.ReplicationControllerName == rCName && status.ApiVersion == apiVersion && status.Kind == kind {
					if status.Ready == ready {
						return false
					} else {
						newRC := ReplicationControllerStatus{rCName,apiVersion,kind,ready}
						newList := append(append(tenancy.ReplicationControllerStatusList[:i], newRC), tenancy.ReplicationControllerStatusList[i+1:]...)
						cs.updateNamespacedChartForNewStatusTenancyFunc(namespace,func (st *StatusTenancy) StatusTenancy{
							return StatusTenancy{
								st.Namespace,
								st.ChartMessages,
								newList,
								st.PodStatusList,
							}
						})
						return true
					}
				}
			}
			break
		}
	}
	return false
}

func (cs *ControllerStatus) RemoveNamespacedChartReplicationControllerStatusListIfExist(namespace,rCName,apiVersion,kind string) {
	for _, tenancy := range cs.UpdatedTenancies {
		if tenancy.Namespace == namespace {
			list := tenancy.ReplicationControllerStatusList
			for i, status := range list {
				if status.ReplicationControllerName == rCName && status.ApiVersion == apiVersion && status.Kind == kind {
					newList := append(tenancy.ReplicationControllerStatusList[:i], tenancy.ReplicationControllerStatusList[i+1:]...)
					cs.updateNamespacedChartForNewStatusTenancyFunc(namespace,func (st *StatusTenancy) StatusTenancy{
						return StatusTenancy{
							st.Namespace,
							st.ChartMessages,
							newList,
							st.PodStatusList,
						}
					})
					break
				}
			}
			break
		}
	}

}

func (cs *ControllerStatus) AppendNamespacedChartReplicationControllerStatusList(namespace,rCName,apiVersion,kind string) {
	for _, tenancy := range cs.UpdatedTenancies {
		if tenancy.Namespace == namespace {
			list := tenancy.ReplicationControllerStatusList
			for _, status := range list {
				if status.ReplicationControllerName == rCName && status.ApiVersion == apiVersion {
					return
				}
			}
			list = append(list,ReplicationControllerStatus{rCName,apiVersion,kind,""})
			cs.updateNamespacedChartForNewStatusTenancyFunc(namespace,func (st *StatusTenancy) StatusTenancy{
				return StatusTenancy{
					st.Namespace,
					st.ChartMessages,
					list,
					st.PodStatusList,
				}
			})
		}
	}
}

func (cs *ControllerStatus) ApplyNamespacedChartPodStatus(namespace,podName,phase string) bool {
	for _, tenancy := range cs.UpdatedTenancies {
		if tenancy.Namespace == namespace {
			for i, status := range tenancy.PodStatusList {
				if status.PodName == podName {
					if status.Phase == phase {
						return false
					} else {
						newPs := PodStatus{podName,phase}
						list := append(append(tenancy.PodStatusList[:i], newPs),tenancy.PodStatusList[i+1:]...)
						cs.updateNamespacedChartForNewStatusTenancyFunc(namespace,func (st *StatusTenancy) StatusTenancy{
							return StatusTenancy{
								st.Namespace,
								st.ChartMessages,
								st.ReplicationControllerStatusList,
								list,
							}
						})
						return true
					}
				}
			}
			//not exist
			newPs := PodStatus{podName,phase}
			list := append(tenancy.PodStatusList, newPs)
			cs.updateNamespacedChartForNewStatusTenancyFunc(namespace,func (st *StatusTenancy) StatusTenancy{
				return StatusTenancy{
					st.Namespace,
					st.ChartMessages,
					st.ReplicationControllerStatusList,
					list,
				}
			})
			return true
		}
	}
	return false
}

func (cs *ControllerStatus) RemoveNamespacedChartPodStatus(namespace,podName string) bool {
	for _, tenancy := range cs.UpdatedTenancies {
		if tenancy.Namespace == namespace {
			for i, status := range tenancy.PodStatusList {
				if status.PodName == podName {
					list := append(tenancy.PodStatusList[:i],tenancy.PodStatusList[i+1:]...)
					cs.updateNamespacedChartForNewStatusTenancyFunc(namespace,func (st *StatusTenancy) StatusTenancy{
						return StatusTenancy{
							st.Namespace,
							st.ChartMessages,
							st.ReplicationControllerStatusList,
							list,
						}
					})
					return true
				}
			}
			break
		}
	}
	return false
}



func (cs *ControllerStatus) updateNamespacedChartForNewStatusTenancyFunc(namespace string,f func (*StatusTenancy) StatusTenancy) {
	needUpdate := false
	if namespace == "" {
		return
	}
	newUpdatedTenancies := []StatusTenancy{}
	newTenancy := StatusTenancy{}
	index := -1
	for i, tenancy := range cs.UpdatedTenancies {
		if tenancy.Namespace == namespace {

			newTenancy = f(&tenancy)

			index = i
			needUpdate = true
			break
		}
	}
	if needUpdate {
		newUpdatedTenancies = append(append(cs.UpdatedTenancies[:index], newTenancy), cs.UpdatedTenancies[index+1:]...)
		cs.UpdatedTenancies = newUpdatedTenancies
	}
}

func (cs *ControllerStatus) UpdateNamespacedChartSettings(chartName,namespace string,sets map[string]string){
	cs.updateNamespacedChartForNewMessageFunc(chartName,namespace,
		func (cm *ChartMessage) ChartMessage {
			return ChartMessage{
				cm.ChartName,
				sets,
				cm.ErrorMessage,
			}
		})
}

func (cs *ControllerStatus) UpdateNamespacedChartErrorMessage(chartName,namespace string,err error){
	if err == nil {
		cs.updateNamespacedChartForNewMessageFunc(chartName,namespace,
			func (cm *ChartMessage) ChartMessage {
				return ChartMessage{
					cm.ChartName,
					cm.SettingMap,
					nil,
				}
			})
	} else {
		errorMessage := err.Error()
		cs.updateNamespacedChartForNewMessageFunc(chartName,namespace,
			func (cm *ChartMessage) ChartMessage {
				return ChartMessage{
					cm.ChartName,
					cm.SettingMap,
					&errorMessage,
				}
			})
	}
}

func (cs *ControllerStatus) updateNamespacedChartForNewMessageFunc(chartName,namespace string,f func (*ChartMessage) ChartMessage){
	needUpdate := false
	if namespace == "" {
		return
	}
	newUpdatedTenancies := []StatusTenancy{}
	newTenancy := StatusTenancy{}
	newMessage := ChartMessage{}
	tenancyCopy := StatusTenancy{}
	index := -1
	jndex := -1
	for i, tenancy := range cs.UpdatedTenancies {
		if tenancy.Namespace == namespace {
			for j, message := range tenancy.ChartMessages {
				if message.ChartName == chartName {

					newMessage = f(&message)

					index = i
					jndex = j
					needUpdate = true
					tenancyCopy = tenancy
					break
				}
			}
			break
		}
	}
	if needUpdate {
		charts := tenancyCopy.ChartMessages
		newTenancy = StatusTenancy{
			tenancyCopy.Namespace,
			append(append(charts[:jndex], newMessage), charts[jndex+1:]...),
			tenancyCopy.ReplicationControllerStatusList,
			tenancyCopy.PodStatusList,
		}
		newUpdatedTenancies = append(append(cs.UpdatedTenancies[:index], newTenancy), cs.UpdatedTenancies[index+1:]...)
		cs.UpdatedTenancies = newUpdatedTenancies
	}

}



