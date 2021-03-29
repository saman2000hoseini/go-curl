package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/saman2000hoseini/go-curl/model"
	"github.com/saman2000hoseini/go-curl/pkg"

	"github.com/spf13/cobra"
)

const defaultMethod = "GET"

func main(request model.Request, arg string) {
	httpURL, err := url.Parse(arg)
	if err != nil {
		log.Fatal("invalid url")
	}

	header := http.Header{}

	for i := range request.Headers {
		headers := strings.Split((request.Headers)[i], ",")
		for _, h := range headers {
			pair := strings.Split(h, ":")
			header.Set(pair[0], pair[1])
		}
	}

	formData := url.Values{}
	var body io.ReadCloser

	for i := range request.FormData {
		if header.Get(pkg.ContentType) == "" {
			header.Set(pkg.ContentType, pkg.FormType)
		}

		data := strings.Split((request.FormData)[i], "&")
		for _, d := range data {
			pair := strings.Split(d, "=")
			if !strings.HasPrefix(pair[1], "x-") {
				fmt.Println("Warning!! form data is not in correct form")
			}

			formData.Set(pair[0], pair[1])
		}

		body = io.NopCloser(strings.NewReader(formData.Encode()))
	}

	if request.JsonData != "" {
		if !json.Valid([]byte(request.JsonData)) {
			log.Println("Warning!! json is not in correct form")
		}

		header.Set(pkg.ContentType, pkg.JsonType)
		body = io.NopCloser(strings.NewReader(request.JsonData))
	}

	if request.FilePath != "" {
		if header.Get(pkg.ContentType) == "" {
			header.Set(pkg.ContentType, pkg.FileType)
		}

		bodyBuf := &bytes.Buffer{}
		bodyWriter := multipart.NewWriter(bodyBuf)

		fileWriter, err := bodyWriter.CreateFormFile("uploadfile", request.FilePath)
		if err != nil {
			log.Fatal("error writing to buffer")
		}

		fh, err := os.Open(request.FilePath)
		if err != nil {
			log.Fatal("error opening file")
		}
		defer fh.Close()

		_, err = io.Copy(fileWriter, fh)
		if err != nil {
			log.Fatal("upload failed")
		}

		bodyWriter.Close()
		body = io.NopCloser(strings.NewReader(bodyBuf.String()))
	}

	if len(request.Queries) > 0 {
		queryParam := strings.Join(request.Queries, "&")
		httpURL, _ = url.Parse(fmt.Sprintf("%s?%s", httpURL, queryParam))
	}

	req := &http.Request{
		Method: request.Method,
		URL:    httpURL,
		Header: header,
		Body:   body,
	}

	if request.Duration != 0 {
		go timeout(request.Duration)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err.Error())
	}

	defer res.Body.Close()

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)

	fmt.Println(res.Status)
	fmt.Println(res.Request.Method)

	for key, values := range res.Header {
		fmt.Printf("%s:\n", key)
		for _, value := range values {
			fmt.Printf("%s\n", value)
		}
		fmt.Println("----------------------------------------------------------------------")
	}
	log.Println(bodyString)

	filepath := path.Base(httpURL.Path)
	if strings.Contains(filepath, ".") {
		os.Mkdir("storage", 0755)

		out, err := os.Create("storage/" + filepath)
		if err != nil {
			log.Println("failed creating file")
			return
		}
		defer out.Close()

		_, err = out.Write(bodyBytes)
		if err != nil {
			log.Println("failed copy file context")
		}
	}
}

func timeout(duration int64) {
	<-time.Tick(time.Duration(duration) * time.Second)

	log.Fatal("timeout exceeded")
}

func NewCommand() *cobra.Command {
	req := model.Request{}

	var command = &cobra.Command{
		Use:  "go-curl",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			main(req, args[0])
		},
	}

	command.Flags().StringVarP(&req.Method, "method", "M", defaultMethod, pkg.MethodUsage)
	command.Flags().StringSliceVarP(&req.Headers, "headers", "H", nil, pkg.HeaderUsage)
	command.Flags().StringSliceVarP(&req.Queries, "queries", "Q", nil, pkg.QueriesUsage)
	command.Flags().StringSliceVarP(&req.FormData, "data", "D", nil, pkg.DataUsage)
	command.Flags().StringVar(&req.JsonData, "json", "", pkg.JsonUsage)
	command.Flags().StringVar(&req.FilePath, "file", "", pkg.FileUsage)
	command.Flags().Int64Var(&req.Duration, "timeout", 0, pkg.TimeoutUsage)

	return command
}
