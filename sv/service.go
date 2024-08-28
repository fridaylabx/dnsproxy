package sv

import (
	"context"
	"fmt"
	"github.com/AdguardTeam/golibs/logutil/slogutil"
	"github.com/fridaylabx/dnsproxy/proxy"
	"github.com/kardianos/service"
	"gopkg.in/natefinch/lumberjack.v2"
	"log"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

type DNSProxyService struct {
	options  *Options
	dnsProxy *proxy.Proxy
	log      *slog.Logger
	ctx      context.Context
}

func NewDNSProxyService(confFile string) (*DNSProxyService, error) {
	opts := &Options{}
	err := parseConfigFile(opts, confFile)
	if err != nil {
		return nil, fmt.Errorf(
			"parsing config file %s: %w",
			confFile,
			err,
		)
	}

	if runtime.GOOS == "windows" {
		if opts.LogOutput != "" {
			opts.LogOutput = filepath.Join(GetCurrentAbPath(), opts.LogOutput)
		}
		if opts.TLSKeyPath != "" {
			opts.TLSKeyPath = filepath.Join(GetCurrentAbPath(), opts.TLSKeyPath)
		}
		if opts.TLSCertPath != "" {
			opts.TLSCertPath = filepath.Join(GetCurrentAbPath(), opts.TLSCertPath)
		}
		if opts.QueryLog && opts.QueryLogPath != "" {
			opts.QueryLogPath = filepath.Join(GetCurrentAbPath(), opts.QueryLogPath)
		}
	}

	// 设置log
	logOutput := os.Stdout
	l := slogutil.New(&slogutil.Config{
		Output:       logOutput,
		Format:       slogutil.FormatDefault,
		AddTimestamp: true,
		Verbose:      opts.Verbose,
	})

	if opts.LogOutput != "" {
		log.SetOutput(&lumberjack.Logger{
			Filename:   opts.LogOutput,
			MaxSize:    30,
			MaxBackups: 5,
		})
	}

	return &DNSProxyService{
		options: opts,
		log:     l,
		ctx:     context.Background(),
	}, nil
}

func (s *DNSProxyService) Start(service service.Service) error {
	// Prepare the proxy server and its configuration.
	conf, err := CreateProxyConfig(s.ctx, s.log, s.options)
	if err != nil {
		return fmt.Errorf("configuring proxy: %w", err)
	}

	dnsProxy, err := proxy.New(conf)
	if err != nil {
		return fmt.Errorf("creating proxy: %w", err)
	}
	dnsProxy.QueryLogChan = make(chan *proxy.QueryLog, 100000)
	s.dnsProxy = dnsProxy

	// Add extra handler if needed.
	if s.options.IPv6Disabled {
		ipv6Config := Ipv6Configuration{
			Logger:       s.log,
			Ipv6Disabled: s.options.IPv6Disabled,
		}
		dnsProxy.RequestHandler = ipv6Config.HandleDNSRequest
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
