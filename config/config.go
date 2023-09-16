// Copyright (c) 2022 Institute of Software, Chinese Academy of Sciences (ISCAS)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gopkg.in/yaml.v3"
)

var position atomic.Uint32 // 用于需要存储位置的变量存储相对位置信息
var positionMtx sync.Mutex

// ParseFlags parses args and sets the result to config and options.
func ParseFlags(args []string, config, options interface{}) error {
	if len(args) < 2 {
		return nil
	}
	positionMtx.Lock()
	defer positionMtx.Unlock()
	position.Store(0)

	flagSet, n2fi, configPath := registerFlags(args[0], options)
	err := flagSet.Parse(args[1:])
	if err != nil {
		return err
	}

	if configPath != nil {
		err = Yaml2Interface(*configPath, config)
		if err != nil {
			return err
		}
	}

	return copyFlagsValue(options, flagSet, n2fi)
}

// Yaml2Interface 解析 yaml 配置文件
func Yaml2Interface(path string, dstInterface interface{}) (err error) {
	// 当参数不合法时，直接返回 nil
	if len(path) == 0 || dstInterface == nil {
		return
	}

	file, err := os.Open(path)
	if err != nil {
		err = fmt.Errorf("open yaml file %q failed: %v", path, err)
		return
	}
	defer func() {
		if e := file.Close(); e != nil {
			if err != nil {
				err = fmt.Errorf("%w and %s", err, e.Error())
			} else {
				err = e
			}
		}
	}()

	err = yaml.NewDecoder(file).Decode(dstInterface)
	if err != nil {
		err = fmt.Errorf("decode yaml file %q failed: %v", path, err)
		return
	}

	return
}

func copyFlagsValue(dst interface{}, src *flag.FlagSet, name2FieldIndex map[string]int) (err error) {
	dstValue := reflect.ValueOf(dst).Elem()
	src.Visit(func(f *flag.Flag) {
		i, ok := name2FieldIndex[f.Name]
		if !ok {
			return
		}
		field := dstValue.Field(i)
		if field.Kind() == reflect.Ptr && field.IsNil() {
			newValue := reflect.New(field.Type().Elem())
			field.Set(newValue)
			field = field.Elem()
		}
		fieldType := field.Type()
		flagValue := reflect.ValueOf(f.Value.(flag.Getter).Get())
		flagValueType := flagValue.Type()

		if fieldType == reflect.TypeOf(Duration{}) {
			durationValue := flagValue.Interface().(time.Duration)
			field.Set(reflect.ValueOf(Duration{Duration: durationValue}))
			return
		}

		if !flagValueType.AssignableTo(fieldType) {
			if flagValueType.ConvertibleTo(fieldType) {
				flagValue = flagValue.Convert(fieldType)
			} else {
				err = fmt.Errorf("can't set flagValue(%v) to field(%v)", flagValueType, fieldType)
				return
			}
		}
		field.Set(flagValue)
	})
	return
}

func registerFlags(flagSetName string, options interface{}) (flagSet *flag.FlagSet, name2FieldIndex map[string]int, config *string) {
	flagSet = flag.NewFlagSet(flagSetName, flag.ExitOnError)
	name2FieldIndex = make(map[string]int)
	value := reflect.ValueOf(options).Elem()
	valueType := value.Type()
	for i := 0; i < valueType.NumField(); i++ {
		field := value.Field(i)
		fieldType := valueType.Field(i)
		name, ok := fieldType.Tag.Lookup("yaml")
		if !ok || name == "-" {
			name, ok = fieldType.Tag.Lookup("arg")
			if !ok {
				continue
			}
		}

		// 用这种方式可以处理 yaml:"xxx,omitempty" 的情况
		name = strings.Split(name, ",")[0]

		name2FieldIndex[name] = i
		usage := fieldType.Tag.Get("usage")

		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				field = reflect.New(field.Type().Elem())
			}
			field = field.Elem()
		}
		flagValueType := reflect.TypeOf((*flag.Value)(nil)).Elem()
		if reflect.PointerTo(field.Type()).Implements(flagValueType) {
			newValue := reflect.New(field.Type())
			flagSet.Var(newValue.Interface().(flag.Value), name, usage)
			continue
		} else if field.Type() == reflect.TypeOf((*time.Duration)(nil)).Elem() {
			flagSet.Duration(name, field.Interface().(time.Duration), usage)
			continue
		}

		// 用这种方式可以处理 type xxx string 的情况
		switch field.Kind() {
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
			flagSet.Uint(name, uint(field.Uint()), usage)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
			flagSet.Int(name, int(field.Int()), usage)
		case reflect.Uint64:
			flagSet.Uint64(name, field.Uint(), usage)
		case reflect.Int64:
			flagSet.Int64(name, field.Int(), usage)
		case reflect.Float64:
			flagSet.Float64(name, field.Float(), usage)
		case reflect.String:
			if name != "config" {
				flagSet.String(name, field.String(), usage)
			} else {
				config = flagSet.String(name, field.String(), usage)
			}
		case reflect.Bool:
			flagSet.Bool(name, field.Bool(), usage)
		default:
			panic(fmt.Sprintf("not supported type %s of field %s", fieldType.Type.Kind().String(), fieldType.Name))
		}
	}
	return
}

// ShowUsage generates and prints the usage document of options.
func ShowUsage(options interface{}) {
	if options == nil {
		panic("options can not be nil")
	}
	flagSet, _, _ := registerFlags(os.Args[0], options)
	flagSet.Usage()
}
