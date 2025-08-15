package xmlrpc

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	url      string
	username string
	password string
	client   *http.Client
}

func NewClient(host string, port int, username, password string) (*Client, error) {
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "http://" + host
	}

	baseURL, err := url.Parse(host)
	if err != nil {
		return nil, fmt.Errorf("invalid host URL: %v", err)
	}

	// 设置认证信息
	if username != "" || password != "" {
		baseURL.User = url.UserPassword(username, password)
	}

	// 添加端口和RPC2路径
	baseURL.Host = fmt.Sprintf("%s:%d", baseURL.Hostname(), port)
	baseURL.Path = "/RPC2"

	return &Client{
		url:      baseURL.String(),
		username: username,
		password: password,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}, nil
}

func (c *Client) Call(method string, args []interface{}) (interface{}, error) {
	// 构建XML-RPC请求
	request := fmt.Sprintf(`<?xml version="1.0"?>
	<methodCall>
		<methodName>%s</methodName>
		<params>%s</params>
	</methodCall>`, method, c.encodeParams(args))

	// 发送请求
	resp, err := c.client.Post(c.url, "text/xml", bytes.NewBufferString(request))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var result struct {
		Params []struct {
			Value interface{} `xml:"value"`
		} `xml:"params>param"`
		Fault *struct {
			Value struct {
				Struct struct {
					FaultCode   int    `xml:"member>value>int"`
					FaultString string `xml:"member>value>string"`
				} `xml:"struct"`
			} `xml:"value"`
		} `xml:"fault"`
	}

	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	// 检查错误
	if result.Fault != nil {
		return nil, fmt.Errorf("XML-RPC fault: [%d] %s",
			result.Fault.Value.Struct.FaultCode,
			result.Fault.Value.Struct.FaultString)
	}

	// 返回结果
	if len(result.Params) > 0 {
		return result.Params[0].Value, nil
	}
	return nil, nil
}

func (c *Client) encodeParams(args []interface{}) string {
	if len(args) == 0 {
		return ""
	}

	params := make([]string, len(args))
	for i, arg := range args {
		params[i] = fmt.Sprintf("<param><value>%s</value></param>", c.encodeValue(arg))
	}

	return strings.Join(params, "")
}

func (c *Client) encodeValue(v interface{}) string {
	switch v := v.(type) {
	case string:
		return fmt.Sprintf("<string>%s</string>", v)
	case int, int32, int64:
		return fmt.Sprintf("<int>%d</int>", v)
	case float32, float64:
		return fmt.Sprintf("<double>%f</double>", v)
	case bool:
		return fmt.Sprintf("<boolean>%d</boolean>", map[bool]int{false: 0, true: 1}[v])
	default:
		return fmt.Sprintf("<string>%v</string>", v)
	}
}
