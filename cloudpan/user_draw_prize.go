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
	"fmt"
	"github.com/phpc0de/ctapi/cloudpan/apierror"
	"github.com/phpc0de/ctlibgo/logger"
	"strings"
)

type (
	userDrawPrizeResp struct {
		ActivityId   string `json:"activityId"`
		Description  string `json:"description"`
		IsUsed       int    `json:"isUsed"`
		PrizeGrade   int    `json:"prizeGrade"`
		PrizeId      string `json:"prizeId"`
		PrizeName    string `json:"prizeName"`
		PrizeStatus  int    `json:"prizeStatus"`
		PrizeType    int    `json:"prizeType"`
		ShowPriority int    `json:"showPriority"`
		UseDate      string `json:"useDate"`
		UserId       int    `json:"userId"`
	}

	UserDrawPrizeResult struct {
		Success bool
		Tip string
	}

	ActivityTaskId string
)

const (
	ActivitySignin ActivityTaskId = "TASK_SIGNIN"
	ActivitySignPhotos ActivityTaskId = "TASK_SIGNIN_PHOTOS"
)

// 抽奖
func (p *PanClient) UserDrawPrize(taskId ActivityTaskId) (*UserDrawPrizeResult, *apierror.ApiError) {
	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "https://m.cloud.189.cn/v2/drawPrizeMarketDetails.action?taskId=%s&activityId=ACT_SIGNIN",
		taskId)
	logger.Verboseln("do request url: " + fullUrl.String())
	body, err := p.client.DoGet(fullUrl.String())
	if err != nil {
		return nil, apierror.NewApiErrorWithError(err)
	}
	logger.Verboseln("response: " + string(body))

	errResp := &apierror.ErrorResp{}
	if err := json.Unmarshal(body, errResp); err == nil {
		if errResp.ErrorCode != "" {
			if errResp.ErrorCode == "User_Not_Chance" {
				return nil, apierror.NewFailedApiError("今日已无抽奖机会")
			}
			return nil, apierror.NewFailedApiError(errResp.ErrorCode)
		}
	}

	item := &userDrawPrizeResp{}
	if err := json.Unmarshal(body, item); err != nil {
		logger.Verboseln("UserDrawPrize parse response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}

	result := UserDrawPrizeResult{}
	if item.PrizeName != "" {
		result.Success = true
		result.Tip = item.PrizeName
		return &result, nil
	}
	return nil, apierror.NewFailedApiError("抽奖失败")
}