package router

import (
	"errors"
	"net/http"
	"path"
	"strings"

	"github.com/zeromicro/go-zero/core/search"
	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/pathvar"
)

const (
	allowHeader          = "Allow"
	allowMethodSeparator = ", "
)

var (
	// ErrInvalidMethod is an error that indicates not a valid http method.
	ErrInvalidMethod = errors.New("not a valid http method")
	// ErrInvalidPath is an error that indicates path is not start with /.
	ErrInvalidPath = errors.New("path must begin with '/'")
)

type patRouter struct {
	trees      map[string]*search.Tree // search.Tree 是搜索树
	notFound   http.Handler
	notAllowed http.Handler
}

// NewRouter returns a httpx.Router.
func NewRouter() httpx.Router {
	return &patRouter{
		trees: make(map[string]*search.Tree),
	}
}

// Handle
// method：http 请求方法
// reqPath：http 请求路径
// handler：响应 handler，包括预置、用户的 middleware
func (pr *patRouter) Handle(method, reqPath string, handler http.Handler) error {
	if !validMethod(method) { // 方法校验
		return ErrInvalidMethod
	}

	if len(reqPath) == 0 || reqPath[0] != '/' { // 路径校验
		return ErrInvalidPath
	}

	// 请求路径预处理
	//	1. Replace multiple slashes with a single slash. "//"->"/"
	//	2. Eliminate each . path name element (the current directory).
	//	3. Eliminate each inner .. path name element (the parent directory)
	//	   along with the non-.. element that precedes it.
	//	4. Eliminate .. elements that begin a rooted path:
	//	   that is, replace "/.." by "/" at the beginning of a path.
	cleanPath := path.Clean(reqPath)
	tree, ok := pr.trees[method]
	if ok { // 该方法的搜索树已被初始化
		return tree.Add(cleanPath, handler) // 添加 路径 & 处理方法
	}

	// 未初始化的搜索树
	// 需要初始化
	tree = search.NewTree()
	pr.trees[method] = tree
	return tree.Add(cleanPath, handler)
}

func (pr *patRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	reqPath := path.Clean(r.URL.Path)
	if tree, ok := pr.trees[r.Method]; ok {
		if result, ok := tree.Search(reqPath); ok {
			if len(result.Params) > 0 {
				r = pathvar.WithVars(r, result.Params)
			}
			result.Item.(http.Handler).ServeHTTP(w, r)
			return
		}
	}

	allows, ok := pr.methodsAllowed(r.Method, reqPath)
	if !ok {
		pr.handleNotFound(w, r)
		return
	}

	if pr.notAllowed != nil {
		pr.notAllowed.ServeHTTP(w, r)
	} else {
		w.Header().Set(allowHeader, allows)
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (pr *patRouter) SetNotFoundHandler(handler http.Handler) {
	pr.notFound = handler
}

func (pr *patRouter) SetNotAllowedHandler(handler http.Handler) {
	pr.notAllowed = handler
}

func (pr *patRouter) handleNotFound(w http.ResponseWriter, r *http.Request) {
	if pr.notFound != nil {
		pr.notFound.ServeHTTP(w, r)
	} else {
		http.NotFound(w, r)
	}
}

func (pr *patRouter) methodsAllowed(method, path string) (string, bool) {
	var allows []string

	for treeMethod, tree := range pr.trees {
		if treeMethod == method {
			continue
		}

		_, ok := tree.Search(path)
		if ok {
			allows = append(allows, treeMethod)
		}
	}

	if len(allows) > 0 {
		return strings.Join(allows, allowMethodSeparator), true
	}

	return "", false
}

func validMethod(method string) bool {
	return method == http.MethodDelete || method == http.MethodGet ||
		method == http.MethodHead || method == http.MethodOptions ||
		method == http.MethodPatch || method == http.MethodPost ||
		method == http.MethodPut
}
