package tenancydirector

import (
	"fmt"
)





func SettingToStringValues(sets map[string]string) (strs []string){
	for k, v := range sets {
		if k != ""{
			strs = append(strs, k + "=" + v)
		}
	}
	return
}

func ErrorsFmt (errFmt string,errs []error) (err error){
	for _, e := range errs {
		errFmt += e.Error()
	}
	return fmt.Errorf("%s",errFmt)
}