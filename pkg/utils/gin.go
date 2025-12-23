package utils

import (
	"sort"

	"github.com/gin-gonic/gin"
)

func DumpRouter(r *gin.Engine) map[string][]string {
	routes := r.Routes()
	routeMap := make(map[string][]string)
	for _, route := range routes {
		routeMap[route.Path] = append(routeMap[route.Path], route.Method)
	}
	for _, route := range routes {
		sortedKeys := routeMap[route.Path]
		sort.Strings(sortedKeys)
		routeMap[route.Path] = sortedKeys
	}
	return routeMap
}
