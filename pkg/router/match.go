package router

import (
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	validRepoRoute = regexp.MustCompile(`^.*\.(yaml|tgz|prov)$`)
)

func match(routes []*Route, method string, url string, contextPath string, depth int) (*Route, []gin.Param) {
	var noRepoPathSplit []string
	var repo, repoPath, noRepoPath string
	var startIndex, numNoRepoPathParts int
	var tryRepoRoutes bool

	if contextPath != "" {
		if url == contextPath {
			url = "/"
		} else if strings.HasPrefix(url, contextPath) {
			url = strings.Replace(url, contextPath, "", 1)
		} else {
			return nil, nil
		}
	}

	isApiRoute := checkApiRoute(url)
	if isApiRoute {
		startIndex = 2
	} else {
		startIndex = 1
	}

	pathSplit := strings.Split(url, "/")
	numParts := len(pathSplit)

	if numParts >= depth+startIndex {
		repoParts := pathSplit[startIndex : depth+startIndex]
		if len(repoParts) == depth {
			tryRepoRoutes = true
			repo = strings.Join(repoParts, "/")
			noRepoPath = "/" + strings.Join(pathSplit[depth+startIndex:], "/")
			repoPath = "/:repo" + noRepoPath
			if isApiRoute {
				repoPath = "/api" + repoPath
				noRepoPath = "/api" + noRepoPath
			}
			noRepoPathSplit = strings.Split(noRepoPath, "/")
			numNoRepoPathParts = len(noRepoPathSplit)
		}
	}

	for _, route := range routes {
		if route.Method != method {
			continue
		}
		if route.Path == url {
			return route, nil
		} else if tryRepoRoutes {
			if route.Path == repoPath {
				return route, []gin.Param{{"repo", repo}}
			} else {
				p := strings.Replace(route.Path, "/:repo", "", 1)
				if routeSplit := strings.Split(p, "/"); len(routeSplit) == numNoRepoPathParts {
					isMatch := true
					var params []gin.Param
					for i, part := range routeSplit {
						if paramSplit := strings.Split(part, ":"); len(paramSplit) > 1 {
							params = append(params, gin.Param{Key: paramSplit[1], Value: noRepoPathSplit[i]})
						} else if routeSplit[i] != noRepoPathSplit[i] {
							isMatch = false
							break
						}
					}
					if isMatch {
						params = append(params, gin.Param{Key: "repo", Value: repo})
						return route, params
					}
				}
			}
		}
	}

	return nil, nil
}

func checkApiRoute(url string) bool {
	return strings.HasPrefix(url, "/api/") && !validRepoRoute.MatchString(url)
}
