package simple

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

type (
	HandleFunc func(Value) Value

	Handler struct {
		auth       bool
		pattern    string
		method     string
		match      *regexp.Regexp
		matchNames []string
		handle     HandleFunc
	}
)

/* ------------- handler ---------*/

func newHandler(prefix, pattern, method string, handle HandleFunc) *Handler {
	// pattern 以 "!" 开头表示该调用需要身份验证。
	var auth bool
	if pattern[0] == '!' {
		pattern = pattern[1:]
		auth = true
	}

	pattern = filepath.Join(prefix, pattern)

	handler := &Handler{
		auth:    auth,
		pattern: pattern,
		method:  strings.ToUpper(method),
		handle:  handle,
	}

	handler.match = handler.regex(pattern)
	handler.matchNames = handler.match.SubexpNames()

	return handler
}

// 将路径匹配模式转换为正则表达式。
func (this *Handler) regex(s string) *regexp.Regexp {
	reg, _ := regexp.Compile(`{(\w+)}`)
	expr := reg.ReplaceAllString(s, `(?P<$1>[^/\?]+)`)

	reg, err := regexp.Compile("^" + expr)
	if err != nil {
		panic(err)
	}

	return reg
}

//从路径中提取参数
func (this *Handler) urlParams(path string, value Value) {
	for i, v := range this.match.FindStringSubmatch(path) {
		if i == 0 {
			continue
		}

		value[this.matchNames[i]] = v
	}
}

//从URL QUERY中提取参数
func (this *Handler) queryParams(url *url.URL, value Value) {
	for k, v := range url.Query() {
		if len(v) == 1 {
			value[k] = v[0]
		} else {
			value[k] = v
		}
	}
}

//从 Stream 中提取 参数
func (this *Handler) streamParams(r io.Reader, value Value) {
	//读取全部内容
	bs, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}

	//JSON
	decode := json.NewDecoder(bytes.NewReader(bs))

	var d Value
	if err := decode.Decode(&d); err == nil {
		for k, v := range d {
			value[k] = v
		}

		return
	}

	//FORM
	u, err := url.Parse("?" + string(bs))
	if err != nil {
		panic(err)
	}

	this.queryParams(u, value)
}
