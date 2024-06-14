// Copyright © 2016- 2024 Sesame Network Technology all right reserved

package chatwiki

import (
	"chatwiki/internal/app/chatwiki/business"
	"chatwiki/internal/app/chatwiki/common"
	"chatwiki/internal/app/chatwiki/define"
	"chatwiki/internal/app/chatwiki/initialize"
	"chatwiki/internal/pkg/lib_define"
	"chatwiki/internal/pkg/lib_web"
	"database/sql"
	"embed"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"strings"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cast"
	"github.com/zhimaAi/go_tools/logs"
	"github.com/zhimaAi/go_tools/mq"
	"github.com/zhimaAi/go_tools/msql"
	"github.com/zhimaAi/go_tools/tool"
)

func Run() {
	//initialize
	initialize.Initialize()
	//postgres table
	PostgresTable()
	//producer handle
	define.ProducerHandle = mq.NewProducerHandle().SetWorkNum(3).SetHostAndPort(define.Config.Nsqd[`host`], cast.ToUint(define.Config.Nsqd[`port`]))
	//consumer handle
	define.ConsumerHandle = mq.NewConsumerHandle().SetHostAndPort(define.Config.NsqLookup[`host`], cast.ToUint(define.Config.NsqLookup[`port`]))
	//web start
	go lib_web.WebRun(define.WebService)
	//pprof api
	go func() {
		err := http.ListenAndServe(":55557", nil)
		if err != nil {
			logs.Error(err.Error())
		}
	}()
	//consumer start
	StartConsumer()
}

func Stop() {
	define.ConsumerHandle.Stop()
	lib_web.Shutdown(define.WebService)
	define.ProducerHandle.Stop()
}

func StartConsumer() {
	common.RunTask(define.ConvertPdfTopic, define.ConvertPdfChannel, 1, business.ConvertPdf)
	common.RunTask(define.ConvertVectorTopic, define.ConvertVectorChannel, 2, business.ConvertVector)
	common.RunTask(lib_define.PushMessage, lib_define.PushChannel, 10, business.AppPush)
	common.RunTask(lib_define.PushEvent, lib_define.PushChannel, 5, business.AppPush)
}

//go:embed data/migrations/*.sql
var embedMigrations embed.FS

func PostgresTable() {
	conn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		define.Config.Postgres["host"], define.Config.Postgres["port"],
		define.Config.Postgres["user"], define.Config.Postgres["password"],
		define.Config.Postgres["dbname"], define.Config.Postgres["sslmode"])

	db, err := sql.Open("postgres", conn)
	if err != nil {
		panic(err)
	}

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		panic(err)
	}

	if err := goose.Up(db, "data/migrations", goose.WithAllowMissing()); err != nil {
		panic(err)
	}

	userId, err := CreateDefaultUser()
	if err != nil {
		logs.Error(err.Error())
	}
	if userId != 0 {
		CreateDefaultRole(userId)
		CreateDefaultBaaiModel(userId)
	}
}

func CreateDefaultUser() (int64, error) {
	m := msql.Model(define.TableUser, define.Postgres)
	user, err := m.Where(`"user_name"`, define.DefaultUser).Find()
	if err != nil {
		return 0, err
	}
	if len(user) > 0 {
		return 0, nil
	}

	salt := tool.Random(20)
	id, err := msql.Model(define.TableUser, define.Postgres).Insert(msql.Datas{
		`user_name`:   define.DefaultUser,
		`salt`:        salt,
		`password`:    tool.MD5(define.DefaultPasswd + salt),
		`user_type`:   define.UserTypeAdmin,
		`user_roles`:  define.UserTypeAdmin,
		`create_time`: tool.Time2Int(),
		`update_time`: tool.Time2Int(),
	}, "id")
	if err != nil {
		logs.Error(`user create err:%s`, err.Error())
		return 0, err
	}
	return id, nil
}
func CreateDefaultRole(userId int64) {
	var defaultRole = []string{define.DefaultRoleRoot, define.DefaultRoleAdmin, define.DefaultRoleUser}
	for k, role := range defaultRole {
		_, err := msql.Model(define.TableRole, define.Postgres).Insert(msql.Datas{
			`name`:        role,
			"role_type":   k + 1,
			`create_name`: "系统",
			`create_time`: tool.Time2Int(),
			`update_time`: tool.Time2Int(),
		})
		if err != nil {
			logs.Error(`role create err:%s`, err.Error())
		}
		if userId <= 0 || role != define.DefaultRoleRoot {
			continue
		}
	}
}
func CreateDefaultBaaiModel(userId int64) {
	modelInfo, ok := common.GetModelInfoByDefine(define.ModelBaai)
	if !ok {
		logs.Error(`modelInfo not found`)
		return
	}
	_, err := msql.Model("chat_ai_model_config", define.Postgres).Insert(msql.Datas{
		`admin_user_id`:   userId,
		`model_define`:    define.ModelBaai,
		`model_types`:     strings.Join(modelInfo.SupportedType, `,`),
		`api_endpoint`:    "http://host.docker.internal:50001",
		`deployment_name`: "",
		`create_time`:     tool.Time2Int(),
		`update_time`:     tool.Time2Int(),
	})
	if err != nil {
		logs.Error("baai model create err:%s", err.Error())
	}
}
