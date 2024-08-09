package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/AdguardTeam/dnsproxy/proxy"
	"github.com/AdguardTeam/golibs/logutil/slogutil"
	"github.com/kardianos/service"
	"log"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	serviceParam string
	confFile     string
)

const configName = "proxy.yaml"

func init() {
	flag.StringVar(&serviceParam, "service", "", "control the system service.")
	flag.StringVar(&confFile, "conf", configName, "define the config file path.")
}

func main() {
	flag.Parse()

	// 获取安装路径
	installationPath := GetCurrentAbPath()

	serviceConfig := &service.Config{
		Name:        "Windows-DNS-Proxy-Service",
		DisplayName: "Windows DNS Proxy Service",
		Description: "The Windows DNS Proxy Service is a DNS proxy service that supports protocols such as TCP, UDP, HTTPDNS, DoT, and DoH",
		Option: map[string]interface{}{
			// 开机自启动
			"DelayedAutoStart": true,
		},
		// 设置配置文件路径
		Arguments: []string{"-conf", filepath.Join(installationPath, configName)},
	}
	dnsProxyService, err := NewDNSProxyService(confFile)
	if err != nil {
		panic(fmt.Errorf("new dns proxy service error [%s]", err.Error()))
	}
	s, err := service.New(dnsProxyService, serviceConfig)
	if err != nil {
		panic(fmt.Errorf("create service error [%s]", err.Error()))
	}

	if serviceParam != "" {
		if err := service.Control(s, serviceParam); err != nil {
			panic(err)
		}
		return
	}

	if err := s.Run(); err != nil {
		panic(err)
	}
}

type DNSProxyService struct {
	options  *Options
	dnsProxy *proxy.Proxy
	log      *slog.Logger
	ctx      context.Context
}

func NewDNSProxyService(confFile string) (*DNSProxyService, error) {
	// 读取配置
	opts := &Options{}
	err := parseConfigFile(opts, confFile)
	if err != nil {
		return nil, fmt.Errorf(
			"parsing config file %s: %w",
			confFile,
			err,
		)
	}

	// 设置log
	logOutput := os.Stdout
	if opts.LogOutput != "" {
		logOutput, err = os.OpenFile(opts.LogOutput, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
		if err != nil {
			return nil, fmt.Errorf("cannot create a log file: %s", err)
		}
	}
	l := slogutil.New(&slogutil.Config{
		Output:       logOutput,
		Format:       slogutil.FormatDefault,
		AddTimestamp: true,
		Verbose:      opts.Verbose,
	})

	return &DNSProxyService{
		options: opts,
		log:     l,
		ctx:     context.Background(),
	}, nil
}

func (s *DNSProxyService) Start(service service.Service) error {
	// Prepare the proxy server and its configuration.
	conf, err := createProxyConfig(s.ctx, s.log, s.options)
	if err != nil {
		return fmt.Errorf("configuring proxy: %w", err)
	}

	dnsProxy, err := proxy.New(conf)
	if err != nil {
		return fmt.Errorf("creating proxy: %w", err)
	}
	s.dnsProxy = dnsProxy

	// Add extra handler if needed.
	if s.options.IPv6Disabled {
		ipv6Config := ipv6Configuration{
			logger:       s.log,
			ipv6Disabled: s.options.IPv6Disabled,
		}
		dnsProxy.RequestHandler = ipv6Config.handleDNSRequest
	}

	// Start the proxy server.
	err = dnsProxy.Start(s.ctx)
	if err != nil {
		return fmt.Errorf("starting dnsproxy: %w", err)
	}

	return nil
}

func (s *DNSProxyService) Stop(service service.Service) error {
	if err := s.dnsProxy.Shutdown(s.ctx); err != nil {
		return fmt.Errorf("stopping dnsproxy: %w", err)
	}
	return nil
}

// 获取安装的路径，对于window使用pwd得到的不一定是安装路径，而是C:\Window\System32
func GetCurrentAbPath() string {
	dir := getCurrentAbPathByExecutable()
	tmpDir, _ := filepath.EvalSymlinks(os.TempDir())
	if tmpDir != "" && strings.Contains(dir, tmpDir) {
		return getCurrentAbPathByCaller()
	}
	return dir
}

// 获取当前执行文件绝对路径
func getCurrentAbPathByExecutable() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	res, _ := filepath.EvalSymlinks(filepath.Dir(exePath))
	return res
}

// 获取当前执行文件绝对路径
func getCurrentAbPathByCaller() string {
	var abPath string
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		abPath = path.Dir(filename)
	}
	return abPath
}
