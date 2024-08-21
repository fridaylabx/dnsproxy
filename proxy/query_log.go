package proxy

import (
	"bytes"
	"fmt"
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"strings"
	"time"
)

type QueryLog struct {
	Msg *dns.Msg
	// 客户端源IP
	SourceIP string
	// 客户端源端口
	SourcePort string
	// 查询时间
	Cost time.Duration
	// 是否命中缓存
	Hit bool
	// XForwardedFor
	XForwardedFor string
	// ECS
	ECS string
	// 协议
	Proto string
}

type DNSQueryLogFormatter struct{}

func (m *DNSQueryLogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}
	timestamp := entry.Time.Format("2006-01-02 15:04:05")
	b.WriteString(fmt.Sprintf("%s %s\n", timestamp, entry.Message))
	return b.Bytes(), nil
}

func FormatQueryLog(queryLog *QueryLog) string {
	question := queryLog.Msg.Question[0].Name
	queryType := dns.TypeToString[queryLog.Msg.Question[0].Qtype]
	rCode := dns.RcodeToString[queryLog.Msg.Rcode]
	hitCache := "F"
	if queryLog.Hit {
		hitCache = "T"
	}
	//xForwardedFor := "N/A"
	//if queryLog.XForwardedFor != "" {
	//	xForwardedFor = queryLog.XForwardedFor
	//}
	ecs := "N/A"
	if queryLog.ECS != "" {
		ecs = queryLog.ECS
	}
	proto := "N/A"
	switch queryLog.Proto {
	case string(ProtoHTTPS):
		proto = "DOH"
	case string(ProtoTLS):
		proto = "DOT"
	default:
		proto = strings.ToUpper(queryLog.Proto)
	}
	var allAnswer []string
	if len(queryLog.Msg.Answer) > 0 {
		for _, r := range queryLog.Msg.Answer {
			if r != nil {
				fields := strings.SplitN(r.String(), "\t", 5)
				if fields[3] != queryType {
					continue
				}
				allAnswer = append(allAnswer, fields[4])
			}
		}
	}
	answer := "N/A"
	if len(allAnswer) != 0 {
		answer = strings.Join(allAnswer, ";")
	}
	return fmt.Sprintf("%s %s %s %s %s %s %s %s %s %dµs",
		proto,
		queryLog.SourceIP,
		queryLog.SourcePort,
		question,
		queryType,
		rCode,
		answer,
		hitCache,
		ecs,
		queryLog.Cost.Microseconds(),
	)
}

func SetQueryLogInfo(enable bool, dnsLogPath string) *logrus.Logger {
	logger := logrus.New()
	if enable && dnsLogPath != "" {
		l := &lumberjack.Logger{
			Filename:   dnsLogPath,
			MaxSize:    50,
			MaxBackups: 1,
		}
		logger.SetFormatter(&DNSQueryLogFormatter{})
		logger.SetOutput(l)
	}
	return logger
}
