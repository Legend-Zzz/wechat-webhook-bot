package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"strings"
	"text/template"

	"github.com/gin-gonic/gin"
)

type Alert struct {
	Status      string            `json:"status"`
	Annotations map[string]string `json:"annotations"`
}

type Notification struct {
	Alerts []Alert `json:"alerts"`
}

// 字符串每行前后增加符号`
func addBackticks(s string) string {
	lines := strings.Split(s, "\n")
	var result []string
	for _, line := range lines {
		if line != "" {
			result = append(result, "`"+line+"`")
		}
	}
	return strings.Join(result, "\n")
}

// 字符串去除空行
func removeEmptyLines(s string) string {
	lines := strings.Split(s, "\n")
	var result []string
	for _, line := range lines {
		if line != "" {
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n")
}

// alertmanager发送webhook(post请求)，经过本程序处理，再将处理后的数据post到企业微信bot机器人接口，进而发送告警通知
func SendMessage(notification Notification) {
	url := os.Getenv("WXWORK_WEBHOOK_BOT_URL")
	if url == "" {
		fmt.Println("未设置 WXWORK_WEBHOOK_BOT_URL 环境变量")
		return
	}

	var (
		firingCount   int
		resolvedCount int
		firingMsg     string
		resolvedMsg   string
		data          string
	)

	for _, item := range notification.Alerts {
		if item.Status == "firing" {
			firingCount++
			firingMsg += item.Annotations["description"] + "\n"
		} else if item.Status == "resolved" {
			resolvedCount++
			resolvedMsg += item.Annotations["description"] + "\n"
		}
	}

	firingMsg = addBackticks(firingMsg)
	resolvedMsg = removeEmptyLines(resolvedMsg)

	if firingCount > 0 && resolvedCount > 0 {
		data = fmt.Sprintf("`[%d]  未恢复的告警`\n%s\n<font color=\\\"info\\\">[%d]  已恢复的告警\n%s</font>", firingCount, firingMsg, resolvedCount, resolvedMsg)
	} else if firingCount > 0 && resolvedCount == 0 {
		data = fmt.Sprintf("`[%d]  未恢复的告警`\n%s", firingCount, firingMsg)
	} else if firingCount == 0 && resolvedCount > 0 {
		data = fmt.Sprintf("<font color=\\\"info\\\">[%d]  已恢复的告警\n%s</font>", resolvedCount, resolvedMsg)
	} else {
		data = "error, no data"
	}

	// 发送告警消息
	messageTemplate := `
	{
		"msgtype": "markdown",
		"markdown": {
			"content": "{{ .Data }}"
		}
	}
	`

	tmpl, err := template.New("message").Parse(messageTemplate)
	if err != nil {
		fmt.Println("模板解析错误:", err)
		return
	}

	var messageBuffer bytes.Buffer
	if err := tmpl.Execute(&messageBuffer, map[string]interface{}{"Data": data}); err != nil {
		fmt.Println("模板渲染错误:", err)
		return
	}

	headers := http.Header{}
	headers.Set("Content-Type", "raw")

	reqBody := bytes.NewBuffer(messageBuffer.Bytes())

	resp, err := http.Post(url, headers.Get("Content-Type"), reqBody)
	if err != nil {
		fmt.Println("POST 请求发生错误:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("HTTP 状态码:", resp.Status)
}

// GET请求返回主页，POST请求调用send_alert方法
func main() {
	r := gin.Default()

	r.LoadHTMLGlob("templates/*")

	r.GET("/*filepath", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

	r.POST("/*filepath", func(c *gin.Context) {
		var notification Notification
		if err := c.BindJSON(&notification); err != nil {
			fmt.Println("JSON 解析错误:", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}
		SendMessage(notification)
		c.JSON(http.StatusOK, gin.H{"message": "Notification sent"})
	})

	r.Run(":8000")
}
