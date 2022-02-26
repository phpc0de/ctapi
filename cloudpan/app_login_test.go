// Copyright (c) 2020 tickstep.
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

package cloudpan

import (
	"encoding/xml"
	"fmt"
	"testing"
)

func TestAppLogin(t *testing.T) {
	r, e := AppLogin("131687xxxxx@189.cn", "12345xxxxx")
	if e != nil {
		fmt.Println(e)
		return
	}
	fmt.Printf("%+v", r)
}

func TestXmlParse(t *testing.T) {
	data := `
<?xml version="1.0" encoding="UTF-8"?>
<userSignResult>
    <result>1</result>
    <resultTip>获得31M空间</resultTip>
    <activityFlag>1</activityFlag>
    <prizeListUrl>https://m.cloud.189.cn/zhuanti/2016/sign/myPrizeList.jsp</prizeListUrl>
    <buttonTip>点击领取红包</buttonTip>
    <buttonUrl>https://m.cloud.189.cn/zhuanti/2016/sign/index.jsp</buttonUrl>
    <activityTip>你今天还有福利可以领取哟，不领就亏啦！</activityTip>
</userSignResult>
`
	item := &userSignResult{}
	if err := xml.Unmarshal([]byte(data), item); err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v", item)
}

func TestGetSessionByAccessToken(t *testing.T) {
	r, e := getSessionByAccessToken("d17faf30472f470d92f226a0dbc25571")
	if e != nil {
		fmt.Println(e)
		return
	}
	fmt.Printf("%+v", r)
}