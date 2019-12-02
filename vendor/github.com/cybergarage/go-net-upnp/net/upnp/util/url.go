// Copyright 2015 Satoshi Konno. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"fmt"
	"net/url"
	"strings"
)

const (
	urlDelim = "/"
)

const (
	errorUrlNotAbsolute   = "url (%s) is not absolute"
	errorUrlUnknownScheme = "url scheme (%s) is unknown"
)

func GetAbsoluteURLFromBaseAndPath(base string, path string) (*url.URL, error) {
	urlobj, err := url.Parse(path)

	if err != nil || !urlobj.IsAbs() {
		base = strings.TrimSuffix(base, urlDelim)
		path = strings.TrimPrefix(path, urlDelim)

		urlStr := base + urlDelim + path
		urlobj, err = url.Parse(urlStr)
		if err != nil {
			return nil, err
		}
	}

	if !urlobj.IsAbs() {
		return nil, fmt.Errorf(errorUrlNotAbsolute, urlobj.String())
	}

	if (urlobj.Scheme != "http") && (urlobj.Scheme != "https") {
		return nil, fmt.Errorf(errorUrlUnknownScheme, urlobj.Scheme)
	}

	return urlobj, nil
}
