package tenancyscheduler

import (
	"fmt"
	"strings"
)

func SettingToStringValues(sets map[string]string) (strs []string) {
	for k, v := range sets {
		if k != "" {
			strs = append(strs, k+"="+v)
		}
	}
	return
}

func ErrorsFmt(errFmt string, errs []error) (err error) {
	for _, e := range errs {
		errFmt += "\n"
		errFmt += e.Error()
	}
	return fmt.Errorf("%s", errFmt)
}

func conversionCheckDataList(data string) []string {
	var checkDatas []string
	datas := strings.Split(data, "---")
	for _, s := range datas {
		if !(strings.Trim(s, "\n") == "") {
			checkDatas = append(checkDatas, s)
		}
	}
	return checkDatas
}

func removeListIfNotChanged(checkDatas, checkStateDatas []string) []string {
	var updateDatas []string
	for _, data := range checkDatas {
		needAppend := true
		for _, stateData := range checkStateDatas {
			if data == stateData {
				needAppend = false
				break
			}
		}
		if needAppend {
			updateDatas = append(updateDatas, data)
		}
	}
	return updateDatas
}
