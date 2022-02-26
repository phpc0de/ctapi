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
	"github.com/phpc0de/ctapi/cloudpan/apierror"
	"github.com/phpc0de/ctapi/cloudpan/apiutil"
	"github.com/phpc0de/ctlibgo/logger"
	"strings"
)

// AppFamilySaveFileToPersonCloud 复制家庭共享文件文件到个人云
func (p *PanClient) AppFamilySaveFileToPersonCloud(familyId int64, familyFileIdList []string) (bool, *apierror.ApiError) {
	if len(familyFileIdList) == 0 {
		return false, nil
	}

	fileIdStrList := []string{}
	for _,item := range familyFileIdList {
		fileIdStrList = append(fileIdStrList, "fileIdList=" + item)
	}
	fileIdListStr := strings.Join(fileIdStrList, "&")

	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "%s/family/file/saveFileToMember.action?familyId=%d&%s&destParentId=&%s",
		API_URL,
		familyId,
		fileIdListStr,
		apiutil.PcClientInfoSuffixParam())

	sessionKey := p.appToken.FamilySessionKey
	sessionSecret := p.appToken.FamilySessionSecret
	httpMethod := "GET"
	dateOfGmt := apiutil.DateOfGmtStr()
	headers := map[string]string {
		"Date": dateOfGmt,
		"SessionKey": sessionKey,
		"Signature": apiutil.SignatureOfHmac(sessionSecret, sessionKey, httpMethod, fullUrl.String(), dateOfGmt),
		"X-Request-ID": apiutil.XRequestId(),
	}

	logger.Verboseln("do request url: " + fullUrl.String())
	respBody, err1 := p.client.Fetch(httpMethod, fullUrl.String(), nil, headers)
	if err1 != nil {
		logger.Verboseln("AppSaveFileToPersonCloud occurs error: ", err1.Error())
		return false, apierror.NewApiErrorWithError(err1)
	}
	logger.Verboseln("response: " + string(respBody))

	er := &apierror.AppErrorXmlResp{}
	if err := xml.Unmarshal(respBody, er); err == nil {
		if er.Code != "" {
			if er.Code == "FileAlreadyExists" {
				return false, apierror.NewApiError(apierror.ApiCodeFileAlreadyExisted, "文件已存在")
			}
			return false, apierror.NewFailedApiError("复制保存文件到个人云出错")
		}
	}
	return true, nil
}


// AppSaveFileToFamilyCloud 复制个人云文件文件到家庭云
func (p *PanClient) AppSaveFileToFamilyCloud(familyId int64, personFileIdList []string) (bool, *apierror.ApiError) {
	if len(personFileIdList) == 0 {
		return false, nil
	}

	fileIdStrList := []string{}
	for _,item := range personFileIdList {
		fileIdStrList = append(fileIdStrList, "fileIdList=" + item)
	}
	fileIdListStr := strings.Join(fileIdStrList, "&")

	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "%s/family/file/shareFileToFamily.action?familyId=%d&%s&destParentId=&%s",
		API_URL,
		familyId,
		fileIdListStr,
		apiutil.PcClientInfoSuffixParam())

	sessionKey := p.appToken.FamilySessionKey
	sessionSecret := p.appToken.FamilySessionSecret
	httpMethod := "GET"
	dateOfGmt := apiutil.DateOfGmtStr()
	headers := map[string]string {
		"Date": dateOfGmt,
		"SessionKey": sessionKey,
		"Signature": apiutil.SignatureOfHmac(sessionSecret, sessionKey, httpMethod, fullUrl.String(), dateOfGmt),
		"X-Request-ID": apiutil.XRequestId(),
	}

	logger.Verboseln("do request url: " + fullUrl.String())
	respBody, err1 := p.client.Fetch(httpMethod, fullUrl.String(), nil, headers)
	if err1 != nil {
		logger.Verboseln("AppSaveFileToFamilyCloud occurs error: ", err1.Error())
		return false, apierror.NewApiErrorWithError(err1)
	}
	logger.Verboseln("response: " + string(respBody))

	er := &apierror.AppErrorXmlResp{}
	if err := xml.Unmarshal(respBody, er); err == nil {
		if er.Code != "" {
			if er.Code == "FileAlreadyExists" {
				return false, apierror.NewApiError(apierror.ApiCodeFileAlreadyExisted, "文件已存在")
			}
			return false, apierror.NewFailedApiError("复制保存文件到家庭云出错")
		}
	}
	return true, nil
}