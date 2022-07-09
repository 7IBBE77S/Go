package main

// import (

// 	"net/http"
// 	"strconv"
// 	"strings"

// 	"github.com/7IBBE77S/web-app/pkg/config"
// 	"github.com/7IBBE77S/web-app/pkg/handlers"
// )
	
// 	func routes(app *config.AppConfig) http.Handler {
	
// 		m := handlers.NewRepo(app)

// 		mux := http.NewServeMux()
// 		mux.HandleFunc("/", m.Home)
// 		mux.HandleFunc("/about", m.About)

// 		return mux
	
// 	}
	
// 	// match reports whether path matches the given pattern, which is a
// 	// path with '+' wildcards wherever you want to use a parameter. Path
// 	// parameters are assigned to the pointers in vars (len(vars) must be
// 	// the number of wildcards), which must be of type *string or *int.
// 	func match(path, pattern string, vars ...any) bool {
// 		for ; pattern != "" && path != ""; pattern = pattern[1:] {
// 			switch pattern[0] {
// 			case '+':
// 				// '+' matches till next slash in path
// 				slash := strings.IndexByte(path, '/')
// 				if slash < 0 {
// 					slash = len(path)
// 				}
// 				segment := path[:slash]
// 				path = path[slash:]
// 				switch p := vars[0].(type) {
// 				case *string:
// 					*p = segment
// 				case *int:
// 					n, err := strconv.Atoi(segment)
// 					if err != nil || n < 0 {
// 						return false
// 					}
// 					*p = n
// 				default:
// 					panic("vars must be *string or *int")
// 				}
// 				vars = vars[1:]
// 			case path[0]:
// 				// non-'+' pattern byte must match path byte
// 				path = path[1:]
// 			default:
// 				return false
// 			}
// 		}
// 		return path == "" && pattern == ""
// 	}
	
// 	func allowMethod(h http.HandlerFunc, method string) http.HandlerFunc {
// 		return func(w http.ResponseWriter, r *http.Request) {
// 			if method != r.Method {
// 				w.Header().Set("Allow", method)
// 				http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
// 				return
// 			}
// 			h(w, r)
// 		}
// 	}
	
// 	func Get(h http.HandlerFunc) http.HandlerFunc {
// 		return allowMethod(h, "GET")
// 	}
	
// 	func Post(h http.HandlerFunc) http.HandlerFunc {
// 		return allowMethod(h, "POST")
// 	}
	
	//LOOK INTO USING THE BELOW CODE

	// func match(path string, pieces ...interface{}) bool {
	// 	// Remove the initial "/" prefix
	// 	if strings.HasPrefix(path, "/") {
	// 		path = path[1:]
	// 	}
	// 	var head string
	// 	for i, piece := range pieces {
	// 		// Shift the next path component into `head`
	// 		head, path = nextComponent(path)
	// 		// Match pieces based on their type
	// 		switch p := piece.(type) {
	// 		case string:
	// 			// Match a specific string
	// 			if p != head {
	// 				return false
	// 			}
	// 		case *string:
	// 			// Match any string, including the empty string
	// 			*p = head
	// 		case *int64:
	// 			// Match any 64-bit integer, including negative integers
	// 			n, err := strconv.ParseInt(head, 10, 64)
	// 			if err != nil {
	// 				return false
	// 			}
	// 			*p = n
	// 		default:
	// 			panic(fmt.Sprintf("each piece must be a string, *string, or *int64. Got %T", piece))
	// 		}
	// 		// If the path is fully consumed, we're done if pieces are also fully consumed
	// 		if path == "" {
	// 			return i == len(pieces)-1
	// 		}
	// 	}
	// 	// Pieces are consumed; return true if the path is too.
	// 	return path == ""
	// }
	
	// // Accepts a path without leading slash and returns two strings:
	// // its first component and the rest without leading slash
	// func nextComponent(path string) (head, tail string) {
	// 	i := strings.IndexByte(path, '/')
	// 	if i == -1 {
	// 		return path, ""
	// 	}
	// 	return path[:i], path[i+1:]
	// }