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
	"math"
	"net/url"
	"path"
	"strings"
	"time"
)

type (
	// HandleFileDirectoryFunc 处理文件或目录的元信息, 返回值控制是否退出递归
	HandleAppFileDirectoryFunc func(depth int, fdPath string, fd *AppFileEntity, apierr *apierror.ApiError) bool

	// AppGetFileInfoParam 获取文件信息参数
	AppGetFileInfoParam struct {
		// 家庭云ID
		FamilyId int64
		// FileId 文件ID，支持文件和文件夹
		FileId   string
		// FilePath 文件绝对路径，支持文件和文件夹
		FilePath string
	}

	// AppGetFileInfoResult 文件信息响应值
	AppGetFileInfoResult struct {
		//XMLName xml.Name `xml:"folderInfo"`
		FileId string `xml:"id"`
		ParentId string `xml:"parentFolderId"`
		FileName string `xml:"name"`
		CreateDate string `xml:"createDate"`
		LastOpTime string `xml:"lastOpTime"`
		Path string `xml:"path"`
		Rev string `xml:"rev"`
		ParentFolderList parentFolderListNode `xml:"parentFolderList"`
		GroupSpaceId string `xml:"groupSpaceId"`
	}
	parentFolderListNode struct {
		FolderList []appGetFolderInfoNode `xml:"folder"`
	}
	appGetFolderInfoNode struct {
		Fid string `xml:"fid"`
		Fname string `xml:"fname"`
	}

	AppOrderBy string

	// FileListParam 文件列表参数
	AppFileListParam struct {
		// 家庭云ID
		FamilyId int64
		// FileId 文件ID
		FileId string
		// OrderBy 排序字段
		OrderBy OrderBy
		// OrderSort 排序顺序
		OrderSort OrderSort
		// PageNum 页数量，从1开始
		PageNum uint
		// PageSize 页大小，默认60
		PageSize uint

		// 默认是不返回Path路径，是否构建
		ConstructPath bool
	}

	// AppFileListResult 文件列表响应值
	AppFileListResult struct {
		LastRev string
		// 总数量
		Count int
		// 文件列表
		FileList AppFileList
	}

	// AppFileEntity 文件/文件夹信息
	AppFileEntity struct {
		// FileId 文件ID
		FileId string `xml:"id"`
		// ParentId 父文件夹ID
		ParentId string `xml:"parentId"`
		// FileMd5 文件MD5，文件夹为空，空文件默认为 D41D8CD98F00B204E9800998ECF8427E
		FileMd5 string `xml:"md5"`
		// FileName 名称
		FileName string `xml:"name"`
		// FileSize 文件大小
		FileSize int64 `xml:"size"`
		// LastOpTime 最后修改时间
		LastOpTime string `xml:"lastOpTime"`
		// CreateTime 创建时间
		CreateTime string `xml:"createDate"`
		// 文件完整路径
		Path string `xml:"path"`
		// MediaType 媒体类型
		MediaType MediaType `xml:"mediaType"`
		// IsFolder 是否是文件夹
		IsFolder bool
		// FileCount 文件夹子文件数量，对文件夹详情有效
		SubFileCount uint `xml:"fileCount"`

		StartLabel int `xml:"startLabel"`
		FavoriteLabel int `xml:"favoriteLabel"`
		Orientation int `xml:"orientation"`
		Rev string `xml:"rev"`
		FileCata int `xml:"fileCata"`
	}
	AppFileList []*AppFileEntity
)

const (
	// AppOrderByName 文件名
	AppOrderByName AppOrderBy = "filename"
	// AppOrderBySize 大小
	AppOrderBySize AppOrderBy = "filesize"
	// AppOrderByTime 时间
	AppOrderByTime AppOrderBy = "lastOpTime"

	// 空文件夹默认MD5
	DefaultEmptyFileMd5 = "D41D8CD98F00B204E9800998ECF8427E"
)

// AppGetBasicFileInfo 根据文件ID或者文件绝对路径获取文件信息，支持文件和文件夹
func (p *PanClient) AppGetBasicFileInfo(param *AppGetFileInfoParam) (*AppGetFileInfoResult, *apierror.ApiError) {
	fullUrl := &strings.Builder{}

	sessionKey := ""
	sessionSecret := ""
	if param.FamilyId <= 0 {
		// 个人云
		fmt.Fprintf(fullUrl, "%s/getFolderInfo.action?folderId=%s&folderPath=%s&pathList=0&dt=3&%s",
			API_URL, param.FileId, url.QueryEscape(param.FilePath), apiutil.PcClientInfoSuffixParam())
		sessionKey = p.appToken.SessionKey
		sessionSecret = p.appToken.SessionSecret
	} else {
		// 家庭云
		if param.FileId == "" {
			return nil, apierror.NewFailedApiError("FileId为空")
		}
		fmt.Fprintf(fullUrl, "%s/family/file/getFolderInfo.action?familyId=%d&folderId=%s&folderPath=%s&pathList=0&%s",
			API_URL, param.FamilyId, param.FileId, url.QueryEscape(param.FilePath), apiutil.PcClientInfoSuffixParam())
		sessionKey = p.appToken.FamilySessionKey
		sessionSecret = p.appToken.FamilySessionSecret
	}
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
		logger.Verboseln("AppGetBasicFileInfo occurs error: ", err1.Error())
		return nil, apierror.NewApiErrorWithError(err1)
	}
	logger.Verboseln("response: " + string(respBody))
	er := &apierror.AppErrorXmlResp{}
	if err := xml.Unmarshal(respBody, er); err == nil {
		if er.Code != "" {
			if er.Code == "FileNotFound" {
				return nil, apierror.NewApiError(apierror.ApiCodeFileNotFoundCode, "文件不存在")
			}
			return nil, apierror.NewFailedApiError("请求出错")
		}
	}
	item := &AppGetFileInfoResult{}
	if param.FamilyId <= 0 {
		if err := xml.Unmarshal(respBody, item); err != nil {
			logger.Verboseln("AppGetBasicFileInfo parse response failed")
			return nil, apierror.NewApiErrorWithError(err)
		}
	} else {
		type familyAppGetFileInfoResult struct {
			FileId string `xml:"id"`
			ParentId string `xml:"parentId"`
			FileName string `xml:"name"`
			CreateDate string `xml:"createDate"`
			LastOpTime string `xml:"lastOpTime"`
			Path string `xml:"path"`
			Rev string `xml:"rev"`
		}
		fitem := &familyAppGetFileInfoResult{}
		if err := xml.Unmarshal(respBody, fitem); err != nil {
			logger.Verboseln("AppGetBasicFileInfo parse response failed")
			return nil, apierror.NewApiErrorWithError(err)
		}
		item = &AppGetFileInfoResult{
			FileId: fitem.FileId,
			ParentId: fitem.ParentId,
			FileName: fitem.FileName,
			CreateDate: fitem.CreateDate,
			LastOpTime: fitem.LastOpTime,
			Rev: fitem.Rev,
		}
	}

	return item, nil
}

func getAppOrderBy(by OrderBy) AppOrderBy {
	switch by {
	case OrderByName:
		return AppOrderByName
	case OrderBySize:
		return AppOrderBySize
	case OrderByTime:
		return AppOrderByTime
	default:
		return AppOrderByName
	}
}

func NewAppFileListParam() *AppFileListParam {
	return &AppFileListParam {
		FamilyId: 0,
		FileId: "-11",
		OrderBy: OrderByName,
		OrderSort: OrderAsc,
		PageNum: 1,
		PageSize: 200,
	}
}

func NewAppFileEntityForRootDir() *AppFileEntity {
	return &AppFileEntity {
		FileId: "-11",
		IsFolder: true,
		FileName: "/",
		ParentId: "",
	}
}


// TotalSize 获取目录下文件的总大小
func (afl AppFileList) TotalSize() int64 {
	var size int64
	for k := range afl {
		if afl[k] == nil {
			continue
		}

		size += afl[k].FileSize
	}
	return size
}


// Count 获取文件总数和目录总数
func (afl AppFileList) Count() (fileN, directoryN int64) {
	for k := range afl {
		if afl[k] == nil {
			continue
		}

		if afl[k].IsFolder {
			directoryN++
		} else {
			fileN++
		}
	}
	return
}

func (f *AppFileEntity) String() string {
	builder := &strings.Builder{}
	builder.WriteString("文件ID: " + f.FileId + "\n")
	builder.WriteString("文件名: " + f.FileName + "\n")
	if f.IsFolder {
		builder.WriteString("文件类型: 目录\n")
	} else {
		builder.WriteString("文件类型: 文件\n")
	}
	builder.WriteString("文件路径: " + f.Path + "\n")
	return builder.String()
}

func (f *AppFileEntity) CreateFileEntity() *FileEntity {
	return &FileEntity{
		FileId: f.FileId,
		ParentId: f.ParentId,
		FileIdDigest: f.FileMd5,
		FileName: f.FileName,
		FileSize: f.FileSize,
		LastOpTime: f.LastOpTime,
		CreateTime: f.CreateTime,
		Path: f.Path,
		MediaType: f.MediaType,
		IsFolder: f.IsFolder,
		SubFileCount: f.SubFileCount,
	}
}

// AppGetAllFileList 获取指定目录下的所有文件列表
func (p *PanClient) AppGetAllFileList(param *AppFileListParam) (*AppFileListResult, *apierror.ApiError)  {
	internalParam := &AppFileListParam{
		FamilyId: param.FamilyId,
		FileId: param.FileId,
		OrderBy: param.OrderBy,
		OrderSort: param.OrderSort,
		PageNum: 1,
		PageSize: param.PageSize,
	}
	if internalParam.PageSize <= 0 {
		internalParam.PageSize = 200
	}
	if internalParam.FamilyId > 0 {
		if internalParam.FileId == "-11" {
			internalParam.FileId = ""
		}
	}

	result := &AppFileListResult{}
	fileResult, err := p.AppFileList(internalParam)
	if err != nil {
		return nil, err
	}
	result.LastRev = fileResult.LastRev
	result.FileList = fileResult.FileList
	result.Count = fileResult.Count

	// more page?
	if fileResult.Count > int(internalParam.PageSize) {
		pageCount := int(math.Ceil(float64(fileResult.Count) / float64(internalParam.PageSize)))
		for page := 2; page <= pageCount; page++ {
			internalParam.PageNum = uint(page)
			fileResult, err = p.AppFileList(internalParam)
			if err != nil {
				logger.Verboseln(err)
				break
			}
			result.FileList = append(result.FileList, fileResult.FileList...)
		}
	}

	// parentId
	for _,fi := range result.FileList {
		fi.ParentId = param.FileId
	}

	// construct path
	if param.ConstructPath {
		parentFullPath,err := p.AppFilePathById(param.FamilyId, param.FileId)
		if err == nil {
			for _,fi := range result.FileList {
				fi.Path = parentFullPath + "/" + fi.FileName
				fi.ParentId = param.FileId
			}
		}
	}

	return result, nil
}

// AppFileList 获取文件列表
func (p *PanClient) AppFileList(param *AppFileListParam) (*AppFileListResult, *apierror.ApiError) {
	fullUrl := &strings.Builder{}

	sessionKey := ""
	sessionSecret := ""
	if param.FamilyId <= 0 {
		// 个人云
		fmt.Fprintf(fullUrl, "%s/listFiles.action?folderId=%s&recursive=0&fileType=0&iconOption=10&mediaAttr=0&orderBy=%s&descending=%t&pageNum=%d&pageSize=%d&%s",
			API_URL,
			param.FileId, getAppOrderBy(param.OrderBy), param.OrderSort == OrderDesc, param.PageNum, param.PageSize,
			apiutil.PcClientInfoSuffixParam())
		sessionKey = p.appToken.SessionKey
		sessionSecret = p.appToken.SessionSecret
	} else {
		// 家庭云
		if param.FileId == "-11" {
			param.FileId = ""
		}
		fmt.Fprintf(fullUrl, "%s/family/file/listFiles.action?folderId=%s&familyId=%d&fileType=0&iconOption=0&mediaAttr=0&orderBy=%d&descending=%t&pageNum=%d&pageSize=%d&%s",
			API_URL,
			param.FileId, param.FamilyId, param.OrderBy, param.OrderSort == OrderDesc, param.PageNum, param.PageSize,
			apiutil.PcClientInfoSuffixParam())
		sessionKey = p.appToken.FamilySessionKey
		sessionSecret = p.appToken.FamilySessionSecret
	}
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
		logger.Verboseln("AppFileList occurs error: ", err1.Error())
		return nil, apierror.NewApiErrorWithError(err1)
	}
	logger.Verboseln("response: " + string(respBody))

	// handler common error
	if apiErr := apierror.ParseAppCommonApiError(respBody); apiErr != nil {
		return nil, apiErr
	}

	er := &apierror.AppErrorXmlResp{}
	if err := xml.Unmarshal(respBody, er); err == nil {
		if er.Code != "" {
			if er.Code == "FileNotFound" {
				return nil, apierror.NewApiError(apierror.ApiCodeFileNotFoundCode, "文件不存在")
			}
		}
	}

	type appFileListResultInternal struct {
		//XMLName xml.Name `xml:"listFiles"`
		LastRev string `xml:"lastRev"`
		// 总数量
		Count int `xml:"fileList>count"`
		// 文件夹列表
		FolderList AppFileList `xml:"fileList>folder"`
		// 文件列表
		FileList AppFileList `xml:"fileList>file"`
	}
	itemResult := &appFileListResultInternal{}
	if err := xml.Unmarshal(respBody, itemResult); err != nil {
		logger.Verboseln("AppFileList parse response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}

	result := &AppFileListResult{
		LastRev: itemResult.LastRev,
		Count: itemResult.Count,
		FileList: AppFileList{},
	}

	if itemResult.FolderList != nil && len(itemResult.FolderList) > 0 {
		for _,item := range itemResult.FolderList {
			item.IsFolder = true

			result.FileList = append(result.FileList, item)
		}
	}
	if itemResult.FileList != nil && len(itemResult.FileList) > 0 {
		for _,item := range itemResult.FileList {
			item.IsFolder = false

			result.FileList = append(result.FileList, item)
		}
	}

	return result, nil
}

// AppGetFilePathById 通过FileId获取文件的绝对路径
func (p *PanClient) AppFilePathById(familyId int64, fileId string) (string, *apierror.ApiError) {
	param := &AppGetFileInfoParam{
		FamilyId: familyId,
		FileId: fileId,
	}

	fullPath := ""
	for {
		fi,err := p.AppGetBasicFileInfo(param)
		if err != nil {
			return "", err
		}

		// 个人云支持
		if fi.Path != "" {
			return fi.Path, nil
		}

		if strings.Index(fi.FileId, "-") == 0 || strings.Index(fi.ParentId, "-") == 0 {
			// root file id
			fullPath = "/" + fullPath
			break
		}
		if fullPath == "" {
			fullPath = fi.FileName
		} else {
			fullPath = fi.FileName + "/" + fullPath
		}

		// next loop
		param.FileId = fi.ParentId
		time.Sleep(time.Duration(100) * time.Millisecond)
	}
	return fullPath, nil
}

// AppFileInfoById 通过FileId获取文件详情
func (p *PanClient) AppFileInfoById(familyId int64, fileId string) (fileInfo *AppFileEntity, error *apierror.ApiError) {
	basicFileInfo, err := p.AppGetBasicFileInfo(&AppGetFileInfoParam{FamilyId: familyId, FileId: fileId})
	if err != nil {
		return nil, err
	}

	param := NewAppFileListParam()
	param.FamilyId = familyId
	param.FileId = basicFileInfo.ParentId
	allFileInfo, err1 := p.AppGetAllFileList(param)
	if err1 != nil {
		return nil, err1
	}

	for _,item := range allFileInfo.FileList {
		if item.FileId == fileId {
			return item, nil
		}
	}
	return nil, nil
}

// AppFileInfoByPath 通过路径获取文件详情，pathStr是绝对路径
func (p *PanClient) AppFileInfoByPath(familyId int64, pathStr string) (fileInfo *AppFileEntity, error *apierror.ApiError) {
	if pathStr == "" {
		pathStr = "/"
	}
	//pathStr = path.Clean(pathStr)
	if !path.IsAbs(pathStr) {
		return nil, apierror.NewFailedApiError("pathStr必须是绝对路径")
	}
	if len(pathStr) > 1 {
		pathStr = path.Clean(pathStr)
	}

	var pathSlice []string
	if pathStr == "/" {
		pathSlice = []string{""}
	} else {
		pathSlice = strings.Split(pathStr, PathSeparator)
		if pathSlice[0] != "" {
			return nil, apierror.NewFailedApiError("pathStr必须是绝对路径")
		}
	}
	return p.getAppFileInfoByPath(familyId, 0, &pathSlice, nil)
}

func (p *PanClient) getAppFileInfoByPath(familyId int64, index int, pathSlice *[]string, parentFileInfo *AppFileEntity) (*AppFileEntity, *apierror.ApiError)  {
	if parentFileInfo == nil {
		// default root "/" entity
		parentFileInfo = NewAppFileEntityForRootDir()
		if index == 0 && len(*pathSlice) == 1 {
			// root path "/"
			return parentFileInfo, nil
		}

		return p.getAppFileInfoByPath(familyId, index + 1, pathSlice, parentFileInfo)
	}

	if index >= len(*pathSlice) {
		return parentFileInfo, nil
	}

	searchPath := NewAppFileListParam()
	searchPath.FileId = parentFileInfo.FileId
	searchPath.FamilyId = familyId
	fileResult, err := p.AppGetAllFileList(searchPath)
	if err != nil {
		return nil, err
	}

	if fileResult == nil || fileResult.FileList == nil || len(fileResult.FileList) == 0  {
		return nil, apierror.NewApiError(apierror.ApiCodeFileNotFoundCode, "文件不存在")
	}
	for _, fileEntity := range fileResult.FileList {
		if fileEntity.FileName == (*pathSlice)[index] {
			fileEntity.ParentId = parentFileInfo.FileId
			fileEntity.Path = getPath(index, pathSlice)
			return p.getAppFileInfoByPath(familyId, index + 1, pathSlice, fileEntity)
		}
	}
	return nil, apierror.NewApiError(apierror.ApiCodeFileNotFoundCode, "文件不存在")
}

func getPath(index int, pathSlice *[]string) string {
	fullPath := ""
	for idx, str := range *pathSlice {
		if idx > index {
			break
		}
		fullPath += "/" + str
	}
	return strings.ReplaceAll(fullPath, "//", "/")
}



// FilesDirectoriesRecurseList 递归获取目录下的文件和目录列表
func (p *PanClient) AppFilesDirectoriesRecurseList(familyId int64, path string, handleAppFileDirectoryFunc HandleAppFileDirectoryFunc) AppFileList {
	targetFileInfo, er := p.AppFileInfoByPath(familyId, path)
	if er != nil {
		if handleAppFileDirectoryFunc != nil {
			handleAppFileDirectoryFunc(0, path, nil, er)
		}
		return nil
	}
	if !targetFileInfo.IsFolder {
		if handleAppFileDirectoryFunc != nil {
			handleAppFileDirectoryFunc(0, path, targetFileInfo, nil)
		}
		return AppFileList{targetFileInfo}
	}

	fld := &AppFileList{}
	ok := p.appRecurseList(familyId, targetFileInfo, 1, handleAppFileDirectoryFunc, fld)
	if !ok {
		return nil
	}
	return *fld
}

func (p *PanClient) appRecurseList(familyId int64, folderInfo *AppFileEntity, depth int, handleAppFileDirectoryFunc HandleAppFileDirectoryFunc, fld *AppFileList) bool {
	flp := NewAppFileListParam()
	flp.FileId = folderInfo.FileId
	flp.FamilyId = familyId
	flp.ConstructPath = true
	r, apiError := p.AppGetAllFileList(flp)
	if apiError != nil {
		if handleAppFileDirectoryFunc != nil {
			handleAppFileDirectoryFunc(depth, folderInfo.Path, nil, apiError) // 传递错误
		}
		return false
	}
	ok := true
	for _, fi := range r.FileList {
		*fld = append(*fld, fi)
		if fi.IsFolder {
			time.Sleep(time.Duration(200) * time.Millisecond)
			ok = p.appRecurseList(familyId, fi, depth+1, handleAppFileDirectoryFunc, fld)
		} else {
			if handleAppFileDirectoryFunc != nil {
				ok = handleAppFileDirectoryFunc(depth, fi.Path, fi, nil)
			}
		}
		if !ok {
			return false
		}
	}
	return true
}