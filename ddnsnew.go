package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net"
    "net/http"
    "os"
    "regexp"
    "time"
)

const (
    interval = 1 * time.Minute // 每分钟更新一次
)

// 定义返回的JSON结构体
type ApiResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
}

// 获取公网IPv4地址
func getIPv4() (string, error) {
    url := "https://ddns.oray.com/checkip"
    client := &http.Client{}
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return "", err
    }
    req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Ubuntu 20.04; x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36")
    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    re := regexp.MustCompile(`(?:\d+\.){3}\d+`)
    matches := re.FindStringSubmatch(string(body))
    if len(matches) > 0 {
        return matches[0], nil
    }
    return "未找到IPv4地址", nil
}

// 获取公网IPv6地址
func getIPv6() string {
    addrs, err := net.InterfaceAddrs()
    if err != nil {
        return "未找到IPv6地址"
    }

    for _, addr := range addrs {
        ipNet, ok := addr.(*net.IPNet)
        if !ok || ipNet.IP.IsLoopback() {
            continue
        }
        if ipNet.IP.To4() == nil { // 判断是否为IPv6地址
            if ipNet.IP.IsGlobalUnicast() { // 确保是全局范围的IPv6地址
                // 检查是否以"240"开头
                if len(ipNet.IP.String()) >= 4 && ipNet.IP.String()[:3] == "240" {
                    return ipNet.IP.String()
                }
            }
        }
    }

    return "未找到IPv6地址"
}

// 更新DNS记录
func updateDNS(domain, token, addr string, apiURL string) {
    url := fmt.Sprintf("%s?domain=%s&token=%s&addr=%s", apiURL, domain, token, addr)
    resp, err := http.Get(url)
    if err != nil {
        fmt.Println("DNS 更新失败:", err)
        return
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Println("读取响应体失败:", err)
        return
    }

    // 解析JSON响应
    var apiResponse ApiResponse
    err = json.Unmarshal(body, &apiResponse)
    if err != nil {
        fmt.Println("解析JSON失败:", err)
        return
    }

    // 格式化输出
    if apiResponse.Success {
        fmt.Printf("DNS 更新成功: %s\n", apiResponse.Message)
    } else {
        fmt.Printf("DNS 更新失败: %s\n", apiResponse.Message)
    }
}

func main() {
    // 从环境变量中读取配置
    domain := os.Getenv("DOMAIN")
    token := os.Getenv("TOKEN")
    apiURL := os.Getenv("API_URL")

    // 检查是否缺少必要的环境变量
    if domain == "" {
        fmt.Println("错误: 环境变量 DOMAIN 未设置")
        return
    }
    if token == "" {
        fmt.Println("错误: 环境变量 TOKEN 未设置")
        return
    }
    if apiURL == "" {
        fmt.Println("错误: 环境变量 API_URL 未设置")
        return
    }

    for {
        // 获取公网IPv4地址
        ipv4, err := getIPv4()
        if err != nil {
            fmt.Println("获取IPv4地址出错:", err)
        } else {
            fmt.Printf("公网IPv4地址: %s\n", ipv4)
        }

        // 获取公网IPv6地址
        ipv6 := getIPv6()
        fmt.Printf("公网IPv6地址: %s\n", ipv6)

        // 如果IPv6和IPv4都有效，则优先使用IPv6更新DNS
        if ipv6 != "未找到IPv6地址" {
            updateDNS(domain, token, ipv6, apiURL)
        } else if ipv4 != "未找到IPv4地址" {
            updateDNS(domain, token, ipv4, apiURL)
        } else {
            fmt.Println("无法获取有效的公网IP地址")
        }

        // 等待指定的时间间隔
        time.Sleep(interval)
    }
}