package researd

import (
	"net/http"
	"strings"
	"time"
)

type Router struct {
	searcher   *Searcher
	addressMap map[string][]string
}

func newRouter(searcher *Searcher) *Router {
	return &Router{
		searcher:   searcher,
		addressMap: map[string][]string{},
	}
}

func (router *Router) handler(w http.ResponseWriter, r *http.Request) {
	// 获取请求的路径并去掉开头的 '/'
	path := strings.TrimPrefix(r.URL.Path, "/")

	// 以 '/' 分割路径，获取第一个参数
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[0] != "GetAddressByServerName" {
		http.Error(w, "路径格式错误，应为 /GetAddressByServerName/<serverName>", http.StatusBadRequest)
		return
	}

	// 提取第一个参数
	name := parts[1]

	// 返回一个字符串作为响应
	response := router.getAddress(name)
	w.Write([]byte(response))
}

func (router *Router) Run(address string) {
	http.HandleFunc("/", router.handler)
	http.ListenAndServe(address, nil)
}

func (router *Router) getAddress(serverName string) string {
	if len(router.addressMap[serverName]) != 2 {
		address, _ := router.searcher.GetHighestPriorityServer(serverName)
		if address != "" {
			router.addressMap[serverName] = []string{address, getCurrentTimeString()}
		} else {
			return ""
		}
	}
	go router.refresh(serverName)
	return router.addressMap[serverName][0]
}

func (router *Router) refresh(serverName string) {
	if isMoreThanTwoSecondsAgo(router.addressMap[serverName][1]) {
		address, _ := router.searcher.GetHighestPriorityServer(serverName)
		router.addressMap[serverName][0] = address
		router.addressMap[serverName][1] = getCurrentTimeString()
	}
}

// 获取当前时间并转换为 UTC 字符串
func getCurrentTimeString() string {
	currentTime := time.Now().UTC() // 设置为 UTC
	return currentTime.Format("2006-01-02 15:04:05")
}

// 将时间字符串解析为 UTC 时间
func parseTimeString(timeString string) (time.Time, error) {
	layout := "2006-01-02 15:04:05"
	parsedTime, err := time.Parse(layout, timeString)
	if err != nil {
		return time.Time{}, err
	}
	return parsedTime.UTC(), nil // 设置为 UTC
}

// 比较字符串时间和当前时间，判断是否超过 2 秒
func isMoreThanTwoSecondsAgo(timeString string) bool {
	parsedTime, err := parseTimeString(timeString)
	if err != nil {
		return true // 如果解析出错，直接返回 true
	}

	currentTime := time.Now().UTC() // 统一设置为 UTC
	twoSecondsLater := parsedTime.Add(2 * time.Second)

	return currentTime.After(twoSecondsLater)
}
