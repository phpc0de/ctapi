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
	"github.com/phpc0de/ctapi/cloudpan/apiutil"
	"github.com/phpc0de/ctlibgo/logger"
	"net/url"
	"strings"
)

type (
	MkdirResult struct {
		// fileId 文件ID
		FileId string `json:"fileId"`
		// isNew 是否创建成功。true为成功，false或者没有返回则为失败，失败原因基本是已存在该文件夹
		IsNew bool `json:"isNew"`
	}
)

func (p *PanClient) Mkdir(parentFileId, dirName string) (*MkdirResult, *apierror.ApiError) {
	if parentFileId == "" {
		// 默认根目录
		parentFileId = "-11"
	}

	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "%s/v2/createFolder.action?parentId=%s&fileName=%s",
		WEB_URL, parentFileId, url.QueryEscape(dirName))
	logger.Verboseln("do request url: " + fullUrl.String())
	body, err := p.client.DoGet(fullUrl.String())
	if err != nil {
		logger.Verboseln("mkdir failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	item := &MkdirResult{}
	if err := json.Unmarshal(body, item); err != nil {
		logger.Verboseln("mkdir response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	if !item.IsNew {
		return item, apierror.NewFailedApiError("文件夹已存在: " + dirName)
	}
	return item, nil
}


func (p *PanClient) MkdirRecursive(parentFileId string, fullPath string, index int, pathSlice []string) (*MkdirResult, *apierror.ApiError) {
	r := &MkdirResult{}
	if parentFileId == "" {
		// default root "/" entity
		parentFileId = NewFileEntityForRootDir().FileId
		if index == 0 && len(pathSlice) == 1 {
			// root path "/"
			r.IsNew = false
			r.FileId = parentFileId
			return r, nil
		}

		fullPath = ""
		return p.MkdirRecursive(parentFileId, fullPath, index + 1, pathSlice)
	}

	if index >= len(pathSlice) {
		r.IsNew = false
		r.FileId = parentFileId
		return r, nil
	}

	listFilePath := NewFileListParam()
	listFilePath.FileId = parentFileId
	fileResult, err := p.FileList(listFilePath)
	if err != nil {
		r.IsNew = false
		r.FileId = ""
		return r, err
	}

	// existed?
	for _, fileEntity := range fileResult.Data {
		if fileEntity.FileName == pathSlice[index] {
			return p.MkdirRecursive(fileEntity.FileId, fullPath + "/" + pathSlice[index], index + 1, pathSlice)
		}
	}

	// not existed, mkdir dir
	name := pathSlice[index]
	if !apiutil.CheckFileNameValid(name) {
		r.IsNew = false
		r.FileId = ""
		return r, apierror.NewFailedApiError("文件夹名不能包含特殊字符：" + apiutil.FileNameSpecialChars)
	}

	rs, err := p.Mkdir(parentFileId, name)
	if err != nil {
		r.IsNew = false
		r.FileId = ""
		return r, err
	}

	if (index+1) >= len(pathSlice) {
		return rs, nil
	} else {
		return p.MkdirRecursive(rs.FileId, fullPath + "/" + pathSlice[index], index + 1, pathSlice)
	}
}