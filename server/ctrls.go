package server

import (
	"fmt"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/0x6666/backup/log"
)

var controllerPtrType = reflect.TypeOf(&CtrlBase{})

var controllers = map[string]*controllerType{}

type methodType struct {
	Type reflect.Type
}

type controllerType struct {
	Type reflect.Type
	//methods     map[string]methodType
	baseCtrlIdx [][]int
}

func initNewController(ctrlType *controllerType, c *gin.Context) interface{} {
	var (
		ctrlPtr = reflect.New(ctrlType.Type)
		ctrl    = ctrlPtr.Elem()
	)

	base := &CtrlBase{}
	base.C = c

	for _, index := range ctrlType.baseCtrlIdx {
		ctrl.FieldByIndex(index).Set(reflect.ValueOf(base))
	}

	return ctrlPtr.Interface()
}

func RegisterController(c interface{}) {

	elem := reflect.TypeOf(c).Elem()

	ct := &controllerType{elem /*, map[string]methodType{}*/, nil}
	controllers[elem.Name()] = ct

	ct.baseCtrlIdx = findControllers(elem)

	fmt.Printf("%v\n", ct.baseCtrlIdx)

	/*	mc := elem.NumMethod()

		for i := 0; i < mc; i++ {
			m := elem.Method(i)
			if unicode.IsUpper((rune)(m.Name[0])) {
				ct.methods[m.Name] = methodType{m.Type}
			}
		}
	*/
	log.Info("reg controller: %v", elem.Name())
}

func findControllers(appControllerType reflect.Type) (indexes [][]int) {
	type nodeType struct {
		val   reflect.Value
		index []int
	}
	appControllerPtr := reflect.New(appControllerType)
	queue := []nodeType{{appControllerPtr, []int{}}}
	for len(queue) > 0 {
		var (
			node     = queue[0]
			elem     = node.val
			elemType = elem.Type()
		)
		if elemType.Kind() == reflect.Ptr {
			elem = elem.Elem()
			elemType = elem.Type()
		}
		queue = queue[1:]

		for i := 0; i < elem.NumField(); i++ {
			structField := elemType.Field(i)
			if !structField.Anonymous {
				continue
			}

			fieldValue := elem.Field(i)
			fieldType := structField.Type

			if fieldType == controllerPtrType {
				indexes = append(indexes, append(node.index, i))
				continue
			}

			queue = append(queue,
				nodeType{fieldValue, append(append([]int{}, node.index...), i)})
		}
	}
	return
}
