package simple

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
)

type RestRouter struct {
	Pattern   string                                                         //路径前缀
	Authorize func(pattern string, request *http.Request, params Value) bool //身份验证函数

	mutex    sync.Mutex
	handlers []*Handler
}

const (
	GET    = "GET"
	POST   = "POST"
	DELETE = "DELETE"
)

/* ----- router ------ */
func NewRestRouter(pattern string) *RestRouter {
	return &RestRouter{
		Pattern:  pattern,
		handlers: make([]*Handler, 0, 10),
	}
}

//添加处理方法
func (this *RestRouter) HandleFunc(pattern, method string, handle HandleFunc) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.handlers = append(this.handlers, newHandler(this.Pattern, pattern, method, handle))
}

//输出日志
func (this *RestRouter) log(request *http.Request, params, result Value) {
	in, _ := json.MarshalIndent(params, "", "	")
	out, _ := json.MarshalIndent(result, "", "	")
	log.Printf("%s\n%s\n%s\n\n", request.URL.Path, string(in), string(out))
}

//输出到客户端
func (this *RestRouter) output(response http.ResponseWriter, request *http.Request, params, result Value) {
	this.log(request, params, result)

	header := response.Header()
	header.Set("Content-Type", "application/json; charset=utf-8")
	header.Set("Cache-Control", "no-store, no-cache, must-revalidate")

	buf := bytes.NewBuffer(nil)
	encode := json.NewEncoder(buf)
	encode.Encode(result)

	//压缩
	if strings.Contains(request.Header.Get("Accept-Encoding"), "gzip") && buf.Len() > 1024 {
		header.Set("Content-Encoding", "gzip")

		w := gzip.NewWriter(response)
		defer w.Close()
		defer w.Flush()

		buf.WriteTo(w)
		return
	}

	buf.WriteTo(response)
}

//处理客户端请求
func (this *RestRouter) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			http.Error(response, "Bad request", http.StatusBadRequest)
		}
	}()

	//
	for _, h := range this.handlers {
		if !h.match.MatchString(request.URL.Path) || !strings.EqualFold(request.Method, h.method) {
			continue
		}

		//分解参数
		params := NewValue()
		h.urlParams(request.URL.Path, params)
		h.queryParams(request.URL, params)

		if request.ContentLength > 0 {
			h.streamParams(request.Body, params)
		}

		//授权检查
		if h.auth && this.Authorize != nil {
			if !this.Authorize(h.pattern, request, params) {
				this.output(response, request, params, NewValue().Failure("authorization failure"))
				return
			}
		}

		//调用处理器
		this.output(response, request, params, h.handle(params))
		return
	}

	http.NotFound(response, request)
}
