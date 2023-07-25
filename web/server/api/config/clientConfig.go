package config

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)
import "github.com/isrc-cas/gt/web/server/model/response"
import "github.com/isrc-cas/gt/web/server/model/request"

func ClientConfig(ctx *gin.Context) {
	// Create a buffer and read the request body into it
	var bodyBytes bytes.Buffer
	_, err := io.Copy(&bodyBytes, ctx.Request.Body)
	if err != nil {
		fmt.Println("Error reading request body: ", err)
		return
	}
	// Replace the request body with a new reader, so it can be read again later
	ctx.Request.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes.Bytes()))

	// DumpRequest returns the given request in its HTTP/1.x wire representation.
	// It should only be used by servers to debug client requests.
	//dump, err := httputil.DumpRequest(ctx.Request, true)
	//if err != nil {
	//	fmt.Println("Error dumping request: ", err)
	//	return
	//}

	//fmt.Println("Request: ", string(dump))

	var client request.ClientConfig
	// Use ShouldBindBodyWith to bind the request body into client struct
	if err := ctx.ShouldBindBodyWith(&client, binding.YAML); err != nil {
		fmt.Println("bind yaml failed", err)
		response.Fail(ctx, nil, "Invalid request ,can't bind")
		return
	}

	fmt.Println("client config: ", client)

	path, err := WriteYamlToFile("/home/seb/Desktop", "request.yaml", bodyBytes.Bytes())
	if err != nil {
		fmt.Println("write request to file failed", err)
		return
	}

	fmt.Println("write request to file: ", path)

	response.Success(ctx, gin.H{
		"message": "Calling client config",
	}, "done with client config")
}

func WriteYamlToFile(path string, filename string, data []byte) (string, error) {
	// Combine the path and filename to create a full file path
	fullPath := filepath.Join(path, filename)

	// Create and write to the file
	file, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return "", err
	}

	// Return the full file path
	return fullPath, nil
}
