package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gogf/gf/container/gmap"
	"github.com/gogf/gf/database/gdb"
	"github.com/gogf/gf/encoding/gjson"
	"github.com/gogf/gf/errors/gcode"
	"github.com/gogf/gf/errors/gerror"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/glog"
	"github.com/gogf/gf/util/gconv"
	uuid "github.com/satori/go.uuid"
	"io"
	"math/rand"
	"os"
	"strings"
)

//将字符串写入文件
func WriteFile(content string, filepath string, replace bool) error {
	var f *os.File
	var err error
	if CheckFileIsExist(filepath) {
		if !replace {
			err = errors.New(" 已存在，需删除才能重新生成...")
		}
		f, err = os.OpenFile(filepath, os.O_WRONLY|os.O_TRUNC, 0666) //打开文件
		if err != nil {
			err = errors.New("打开文件出错")
		}
	} else {
		f, err = os.Create(filepath)
		if err != nil {
			err = errors.New("创建文件出错")

		}
	}
	defer f.Close()
	_, err = io.WriteString(f, content)
	if err != nil {
		err = errors.New("文件写入出错")
	}

	return err
}

// DeleEmptyValueMpa 删除map中的空值
func DeleEmptyValueMpa(vm map[string]interface{}) map[string]interface{} {
	for k, v := range vm {
		if gconv.String(v) == "" {
			delete(vm, k)
		}
	}
	return vm
}

// MapToChatString 将map按k+分割字符串+v方式生成字符串
func MapToChatString(vm map[string]interface{}, chat string) string {

	rs := ""
	if len(vm) > 0 {
		for k, v := range vm {
			rs = rs + gconv.String(k) + chat + gconv.String(v) + chat
		}
		return rs[:len(rs)-1]
	} else {
		return ""
	}

}

// KvSplitMap k+分隔符+v 这种字符串，分割成map
func KvSplitMap(chat, s string) map[string]string {

	ts := strings.Split(s, chat)
	i := 0
	res := make(map[string]string)
	for i = 0; i < len(ts)-1; i++ {
		if i%2 == 0 {
			res[gconv.String(ts[i])] = gconv.String(ts[i+1])
		}
	}

	return res

}

// Map_merge map合并，map2追加合并到到map1中
func Map_merge(map1, map2 map[int]interface{}) map[int]interface{} {
	map3 := make(map[int]interface{}, len(map1)+len(map2))
	if map1len := len(map1); map1len >= 0 {
		i := 0
		for _, v1 := range map1 {
			map3[i] = v1
			i = i + 1
		}
		for _, v2 := range map2 {
			map3[i] = v2
			i = i + 1
		}

	}
	return map3
}

// GetCount 显示查询结果总条数
func GetCount(gmodel gdb.DB, mode interface{}, where ...interface{}) int {
	total, err := gmodel.Model(mode).FindCount(where...)
	if err != nil {
		glog.Error("获取用户数量错误", err)
		return 0
	}

	return total
}

// FailOnError 统一异常处理，error错误用于log记录，msg显示前端错误信息，错误代码:1000以上显示前端，1000以下前端不显示
func FailOnError(err error, msg string, code int, skip ...int) {
	if err != nil {
		if len(skip) == 1 {
			g.Log().Stdout(false).Skip(skip[0]).Stack(true).Line(true).Printf(msg + "---*****---" + err.Error())
		} else {
			g.Log().Stdout(false).Skip(1).Stack(true).Line(true).Printf(msg + "---*****---" + err.Error())
		}
		err = gerror.NewCode(gcode.New(code, msg, nil))
		panic(err)
	}
}

// 生成uuid
func GetUUID() string {
	u2 := uuid.NewV4()
	return u2.String()
}

// 生成随机字符串
func RandStringBytes(n int) string {
	const letterBytes = "123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// GetAcceessToken 获取token
func GetAcceessToken() (string, error) {

	appid := g.Cfg().Get("wechat.app_id") //获取微信appid

	appToken, err := g.Redis().DoVar("GET", appid) //首先判断redis中有没有已经获取的token
	FailOnError(err, "redis 获取失败", 1003)
token:
	if appToken.Val() == nil {
		//获取apptoken token:
		access_token := "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s"
		url := fmt.Sprintf(access_token, appid, g.Cfg().Get("wechat.app_secret"))
		get, err := g.Client().Get(url)
		FailOnError(err, "获取token错误", 1001)
		mapResult := gmap.New()

		err = json.Unmarshal([]byte(get.ReadAllString()), &mapResult)
		FailOnError(err, "获取token错误", 1001)

		if mapResult.Contains("errcode") {
			g.Log(gconv.String(mapResult.Get("errmsg")))
			FailOnError(err, "获取token错误", 1001)
		}

		//获取成功，存入redis，并设置90分钟超时
		appToken.Set(mapResult.Get("access_token"))
		_, err = g.Redis().Do("SETEX", appid, "5400", appToken.Val())
		FailOnError(err, "存入redis错误", 1002)
	} else { //测试token是否有效
		test_url := "https://api.weixin.qq.com/cgi-bin/message/template/send?access_token=" + gconv.String(appToken.Val())
		testjson := "{\"begin_date\": \"2021-01-01\",\"end_date\": \"2021-01-06\"\n}"
		post, err := g.Client().Post(test_url, testjson)
		FailOnError(err, "token测试失败", 1002)
		js := gjson.New(post.ReadAllString())
		mapResult := js.Map()
		if mapResult["errcode"] == 0 {
			return "发送成功", nil
		}
		//token无效重新获取token
		if mapResult["errcode"] != 0 {
			appToken.Set(nil)
			goto token //重新获取accesstoken
		}

	}
	return gconv.String(appToken.Val()), nil

}

// SendMsg 发送微信模版消息
func SendMsg(Body, To, Templateid, url, mini string) (string, error) {
	j1 := gjson.New(Body) //body是json对象，要转义获得，否则是字符串
	j2 := gjson.New(mini)

	sendData := g.Map{"touser": To, "template_id": Templateid, "url": url, "miniprogram": j2, "data": j1} //拼装请求参数
	token, err := GetAcceessToken()                                                                       //获取token
	if err == nil {
		url := "https://api.weixin.qq.com/cgi-bin/message/template/send?access_token=" + token
		mjson, _ := json.Marshal(sendData)
		post, err := g.Client().Post(url, string(mjson))
		FailOnError(err, "发送信息错误", 1004)

		js := gjson.New(post.ReadAllString()) //获取返回信息
		mapResult := js.Map()
		if mapResult["errcode"] == 0 {
			return "发送成功", nil
		}

	} else {
		FailOnError(err, "token错误", 1005)
	}

	return "", nil

}

// CheckFileIsExist 检查文件是否存在
func CheckFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}
