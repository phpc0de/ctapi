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
	"encoding/json"
	"github.com/phpc0de/ctapi/cloudpan/apierror"
	"github.com/phpc0de/ctlibgo/logger"
	"strings"
)

type (
	UserVip int

	UserInfo struct {
		// 用户UID
		UserId uint64 `json:"userId"`
		// 用户登录名，一般为 xxx@189.cn
		UserAccount string `json:"userAccount"`
		// 昵称，如果没有设置则为空
		Nickname string `json:"nickname"`
		// 域名称，默认和UserId一样
		DomainName string `json:"domainName"`
		// 189邮箱已使用空间大小
		Used189Size uint64 `json:"used189Size"`
		// 已使用个人空间大小
		UsedSize uint64 `json:"usedSize"`
		// 个人空间总大小
		Quota uint64 `json:"quota"`
		// 会员开始时间
		SuperBeginTime string `json:"superBeginTime"`
		// 会员结束时间
		SuperEndTime string `json:"superEndTime"`
		// 今天是否已签到
		IsSign bool `json:"isSign"`
		// VIP会员标志位
		SuperVip UserVip `json:"superVip"`
	}

	UserDetailInfo struct {
		// 性别 F-女 M-男
		Gender string `json:"gender"`
		// 省代码
		ProvinceCode string `json:"provinceCode"`
		// 城市代码
		CityCode string `json:"cityCode"`
		// 登录名
		UserAccount string `json:"userAccount"`
		// 手机号，模糊处理过的，没有设定则为空
		SafeMobile string `json:"safeMobile"`
		// 域名称
		DomainName string `json:"domainName"`
		// 昵称
		Nickname string `json:"nickname"`
		// 邮箱，没有设定则为空
		Email string `json:"email"`
	}
)

const (
	// VipFamilyGold 家庭黄金会员
	VipFamilyGold UserVip = 99

	// VipGold 黄金会员
	VipGold UserVip = 100

	// VipFamilyPlatnum 家庭铂金会员
	VipFamilyPlatnum UserVip = 199

	// VipPlatnum 铂金会员
	VipPlatnum UserVip = 200

	// VipUser 普通会员
	VipUser UserVip = 0
)

func (p *PanClient) GetUserInfo() (userInfo *UserInfo, error *apierror.ApiError) {
	url := WEB_URL + "/v2/getLoginedInfos.action"
	body, err := p.client.DoGet(url)
	if err != nil {
		logger.Verboseln("get user info failed")
		return nil, apierror.NewApiErrorWithError(err)
	}

	es := &apierror.ErrorResp{}
	if err := json.Unmarshal(body, es); err == nil {
		if es.ErrorCode == "InvalidSessionKey" {
			logger.Verboseln("get user info failed")
			return nil, apierror.NewApiError(apierror.ApiCodeTokenExpiredCode, "登录超时")
		}
	}
	if strings.Contains(string(body), "登录页页面") {
		logger.Verboseln("token expired")
		return nil, apierror.NewApiError(apierror.ApiCodeTokenExpiredCode, "登录超时")
	}

	ui := &UserInfo{}
	if err := json.Unmarshal(body, ui); err != nil {
		logger.Verboseln("get user info failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	return ui, nil
}

func (p *PanClient) GetUserDetailInfo() (userDetailInfo *UserDetailInfo, error *apierror.ApiError) {
	url := WEB_URL + "/v2/getUserDetailInfo.action"
	body, err := p.client.DoGet(url)
	if err != nil {
		logger.Verboseln("get user detail info failed")
		return nil, apierror.NewApiErrorWithError(err)
	}

	es := &apierror.ErrorResp{}
	if err := json.Unmarshal(body, es); err == nil {
		if es.ErrorCode == "InvalidSessionKey" {
			logger.Verboseln("get user detail info failed")
			return nil, apierror.NewApiError(apierror.ApiCodeTokenExpiredCode, "登录超时")
		}
	}

	ui := &UserDetailInfo{}
	if err := json.Unmarshal(body, ui); err != nil {
		logger.Verboseln("get user detail info failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	return ui, nil
}