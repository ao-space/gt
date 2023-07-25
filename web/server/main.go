package main

import (
	"github.com/gin-gonic/gin"
	"github.com/isrc-cas/gt/web/server/router"
)

func main() {
	r := gin.Default()

	router.Router(r)
	//executeDir := "/home/seb/Desktop/gt/release"
	//executePath := "/home/seb/Desktop/gt/release/linux-amd64-client"
	//configPath := "/home/seb/Desktop/request.yaml"
	//if err := os.Chdir(executeDir); err != nil {
	//	fmt.Println("change dir failed", err)
	//}
	//dir, _ := os.Getwd()
	//fmt.Println("current dir: ", dir)
	//fmt.Println("start server")
	//service.StartServer(executePath, configPath)
	//pid, err := service.StartService(executePath, configPath)
	//if err != nil {
	//	fmt.Println("start client failed", err)
	//} else {
	//	fmt.Println("start client successfully:", pid)
	//	time.Sleep(3 * time.Second)
	//	//err := service.SendInterruptSignal(pid)
	//	//if err != nil {
	//	//	fmt.Println("send interrupt signal failed", err)
	//	//}
	//}

	//return
	r.Run(":8080")
}
