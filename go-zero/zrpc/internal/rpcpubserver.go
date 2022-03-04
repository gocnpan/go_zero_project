package internal

import (
	"os"
	"strings"

	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/core/netx"
)

const (
	allEths  = "0.0.0.0"
	envPodIp = "POD_IP"
)

// NewRpcPubServer returns a Server.
// 在 etcd 注册服务，定时更新
// 开启 rpc 服务
func NewRpcPubServer(etcd discov.EtcdConf, listenOn string, opts ...ServerOption) (Server, error) {
	registerEtcd := func() error {
		pubListenOn := figureOutListenOn(listenOn) // 获取 host:port
		var pubOpts []discov.PubOption
		if etcd.HasAccount() { // 有账号密码的情况下，填入账号密码
			pubOpts = append(pubOpts, discov.WithPubEtcdAccount(etcd.User, etcd.Pass))
		}
		if etcd.HasTLS() { // 有 tls 配置时，填入
			pubOpts = append(pubOpts, discov.WithPubEtcdTLS(etcd.CertFile, etcd.CertKeyFile,
				etcd.CACertFile, etcd.InsecureSkipVerify))
		}
		// 配置和保持更新
		pubClient := discov.NewPublisher(etcd.Hosts, etcd.Key, pubListenOn, pubOpts...)
		return pubClient.KeepAlive()
	}
	server := keepAliveServer{
		registerEtcd: registerEtcd,
		Server:       NewRpcServer(listenOn, opts...), // 开启 RPC 服务
	}

	return server, nil
}

type keepAliveServer struct {
	registerEtcd func() error
	Server
}

func (ags keepAliveServer) Start(fn RegisterFn) error {
	if err := ags.registerEtcd(); err != nil {
		return err
	}

	return ags.Server.Start(fn)
}

func figureOutListenOn(listenOn string) string {
	fields := strings.Split(listenOn, ":")
	if len(fields) == 0 { // 纯端口？
		return listenOn
	}

	host := fields[0]
	if len(host) > 0 && host != allEths { // host正确，返回 host:port
		return listenOn
	}

	ip := os.Getenv(envPodIp)
	if len(ip) == 0 {
		ip = netx.InternalIp()
	}
	if len(ip) == 0 {
		return listenOn
	}

	return strings.Join(append([]string{ip}, fields[1:]...), ":")
}
