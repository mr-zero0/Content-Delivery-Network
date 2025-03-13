package backendTestMock

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type RequestHandlerMock struct {
	HttpStatuscode      int
	Header              http.Header
	ResponseBodyContent string
	ExpectedReqBody     string
}

func (ms RequestHandlerMock) Do(req *http.Request) (*http.Response, error) {
	resp := http.Response{
		StatusCode: ms.HttpStatuscode,
		Header:     ms.Header,
		Body:       nil,
	}
	fmt.Println("Hi welcome to enetered mock handler")
	switch req.Method {
	case http.MethodGet:
		buf := bytes.NewBuffer([]byte(ms.ResponseBodyContent))
		resp.Body = io.NopCloser(buf)
	case http.MethodPut, http.MethodPost:
		content, err := io.ReadAll(req.Body)
		if err != nil {
			resp.StatusCode = http.StatusInternalServerError
			return &resp, err
		}
		fmt.Println(content)
		if !bytes.Equal(content, []byte(ms.ExpectedReqBody)) {
			fmt.Println("Hi welcome to interface mock 2")
			resp.StatusCode = http.StatusInternalServerError
			return &resp, errors.New("expected response body not arrived")
		}
	}
	return &resp, nil
}
