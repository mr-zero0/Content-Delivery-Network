package helper

import (
	"net/http"
)

func GetString(req *http.Request) (ret string) {
	/*	if req.URL.Scheme == "" {
			ret += "http://"
		} else {
			ret += req.URL.Scheme + "://"
		}
		if req.URL.Host == "" {
			ret += req.Host
		}
		ret += req.URL.String()
	*/
	return req.URL.String()
}
