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

package server

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/isrc-cas/gt/config"
	"github.com/isrc-cas/gt/util"
)

func TestUser(t *testing.T) {
	args := []string{
		"server",
		"-config", "./testdata/config.yaml",
		"-users", "./testdata/users.yaml",
		"-id", "id1",
		"-secret", "secret1-overwrite-overwrite",
		"-id", "id5",
		"-secret", "secret5",
		"-tcpNumber", "7",
		"-tcpRange", "7-7",
		"-tcpNumber", "8",
		"-tcpRange", "8-8",
		"-tcpNumber", "11", // 测试优先级是否高于 config 配置文件
		"-tcpRange", "1-1",
		"-hostRegex", "^b$",
		"-hostRegex", "^e$",
	}
	s, err := New(args, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = s.users.mergeUsers(s.config.Users, nil, nil)
	if err != nil {
		return
	}
	u := make(map[string]user)
	err = config.Yaml2Interface(s.config.Options.Users, u)
	if err != nil {
		return
	}
	err = s.users.mergeUsers(u, s.config.IDs, s.config.Secrets)
	if err != nil {
		return
	}
	err = s.parseTCPs()
	if err != nil {
		return
	}
	err = s.parseHost()
	if err != nil {
		return
	}

	expectedResult := users{}
	globalTCPs := []tcp{
		{
			Range:  "1-1",
			Number: 11,
			PortRange: &util.PortRange{
				Min: 1,
				Max: 1,
			},
		},
		{
			Range:  "2-2",
			Number: 2,
			PortRange: &util.PortRange{
				Min: 2,
				Max: 2,
			},
		},
		{
			Range:  "7-7",
			Number: 7,
			PortRange: &util.PortRange{
				Min: 7,
				Max: 7,
			},
		},
		{
			Range:  "8-8",
			Number: 8,
			PortRange: &util.PortRange{
				Min: 8,
				Max: 8,
			},
		},
	}
	globalHostRegexStr := config.Slice[string]{
		"^a$",
		"^b$",
		"^e$",
	}
	var globalHostRegex []*regexp.Regexp
	for _, str := range globalHostRegexStr {
		regex, err := regexp.Compile(str)
		if err != nil {
			t.Fatal(err)
		}
		globalHostRegex = append(globalHostRegex, regex)
	}
	expectedResult.Store("id1", user{
		Secret: "secret1-overwrite-overwrite",
		TCPs:   globalTCPs,
		Host: host{
			RegexStr: &globalHostRegexStr,
			Regex:    &globalHostRegex,
		},
	})
	expectedResult.Store("id2", user{
		Secret: "secret2-overwrite",
		TCPs: []tcp{
			{
				Range:  "5-5",
				Number: 5,
				PortRange: &util.PortRange{
					Min: 5,
					Max: 5,
				},
			},
			{
				Range:  "6-6",
				Number: 6,
				PortRange: &util.PortRange{
					Min: 6,
					Max: 6,
				},
			},
		},
		Host: host{
			RegexStr: &config.Slice[string]{},
			Regex: func() *[]*regexp.Regexp {
				var result []*regexp.Regexp
				return &result
			}(),
		},
	})
	expectedResult.Store("id3", user{
		Secret: "secret3",
		TCPs: []tcp{
			{
				Range:  "3-3",
				Number: 3,
				PortRange: &util.PortRange{
					Min: 3,
					Max: 3,
				},
			},
			{
				Range:  "4-4",
				Number: 4,
				PortRange: &util.PortRange{
					Min: 4,
					Max: 4,
				},
			},
		},
		Host: host{
			RegexStr: &config.Slice[string]{
				"^c$",
				"^d$",
			},
			Regex: func() *[]*regexp.Regexp {
				var result []*regexp.Regexp
				regex, err := regexp.Compile("^c$")
				if err != nil {
					t.Fatal(err)
				}
				result = append(result, regex)
				regex, err = regexp.Compile("^d$")
				if err != nil {
					t.Fatal(err)
				}
				result = append(result, regex)
				return &result
			}(),
		},
	})
	expectedResult.Store("id4", user{
		Secret: "secret4",
		TCPs:   globalTCPs,
		Host: host{
			RegexStr: &globalHostRegexStr,
			Regex:    &globalHostRegex,
		},
	})
	expectedResult.Store("id5", user{
		Secret: "secret5",
		TCPs:   globalTCPs,
		Host: host{
			RegexStr: &globalHostRegexStr,
			Regex:    &globalHostRegex,
		},
	})
	expectedResult.Range(func(key, value1 interface{}) bool {
		value2, ok := s.users.Load(key)
		if !ok {
			t.Fatalf("%q does not exist", key)
		}
		user1 := value1.(user)
		user2 := value2.(user)

		if !reflect.DeepEqual(user1.Secret, user2.Secret) {
			t.Fatalf("user secret does not match\n%+v\n%+v", user1.Secret, user2.Secret)
		}

		if len(user1.TCPs) != len(user2.TCPs) {
			t.Fatal("user TCPs length does not match")
		}
		for _, tcp1 := range user1.TCPs {
			contains := false
			for _, tcp2 := range user2.TCPs {
				if reflect.DeepEqual(&tcp1, &tcp2) {
					contains = true
					break
				}
			}
			if !contains {
				t.Fatalf("TCPs item does not match")
			}
		}

		if len(*user1.Host.Regex) != len(*user2.Host.Regex) {
			t.Fatal("user hostRegex length does not match")
		}
		for _, regex1 := range *user1.Host.Regex {
			contains := false
			for _, regex2 := range *user2.Host.Regex {
				if reflect.DeepEqual(regex1, regex2) {
					contains = true
					break
				}
			}
			if !contains {
				t.Fatalf("hostRegex item does not match")
			}
		}

		return true
	})
}
