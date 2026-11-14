package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"time"
)

func CalculateMultipartFileMD5(file multipart.File) (string, error) {
	// 1. 创建 MD5 哈希计算器
	hash := md5.New()

	// 2. 分块读取文件内容并写入哈希计算器
	// 使用 io.Copy 自动处理缓冲区，适合大文件
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("读取上传文件内容失败: %w", err)
	}

	// 3. 计算 MD5 并转换为十六进制字符串
	md5Bytes := hash.Sum(nil)
	md5Str := hex.EncodeToString(md5Bytes)

	// 4. 重置文件指针到起始位置，方便后续操作（可选但建议）
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return md5Str, fmt.Errorf("重置文件指针失败: %w", err)
	}

	return md5Str, nil
}

// IsSameDay 判断传入的 time.Time 和 yyyy-MM-dd 格式的日期字符串是否为同一天
func IsSameDay(t time.Time, dateStr string) bool {
	// 解析日期字符串
	parsed, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return false
	}

	// 将两个时间都转换到同一时区再比较
	// 这里以 t 的时区为准，避免时区差异导致的误判
	y1, m1, d1 := t.Date()
	y2, m2, d2 := parsed.In(t.Location()).Date()

	return y1 == y2 && m1 == m2 && d1 == d2
}
