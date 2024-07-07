package main

import (
	"bufio"
	"fmt"
	"twitterDownload/pkg/config"
	"twitterDownload/pkg/download"
	"twitterDownload/pkg/user"
	"twitterDownload/pkg/utils"

	"os"
	"sync"
)

var task sync.WaitGroup

var csvList = []utils.CSV{}

func downloadByUser(userName string) {
	userInfo,_ := user.FetchUserInfo(userName)
	userInfo.SaveDir = userName + "/"
	download.DownloadTwitterMedia(&userInfo, &csvList)
};

func menu() {
	fmt.Println("1. Get media by user")
	fmt.Println("2. Get media by userList")
	fmt.Println("3. Exit")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	choice := scanner.Text()

	switch choice {
	case "1":
		fmt.Println("Enter user name:")
		scanner.Scan()
		username := scanner.Text()
		downloadByUser(username)
	case "2":
		// 调用其他功能
		for _, user := range  config.SettingConfig.UserList {
			downloadByUser(user)
		}
	case "3":
		task.Done()
	default:
		fmt.Println("Invalid choice")
		menu()
	}
}

func main() {
	task.Add(1)
	menu()
	task.Wait()
}
