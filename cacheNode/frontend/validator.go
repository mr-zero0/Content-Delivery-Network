package frontend

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"errors"
	"bytes"
	"io"
	"time"
	"log/slog"
	"github.com/hcl/cdn/cacheNode/common"
)

type Validator struct {
	BackendRequest common.CachedRequestHandler
	RespFromBackend http.Response
	ExpiredCnt, UsableCnt uint32
}

//CopyResponse copies the response headers and body from one response to another
func CopyResponse(src *http.Response, dest *http.Response) error {
	slog.Info("FE validator.go : CopyResponse() - Start")

	// Copy headers
	for key, extVal := range src.Header {
		for _, innerVal := range extVal {
			dest.Header.Add(key, innerVal)
		}
	}

	// Copy status code
	dest.StatusCode = src.StatusCode
	dest.Proto = src.Proto
	dest.ProtoMajor = src.ProtoMajor
	dest.ProtoMinor = src.ProtoMinor
	dest.Request = src.Request

	buf := &bytes.Buffer{}
	// Copy the body
	_, err := io.Copy(buf, src.Body)
	if err != nil {
		slog.Error("FE validator.go : Error copying response body", "error", err.Error())
		return err
	}

	dest.Body = io.NopCloser(bytes.NewReader(buf.Bytes()))
	defer src.Body.Close()

	dest.ContentLength = int64(len(buf.Bytes()))

	slog.Info("FE validator.go : CopyResponse() - Success")
	return nil
}

// IsResourceUsable checks if the resource is still usable based on the Age and Max-Age headers
func (v *Validator) IsResourceUsable(resp *http.Response, req *http.Request) (bool, error) {
	slog.Info("FE validator.go : IsResourceUsable() - Start")

	// Get Value of the Age header field from Response
	ageStr := resp.Header.Get("Age")
	age := 0
	if len(ageStr) > 0 {
		var err error
		// Convert from string to integer
		age, err = strconv.Atoi(ageStr)
		if err != nil {
			slog.Error("FE validator.go : Invalid Age header", "error", err.Error())
			return false, fmt.Errorf("invalid Age header: %v", err)
		}
	}

	cacheControl := resp.Header.Get("Cache-Control")
	var maxAge int
	if cacheControl != "" {
		for _, directive := range strings.Split(cacheControl, ",") {
			directive = strings.TrimSpace(directive)
			// Use 'Prefix' logic to extract the max-age value in seconds
			if strings.HasPrefix(directive, "max-age=") {
				// Extract the time value for max-age directive in string
				maxAgeStr := strings.TrimPrefix(directive, "max-age=")
				var err error
				// Convert string to integer
				maxAge, err = strconv.Atoi(maxAgeStr)
				if err != nil {
					slog.Error("FE validator.go : Invalid Max-Age value", "error", err.Error())
					return false, fmt.Errorf("invalid Max-Age value: %v", err)
				}
			}
		}
	}

	// Compare Age and Max-Age
	if maxAge > 0 && age <= maxAge {
		slog.Info("FE validator.go : Resource is usable")
		slog.Info("FE validator.go : IsResourceUsable() - End (Storage SUCCESS)")

		//Adding a custom header to indicate that the response is from the Storage cache
		resp.Header.Add("X-Is-Cached", "1")

		return true, nil // Resource is usable
	} else if (maxAge < age || maxAge == 0) {
		// Resource expired - invoke Backend.ReDo()
		slog.Info(fmt.Sprintf("FE validator.go : WARN : Resource expired, calling Backend.ReDo(), maxAge : %d, age : %d", maxAge, age))
		rspFromBknd, err := v.BackendRequest.ReDo(req, resp)

		if err != nil {
			slog.Error("FE validator.go : Backend.ReDo() failed ", "error", err.Error())
			return false, err
		} else {
			slog.Info(fmt.Sprintf("FE validator.go : Backend.ReDo() returned Status Code : %d", rspFromBknd.StatusCode))
			slog.Info("FE validator.go : Copying response from Backend to Validator's backend member")

			if rspFromBknd.StatusCode == http.StatusNotFound{
				slog.Info("FE validator.go : WARNING: Backend.ReDo() did not found the resource")
			}

			tempRsp := &http.Response{
				Header : make(http.Header),
			}

			err := CopyResponse(rspFromBknd, tempRsp)
			if err != nil {
				slog.Error("FE validator.go : CopyResponse() error ", "error", err.Error())
				return false, err
			}
			slog.Info("FE validator.go : IsResourceUsable() - End (Backend ReDo() SUCCESS)")
			v.RespFromBackend = *tempRsp
			return true, nil
		}
	}

	// Final
	slog.Info("FE validator.go : Resource is not usable")
	slog.Info("FE validator.go : IsResourceUsable() - End")
	return false, nil
}

// Do method handles the main validator process, checking if a resource should be fetched again from the backend
func (v *Validator) Do(req *http.Request, oResp *http.Response) (*http.Response, error) {
	slog.Info("FE validator.go : Do() - Start")

	bkndResp := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Status:     "500 Internal Server Error",
		Header:     make(http.Header),
	}
	bkndResp.Header.Set("Accept", "*/*")
	// Set the Date header
	currentTime := time.Now().UTC().Format(http.TimeFormat)
	// Format the date in RFC1123 format
	bkndResp.Header.Set("Date", currentTime)
	// Set the User-Agent header
	bkndResp.Header.Set("User-Agent", "Go Server Agent/1.23.4")

	v.RespFromBackend.StatusCode = http.StatusInternalServerError           
	// Check the usability status of the Storage Response
	isUsable, err := v.IsResourceUsable(oResp, req)

	if err != nil {
		slog.Error("FE validator.go : IsResourceUsable Error", "error", err.Error())
		bkndResp.ContentLength = int64(len(err.Error()))
		bkndResp.Body = io.NopCloser(strings.NewReader(err.Error()))
		return bkndResp, nil
	}

	if isUsable && v.RespFromBackend.StatusCode == http.StatusInternalServerError {
		// Storage's Status Code returned : 500
		slog.Info("FE validator.go : Storage has the requested resource")
		slog.Info("FE validator.go : Do() - End (Storage : SUCCESS)")
		return oResp, nil

	} else if (isUsable && (v.RespFromBackend.StatusCode < http.StatusInternalServerError  &&  v.RespFromBackend.StatusCode >= http.StatusOK)){		
		slog.Info("FE validator.go : Storage has the expired resource and backend responded for the same")
		slog.Info("FE validator.go : Do() - End (Backend Response)")
		return &v.RespFromBackend, nil

	} else if !isUsable {
		slog.Info("FE validator.go : Validator Failure")
		slog.Info("FE validator.go : Do() - End")
		
		errStr := errors.New("Internal Server Error")
		bkndResp.ContentLength = int64(len(errStr.Error()))
		bkndResp.Body = io.NopCloser(strings.NewReader(errStr.Error()))
		return bkndResp, nil
	}

	// Default error case
	slog.Info("FE validator.go : Validator Error")
	slog.Info("FE validator.go : Do() - End")

	errStr := errors.New("Internal Server Error")
	bkndResp.ContentLength = int64(len(errStr.Error()))
	bkndResp.Body = io.NopCloser(strings.NewReader(errStr.Error()))
	return bkndResp, nil
}
