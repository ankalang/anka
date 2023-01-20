package repl

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strings"

	"github.com/ankalang/anka/util"
)


const ANK_INIT_FILE = "~/.ankrc"

func getAbsInitFile(interactive bool) {
	
	initFile := os.Getenv("ANK_INIT_FILE")
	if len(initFile) == 0 {
		initFile = ANK_INIT_FILE
	}
	
	filePath, err := util.ExpandPath(initFile)
	if err != nil {
		fmt.Printf("Unable to expand ANK init file path: %s\nError: %s\n", initFile, err.Error())
		os.Exit(99)
	}
	initFile = filePath
	
	code, err := ioutil.ReadFile(initFile)
	if err != nil {
		
		return
	}
	Run(string(code), interactive)
}


const ANK_PROMPT_PREFIX = ">  "


func formatLivePrefix(prefix string) string {
	livePrefix := prefix
	if strings.Contains(prefix, "{") {
		userInfo, _ := user.Current()
		user := userInfo.Username
		host, _ := os.Hostname()
		dir, _ := os.Getwd()
		
		homeDir := userInfo.HomeDir
		dir = strings.Replace(dir, homeDir, "~", 1)
		
		livePrefix = strings.Replace(livePrefix, "{user}", user, 1)
		livePrefix = strings.Replace(livePrefix, "{host}", host, 1)
		livePrefix = strings.Replace(livePrefix, "{dir}", dir, 1)
	}
	return livePrefix
}
