package repo

import (
	"bufio"
	"net/http"
	"strings"
	"sync"

	"github.com/richardwilkes/gopathdep/util"
)

const goImportMeta = `<meta name="go-import" content="`

var (
	gitRemoteCache     = make(map[string]string)
	gitRemoteCacheLock sync.Mutex
)

// GitRemote returns the git URL for the package.
func GitRemote(pkg string) string {
	gitRemoteCacheLock.Lock()
	defer gitRemoteCacheLock.Unlock()
	url, exists := gitRemoteCache[pkg]
	if !exists {
		url = scanForGoImport("https", pkg)
		if url == "" {
			url = scanForGoImport("http", pkg)
		}
		if url == "" {
			url = "https://" + pkg
		}
		gitRemoteCache[pkg] = url
	}
	return url
}

func scanForGoImport(protocol, pkg string) string {
	if resp, err := http.Get(protocol + "://" + pkg + "?go-get=1"); err == nil {
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if i := strings.Index(line, goImportMeta); i != -1 {
				line = line[i+len(goImportMeta):]
				if i = strings.Index(line, `"`); i != -1 {
					parts := strings.Split(line[:i], " ")
					if len(parts) >= 3 && parts[0] == pkg && parts[1] == "git" {
						return parts[2]
					}
					break
				}
			}
		}
		if err = resp.Body.Close(); err != nil {
			util.Ignore()
		}
	}
	return ""
}
