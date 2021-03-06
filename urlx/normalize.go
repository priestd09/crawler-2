// Package urlx implements some URL utility functions.
package urlx

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"path"
	"regexp"
	"strings"
	"unicode/utf8"

	"golang.org/x/net/idna"
)

var domainRegexp = regexp.MustCompile(
	`^([a-zA-Z0-9-]{1,63}\.)+[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9]$`,
)

func validateHost(host string) (string, error) {
	lower := strings.ToLower(host)
	if domainRegexp.MatchString(lower) || lower == "localhost" ||
		net.ParseIP(lower) != nil {
		return lower, nil
	}
	// The URL will be used by net/http, where IDNA is not supported.
	if punycode, err := idna.ToASCII(host); err != nil {
		return "", err
	} else if domainRegexp.MatchString(punycode) {
		return punycode, nil
	}
	return "", errors.New("not valid domain name or IP address")
}

func Normalize(u *url.URL) error {
	if !utf8.ValidString(u.String()) {
		return fmt.Errorf("normalize URL: invalid UTF-8 string: %q", u.String())
	}

	u.Scheme = strings.ToLower(u.Scheme)
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("normalize URL: unsupported scheme: %v", u.Scheme)
	}
	host, port, err := net.SplitHostPort(u.Host)
	if err != nil { // missing port
		host, port = u.Host, ""
	}
	if host == "" {
		return errors.New("normalize URL: empty host")
	} else if v, err := validateHost(host); err != nil {
		return fmt.Errorf("normalize URL: invalid host %q: %v", host, err)
	} else {
		u.Host = v
	}

	if (u.Scheme == "http" && port == "80") ||
		(u.Scheme == "https" && port == "443") {
		port = ""
	}
	if port != "" {
		u.Host = net.JoinHostPort(u.Host, port)
	}

	clean := func(pth string) string {
		p := path.Clean(pth)
		if p == "." {
			p = ""
		} else if strings.HasSuffix(pth, "/") && !strings.HasSuffix(p, "/") {
			p += "/"
		}
		return p
	}
	u.Path = clean(u.Path)
	u.RawPath = clean(u.RawPath)
	u.Fragment = ""
	return nil
}
