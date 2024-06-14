// Copyright © 2016- 2024 Sesame Network Technology all right reserved

package manage

import (
	"chatwiki/internal/app/chatwiki/common"
	"chatwiki/internal/app/chatwiki/define"
	"chatwiki/internal/app/chatwiki/i18n"
	"chatwiki/internal/pkg/lib_redis"
	"chatwiki/internal/pkg/lib_web"
	"errors"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/tmc/langchaingo/textsplitter"
	"github.com/zhimaAi/go_tools/logs"
	"github.com/zhimaAi/go_tools/msql"
	"github.com/zhimaAi/go_tools/tool"
)

func GetLibFileList(c *gin.Context) {
	var userId int
	if userId = GetAdminUserId(c); userId == 0 {
		return
	}
	libraryId := cast.ToInt(c.Query(`library_id`))
	if libraryId <= 0 {
		c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `param_lack`))))
		return
	}
	info, err := common.GetLibraryInfo(libraryId, userId)
	if err != nil {
		logs.Error(err.Error())
		c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `sys_err`))))
		return
	}
	if len(info) == 0 {
		c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `no_data`))))
		return
	}
	page := max(1, cast.ToInt(c.Query(`page`)))
	size := max(1, cast.ToInt(c.Query(`size`)))
	m := msql.Model(`chat_ai_library_file`, define.Postgres)
	m.Where(`admin_user_id`, cast.ToString(userId)).Where(`library_id`, cast.ToString(libraryId))
	fileName := strings.TrimSpace(c.Query(`file_name`))
	if len(fileName) > 0 {
		m.Where(`file_name`, `like`, fileName)
	}
	list, total, err := m.Field(`id,file_name,status,errmsg,file_ext,file_size,file_url,pdf_url,create_time`).
		Order(`id desc`).Paginate(page, size)
	if err != nil {
		logs.Error(err.Error())
		c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `sys_err`))))
		return
	}
	data := map[string]any{`info`: info, `list`: list, `total`: total, `page`: page, `size`: size}
	c.String(http.StatusOK, lib_web.FmtJson(data, nil))
}

func AddLibraryFile(c *gin.Context) {
	var userId int
	if userId = GetAdminUserId(c); userId == 0 {
		return
	}
	libraryId := cast.ToInt(c.PostForm(`library_id`))
	if libraryId <= 0 {
		c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `param_lack`))))
		return
	}
	info, err := common.GetLibraryInfo(libraryId, userId)
	if err != nil {
		logs.Error(err.Error())
		c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `sys_err`))))
		return
	}
	if len(info) == 0 {
		c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `no_data`))))
		return
	}
	//document uploaded
	libraryFiles, errs := common.SaveUploadedFileMulti(c, `library_files`, define.LibFileLimitSize, userId, `library_file`, define.LibFileAllowExt)
	if len(libraryFiles) == 0 {
		c.String(http.StatusOK, lib_web.FmtJson(errs, errors.New(i18n.Show(common.GetLang(c), `upload_empty`))))
		return
	}
	//database dispose
	fileIds := make([]int64, 0)
	m := msql.Model(`chat_ai_library_file`, define.Postgres)
	for _, uploadInfo := range libraryFiles {
		status := define.FileStatusInitial
		isTableFile := define.IsTableFile(uploadInfo.Ext)
		if isTableFile {
			status = define.FileStatusWaitSplit
		}
		fileId, err := m.Insert(msql.Datas{
			`admin_user_id`: userId,
			`library_id`:    libraryId,
			`file_url`:      uploadInfo.Link,
			`file_name`:     uploadInfo.Name,
			`status`:        status,
			`file_ext`:      uploadInfo.Ext,
			`file_size`:     uploadInfo.Size,
			`create_time`:   tool.Time2Int(),
			`update_time`:   tool.Time2Int(),
			`is_table_file`: cast.ToInt(isTableFile),
		}, `id`)
		//clear cached data
		lib_redis.DelCacheData(define.Redis, &common.LibFileCacheBuildHandler{FileId: int(fileId)})
		if err != nil {
			logs.Error(err.Error())
		} else {
			fileIds = append(fileIds, fileId)
			if !isTableFile { //async task:convert pdf
				if message, err := tool.JsonEncode(map[string]any{`file_id`: fileId, `file_url`: uploadInfo.Link}); err != nil {
					logs.Error(err.Error())
				} else if err := common.AddJobs(define.ConvertPdfTopic, message); err != nil {
					logs.Error(err.Error())
				}
			}
		}
	}
	c.String(http.StatusOK, lib_web.FmtJson(map[string]any{`file_ids`: fileIds}, nil))
}

func DelLibraryFile(c *gin.Context) {
	var userId int
	if userId = GetAdminUserId(c); userId == 0 {
		return
	}
	id := cast.ToInt(c.PostForm(`id`))
	if id <= 0 {
		c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `param_lack`))))
		return
	}
	info, err := common.GetLibFileInfo(id, userId)
	if err != nil {
		logs.Error(err.Error())
		c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `sys_err`))))
		return
	}
	if len(info) == 0 {
		c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `no_data`))))
		return
	}
	_, err = msql.Model(`chat_ai_library_file`, define.Postgres).Where(`id`, cast.ToString(id)).Delete()
	if err != nil {
		logs.Error(err.Error())
		c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `sys_err`))))
		return
	}
	//clear cached data
	lib_redis.DelCacheData(define.Redis, &common.LibFileCacheBuildHandler{FileId: id})
	//dispose relation data
	_, err = msql.Model(`chat_ai_library_file_data`, define.Postgres).Where(`file_id`, cast.ToString(id)).Delete()
	if err != nil {
		logs.Error(err.Error())
	}
	_, err = msql.Model(`chat_ai_library_file_data_index`, define.Postgres).Where(`file_id`, cast.ToString(id)).Delete()
	if err != nil {
		logs.Error(err.Error())
	}
	c.String(http.StatusOK, lib_web.FmtJson(nil, nil))
}

func GetLibFileInfo(c *gin.Context) {
	var userId int
	if userId = GetAdminUserId(c); userId == 0 {
		return
	}
	id := cast.ToInt(c.Query(`id`))
	if id <= 0 {
		c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `param_lack`))))
		return
	}
	info, err := common.GetLibFileInfo(id, userId)
	if err != nil {
		logs.Error(err.Error())
		c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `sys_err`))))
		return
	}
	if len(info) == 0 {
		c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `no_data`))))
		return
	}

	var separators []string
	for _, noStr := range strings.Split(info[`separators_no`], `,`) {
		if len(noStr) == 0 {
			continue
		}
		no := cast.ToInt(noStr)
		separators = append(separators, cast.ToString(define.SeparatorsList[no-1][`name`]))
	}
	info[`separators`] = strings.Join(separators, ", ")

	c.String(http.StatusOK, lib_web.FmtJson(info, nil))
}

func GetSeparatorsList(c *gin.Context) {
	var userId int
	if userId = GetAdminUserId(c); userId == 0 {
		return
	}
	list := make([]map[string]any, 0)
	for _, item := range define.SeparatorsList {
		name := i18n.Show(common.GetLang(c), cast.ToString(item[`name`]))
		list = append(list, map[string]any{`no`: item[`no`], `name`: name})
	}
	c.String(http.StatusOK, lib_web.FmtJson(list, nil))
}

func GetLibFileSplit(c *gin.Context) {
	var userId int
	if userId = GetAdminUserId(c); userId == 0 {
		return
	}
	id := cast.ToInt(c.Query(`id`))
	if id <= 0 {
		c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `param_lack`))))
		return
	}
	info, err := common.GetLibFileInfo(id, userId)
	if err != nil {
		logs.Error(err.Error())
		c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `sys_err`))))
		return
	}
	if len(info) == 0 {
		c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `no_data`))))
		return
	}
	if cast.ToInt(info[`status`]) != define.FileStatusWaitSplit {
		c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `status_exception`))))
		return
	}
	//check params
	splitParams, err := common.CheckSplitParams(c, cast.ToInt(info[`is_table_file`]))
	if err != nil {
		c.String(http.StatusOK, lib_web.FmtJson(nil, err))
		return
	}
	//read document content
	var list []define.DocSplitItem
	var wordTotal = 0

	if cast.ToInt(info[`is_table_file`]) == define.FileIsTable && splitParams.IsQaDoc == define.DocTypeQa {
		list, wordTotal, err = common.ReadQaTab(info[`file_url`], info[`file_ext`], splitParams)
	} else if cast.ToInt(info[`is_table_file`]) == define.FileIsTable && splitParams.IsQaDoc != define.DocTypeQa {
		list, wordTotal, err = common.ReadTab(info[`file_url`], info[`file_ext`])
	} else {
		switch strings.ToLower(info[`file_ext`]) {
		case `docx`:
			list, wordTotal, err = common.ReadDocx(info[`file_url`])
		case `txt`, `md`:
			list, wordTotal, err = common.ReadTxt(info[`file_url`], false)
		case `html`:
			list, wordTotal, err = common.ReadTxt(info[`file_url`], true)
		default:
			list, wordTotal, err = common.ReadPdf(info[`pdf_url`])
		}
	}

	if err != nil {
		logs.Error(err.Error())
		c.String(http.StatusOK, lib_web.FmtJson(nil, err))
		return
	}
	if len(list) == 0 || wordTotal == 0 {
		c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `doc_empty`))))
		return
	}
	//initialize RecursiveCharacter
	split := textsplitter.NewRecursiveCharacter()
	if splitParams.IsDiySplit == define.SplitTypeDiy {
		split.Separators = append(splitParams.Separators, split.Separators...)
		split.ChunkSize = splitParams.ChunkSize
		split.ChunkOverlap = splitParams.ChunkOverlap
	}
	// split by document type
	if splitParams.IsQaDoc == define.DocTypeQa {
		if cast.ToInt(info[`is_table_file`]) != define.FileIsTable {
			list = common.QaDocSplit(splitParams, list)
		}
	} else {
		list = common.MultDocSplit(split, list)
	}

	for i := range list {
		list[i].Number = i + 1 //serial number
		if splitParams.IsQaDoc == define.DocTypeQa {
			list[i].WordTotal = utf8.RuneCountInString(list[i].Question) + utf8.RuneCountInString(list[i].Answer)
		} else {
			list[i].WordTotal = utf8.RuneCountInString(list[i].Content)
		}
	}
	data := map[string]any{`split_params`: splitParams, `list`: list, `word_total`: wordTotal}
	c.String(http.StatusOK, lib_web.FmtJson(data, nil))
}

func SaveLibFileSplit(c *gin.Context) {
	var userId int
	if userId = GetAdminUserId(c); userId == 0 {
		return
	}
	fileId := cast.ToInt(c.PostForm(`id`))
	wordTotal := cast.ToInt(c.PostForm(`word_total`))
	splitParams, list := define.SplitParams{}, make([]define.DocSplitItem, 0)
	qaIndexType := cast.ToInt(c.PostForm(`qa_index_type`))
	if err := tool.JsonDecodeUseNumber(c.PostForm(`split_params`), &splitParams); err != nil {
		c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `param_invalid`, `split_params`))))
		return
	}
	if err := tool.JsonDecodeUseNumber(c.PostForm(`list`), &list); err != nil {
		c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `param_invalid`, `list`))))
		return
	}
	if fileId <= 0 || wordTotal <= 0 || len(list) == 0 {
		c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `param_lack`))))
		return
	}
	info, err := common.GetLibFileInfo(fileId, userId)
	if err != nil {
		logs.Error(err.Error())
		c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `sys_err`))))
		return
	}
	if len(info) == 0 {
		c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `no_data`))))
		return
	}
	if cast.ToInt(info[`status`]) != define.FileStatusWaitSplit {
		c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `status_exception`))))
		return
	}
	//check params
	if splitParams.IsQaDoc == define.DocTypeQa { // qa
		for i := range list {
			list[i].Number = i + 1 //serial number
			list[i].WordTotal = utf8.RuneCountInString(list[i].Question + list[i].Answer)
			if utf8.RuneCountInString(list[i].Question) < 1 || utf8.RuneCountInString(list[i].Question) > define.MaxContent {
				c.String(http.StatusOK, lib_web.FmtJson(map[string]int{`index`: i + 1}, errors.New(i18n.Show(common.GetLang(c), `length_err`, i+1))))
				return
			}
			if utf8.RuneCountInString(list[i].Answer) < 1 || utf8.RuneCountInString(list[i].Answer) > define.MaxContent {
				c.String(http.StatusOK, lib_web.FmtJson(map[string]int{`index`: i + 1}, errors.New(i18n.Show(common.GetLang(c), `length_err`, i+1))))
				return
			}
		}
	} else {
		for i := range list {
			list[i].Number = i + 1 //serial number
			list[i].WordTotal = utf8.RuneCountInString(list[i].Content)
			if list[i].WordTotal < 1 || list[i].WordTotal > define.MaxContent {
				c.String(http.StatusOK, lib_web.FmtJson(map[string]int{`index`: i + 1}, errors.New(i18n.Show(common.GetLang(c), `length_err`, i+1))))
				return
			}
		}
	}

	if splitParams.IsQaDoc == define.DocTypeQa {
		if qaIndexType != define.QAIndexTypeQuestionAndAnswer && qaIndexType != define.QAIndexTypeQuestion {
			c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `param_invalid`, `qa_index_type`))))
			return
		}
	}

	//add lock dispose
	if !lib_redis.AddLock(define.Redis, define.LockPreKey+`SaveLibFileSplit`+cast.ToString(fileId), time.Minute*5) {
		c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `op_lock`))))
		return
	}
	//database dispose
	m := msql.Model(`chat_ai_library_file`, define.Postgres)
	err = m.Begin()
	if err != nil {
		c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `sys_err`))))
		lib_redis.UnLock(define.Redis, define.LockPreKey+`SaveLibFileSplit`+cast.ToString(fileId))
		return
	}

	data := msql.Datas{
		`status`:          define.FileStatusLearning,
		`errmsg`:          `success`,
		`word_total`:      wordTotal,
		`split_total`:     len(list),
		`is_qa_doc`:       splitParams.IsQaDoc,
		`is_diy_split`:    splitParams.IsDiySplit,
		`separators_no`:   splitParams.SeparatorsNo,
		`chunk_size`:      splitParams.ChunkSize,
		`chunk_overlap`:   splitParams.ChunkOverlap,
		`question_lable`:  splitParams.QuestionLable,
		`answer_lable`:    splitParams.AnswerLable,
		`question_column`: splitParams.QuestionColumn,
		`answer_column`:   splitParams.AnswerColumn,
		`qa_index_type`:   qaIndexType,
		`update_time`:     tool.Time2Int(),
	}
	if qaIndexType != 0 {
		data[`qa_index_type`] = qaIndexType
	}

	_, err = m.Where(`id`, cast.ToString(fileId)).Update(data)
	if err != nil {
		logs.Error(err.Error())
	}
	//clear cached data
	lib_redis.DelCacheData(define.Redis, &common.LibFileCacheBuildHandler{FileId: fileId})

	//database dispose
	vm := msql.Model("chat_ai_library_file_data", define.Postgres)
	var indexIds []int64
	for i, item := range list {
		if utf8.RuneCountInString(item.Content) > define.MaxContent || utf8.RuneCountInString(item.Question) > define.MaxContent || utf8.RuneCountInString(item.Answer) > define.MaxContent {
			_ = m.Rollback()
			c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `length_err`, i+1))))
			lib_redis.UnLock(define.Redis, define.LockPreKey+`SaveLibFileSplit`+cast.ToString(fileId))
			return
		}

		data := msql.Datas{
			`admin_user_id`: info[`admin_user_id`],
			`library_id`:    info[`library_id`],
			`file_id`:       fileId,
			`number`:        item.Number,
			`page_num`:      item.PageNum,
			`title`:         item.Title,
			`word_total`:    item.WordTotal,
			`create_time`:   tool.Time2Int(),
			`update_time`:   tool.Time2Int(),
		}
		if splitParams.IsQaDoc == define.DocTypeQa {
			if splitParams.IsTableFile == define.FileIsTable {
				data[`type`] = define.ParagraphTypeExcelQA
			} else {
				data[`type`] = define.ParagraphTypeDocQA
			}
			data[`question`] = strings.TrimSpace(item.Question)
			data[`answer`] = strings.TrimSpace(item.Answer)
			id, err := vm.Insert(data, `id`)
			if err != nil {
				logs.Error(err.Error())
				_ = m.Rollback()
				c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `sys_err`))))
				lib_redis.UnLock(define.Redis, define.LockPreKey+`SaveLibFileSplit`+cast.ToString(fileId))
				return
			}
			vectorID, err := common.SaveVector(
				cast.ToInt64(info[`admin_user_id`]),
				cast.ToInt64(info[`library_id`]),
				cast.ToInt64(fileId),
				id,
				cast.ToString(define.VectorTypeQuestion),
				strings.TrimSpace(item.Question),
			)
			if err != nil {
				logs.Error(err.Error())
				_ = m.Rollback()
				c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `sys_err`))))
				lib_redis.UnLock(define.Redis, define.LockPreKey+`SaveLibFileSplit`+cast.ToString(fileId))
				return
			}
			indexIds = append(indexIds, vectorID)
			if qaIndexType == define.QAIndexTypeQuestionAndAnswer {
				vectorID, err = common.SaveVector(
					cast.ToInt64(info[`admin_user_id`]),
					cast.ToInt64(info[`library_id`]),
					cast.ToInt64(fileId),
					id,
					cast.ToString(define.VectorTypeAnswer),
					strings.TrimSpace(item.Answer),
				)
				if err != nil {
					logs.Error(err.Error())
					_ = m.Rollback()
					c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `sys_err`))))
					lib_redis.UnLock(define.Redis, define.LockPreKey+`SaveLibFileSplit`+cast.ToString(fileId))
					return
				}
				indexIds = append(indexIds, vectorID)
			}
		} else {
			data[`type`] = define.ParagraphTypeNormal
			if splitParams.IsQaDoc == define.DocTypeQa {
				data[`type`] = define.ParagraphTypeDocQA
			}
			data[`content`] = strings.TrimSpace(item.Content)
			id, err := vm.Insert(data, `id`)
			if err != nil {
				logs.Error(err.Error())
				_ = m.Rollback()
				c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `sys_err`))))
				lib_redis.UnLock(define.Redis, define.LockPreKey+`SaveLibFileSplit`+cast.ToString(fileId))
				return
			}
			vectorID, err := common.SaveVector(
				cast.ToInt64(info[`admin_user_id`]),
				cast.ToInt64(info[`library_id`]),
				cast.ToInt64(fileId),
				id,
				cast.ToString(define.VectorTypeParagraph),
				strings.TrimSpace(item.Content),
			)
			if err != nil {
				logs.Error(err.Error())
				_ = m.Rollback()
				c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `sys_err`))))
				lib_redis.UnLock(define.Redis, define.LockPreKey+`SaveLibFileSplit`+cast.ToString(fileId))
				return
			}
			indexIds = append(indexIds, vectorID)
		}
	}
	err = m.Commit()
	if err != nil {
		logs.Error(err.Error())
		c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `sys_err`))))
		lib_redis.UnLock(define.Redis, define.LockPreKey+`SaveLibFileSplit`+cast.ToString(fileId))
		return
	}

	//async task:convert vector
	for _, id := range indexIds {
		if message, err := tool.JsonEncode(map[string]any{`id`: id, `file_id`: fileId}); err != nil {
			logs.Error(err.Error())
		} else if err := common.AddJobs(define.ConvertVectorTopic, message); err != nil {
			logs.Error(err.Error())
		}
	}

	//unlock dispose
	lib_redis.UnLock(define.Redis, define.LockPreKey+`SaveLibFileSplit`+cast.ToString(fileId))
	c.String(http.StatusOK, lib_web.FmtJson(nil, nil))
}

func GetLibFileExcelTitle(c *gin.Context) {
	var userId int
	if userId = GetAdminUserId(c); userId == 0 {
		return
	}
	id := cast.ToInt(c.Query(`id`))
	if id <= 0 {
		c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `param_lack`))))
		return
	}
	info, err := common.GetLibFileInfo(id, userId)
	if err != nil {
		logs.Error(err.Error())
		c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `sys_err`))))
		return
	}
	if len(info) == 0 {
		c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `no_data`))))
		return
	}
	if info[`is_table_file`] != cast.ToString(define.FileIsTable) {
		c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `is_not_excel`))))
		return
	}
	rows, err := common.ParseTabFile(info[`file_url`], info[`file_ext`])
	if err != nil {
		c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `sys_err`))))
		return
	}
	if len(rows) < 2 {
		c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `excel_less_row`))))
		return
	}

	var data = make(map[string]string)
	for i, v := range rows[0] {
		column, err := common.IdentifierFromColumnIndex(i)
		if err != nil {
			c.String(http.StatusOK, lib_web.FmtJson(nil, errors.New(i18n.Show(common.GetLang(c), `sys_err`))))
			return
		}
		data[column] = v
	}

	c.String(http.StatusOK, lib_web.FmtJson(data, nil))
	return
}
