package httpError

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/armineyvazi/common.git/pkg/ports"
)

const (
	FieldHeaderOrigin = "Origin"
	FieldReferrer     = "Referer"
	FieldUserAgent    = "User-Agent"
)

type HttpErrorContext struct {
	Time       time.Time `json:"time"`
	IPAddress  []string  `json:"ip_address"`
	RequestURL string    `json:"requestURL"`
	Referrer   string    `json:"referrer"`
	UserAgent  string    `json:"user_agent"`
	RealDomain string    `json:"real_domain"`
	Status     int       `json:"status"`
}

func (e *httpErr) GetMetaData(c *ports.HttpContext) interface{} {
	ipAddresses := GetIPAddressList(c)
	requestURL := c.OriginalURL()
	userAgent := c.Get(FieldUserAgent)
	referrer := c.Get(FieldReferrer)
	realDomain := c.Get(FieldHeaderOrigin)
	return HttpErrorContext{
		Time:       time.Now(),
		Status:     e.GetHttpStatus(),
		IPAddress:  ipAddresses,
		RequestURL: requestURL,
		Referrer:   referrer,
		UserAgent:  userAgent,
		RealDomain: realDomain,
	}
}

func GetIPAddressList(c *ports.HttpContext) []string {
	var ipAddressList []string
	forwardedFor := c.Get(fiber.HeaderXForwardedFor)

	if forwardedFor == "" {
		return ipAddressList
	}

	ipAddresses := strings.Split(forwardedFor, ",")
	for _, ip := range ipAddresses {
		ipAddressList = append(ipAddressList, strings.TrimSpace(ip))
	}

	return ipAddressList
}
