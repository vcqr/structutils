package structutils

import (
	"reflect"
	"strings"
)

type StructUtil interface {
	CopyProperties(target, source interface{}) error
}

type StructUtils struct {
	assert *Assert
}

func NewStructUtils() *StructUtils {
	assert := NewAssert()

	return &StructUtils{
		assert: assert,
	}
}

type ReflectInfo struct {
	t reflect.Type
	v reflect.Value
}

// 拷贝结构体相关属性
// 拷贝条件：
//      a. 字段名字一样
//      b. 字段类型一样
//      c. 包不同，但类型名称一样
func (st *StructUtils) CopyProperties(target, source interface{}) error {
	srcV, srcT, err := st.getSourceReflectInfo(source)
	if err != nil {
		return err
	}

	destV, destT, err := st.getTargetReflectInfo(target)
	if err != nil {
		return err
	}

	srcMap := make(map[string]*ReflectInfo)

	for i := 0; i < srcT.NumField(); i++ {
		rf := &ReflectInfo{
			t: srcT.Field(i).Type,
			v: srcV.Field(i),
		}

		srcMap[srcT.Field(i).Name] = rf
	}

	for i := 0; i < destT.NumField(); i++ {
		destFieldV := destV.Field(i)
		if !destFieldV.CanSet() {
			continue
		}

		destFieldT := destT.Field(i)

		if ri, ok := srcMap[destFieldT.Name]; ok {
			if ri.t == destFieldT.Type {
				destV.Field(i).Set(ri.v)
			} else {
				riEndName := st.getTypeEndName(ri.t.String())
				dtEndName := st.getTypeEndName(destFieldT.Type.String())

				if riEndName == dtEndName {
					if destFieldV.Kind() == reflect.Ptr {
						// 目标结构是指针，则需要初始化
						if destFieldV.IsNil() {
							destFieldV.Set(reflect.New(destFieldT.Type.Elem()))
						}

						err := st.CopyProperties(destFieldV.Interface(), ri.v.Interface())
						if err != nil {
							return err
						}
					} else {
						err := st.CopyProperties(destFieldV.Addr().Interface(), ri.v.Interface())
						if err != nil {
							return err
						}
					}
				}
			}
		}
	}

	return nil
}

func (st *StructUtils) copyProperties(dest, src interface{}) error {
	return st.CopyProperties(dest, src)
}

// 获取不同包下的同名类型的名字
func (st *StructUtils) getTypeEndName(typeName string) string {
	strArr := strings.Split(typeName, ".")
	return strArr[len(strArr)-1]
}

// 反射出源结构体信息
func (st *StructUtils) getSourceReflectInfo(obj interface{}) (reflect.Value, reflect.Type, error) {
	st.assert.NotNil(obj, "CopyProperties: Source must not be nil")

	v := reflect.ValueOf(obj)
	t := reflect.TypeOf(obj)

	if t.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return reflect.Zero(v.Type()), t, &StructUtilsError{"CopyProperties: Source are not struct", nil, obj}
	}

	return v, t, nil
}

// 反射出目标结构体信息
func (st *StructUtils) getTargetReflectInfo(obj interface{}) (reflect.Value, reflect.Type, error) {
	st.assert.NotNil(obj, "CopyProperties: Target must not be nil")

	v := reflect.ValueOf(obj)
	t := reflect.TypeOf(obj)

	if t.Kind() != reflect.Ptr {
		return reflect.Zero(v.Type()), t, &StructUtilsError{"CopyProperties: Target are not ptr", obj, nil}
	} else {
		v = v.Elem()
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return reflect.Zero(v.Type()), t, &StructUtilsError{"CopyProperties: Target are not struct", obj, nil}
	}

	return v, t, nil
}