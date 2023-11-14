package main

import (
	"bytes"
	"fmt"
	"log"
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

// addBackticks 将每行字符串包裹在反引号中
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

// removeEmptyLines 去除空行
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

// sendMessage 发送告警消息到企业微信bot机器人
func sendMessage(notification Notification) error {
	url := os.Getenv("WXWORK_WEBHOOK_BOT_URL")
	if url == "" {
		return fmt.Errorf("未设置 WXWORK_WEBHOOK_BOT_URL 环境变量")
	}

	_, _, data := processAlerts(notification.Alerts)

	message := createMessage(data)
	if err := postMessage(url, message); err != nil {
		return err
	}

	log.Println("消息发送成功")
	return nil
}

// processAlerts 处理告警并返回消息数据
func processAlerts(alerts []Alert) (int, int, string) {
	var firingCount, resolvedCount int
	var firingMsg, resolvedMsg, data string

	for _, item := range alerts {
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

	return firingCount, resolvedCount, data
}

// createMessage 创建要发送的消息内容
func createMessage(data string) []byte {
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
		log.Println("模板解析错误:", err)
		return nil
	}

	var messageBuffer bytes.Buffer
	if err := tmpl.Execute(&messageBuffer, map[string]interface{}{"Data": data}); err != nil {
		log.Println("模板渲染错误:", err)
		return nil
	}

	return messageBuffer.Bytes()
}

// postMessage 发送HTTP POST请求
func postMessage(url string, message []byte) error {
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(message))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("HTTP请求失败: %s", resp.Status)
		return fmt.Errorf("HTTP请求失败: %s", resp.Status)
	}

	return nil
}

// 主函数
func main() {
	r := gin.Default()

	r.LoadHTMLGlob("templates/*")

	r.GET("/*filepath", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

	r.POST("/*filepath", func(c *gin.Context) {
		var notification Notification
		if err := c.BindJSON(&notification); err != nil {
			log.Printf("JSON 解析错误: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}

		if err := sendMessage(notification); err != nil {
			log.Printf("发送消息失败: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Message delivery failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Notification sent"})
	})

	r.Run(":8000")
}
