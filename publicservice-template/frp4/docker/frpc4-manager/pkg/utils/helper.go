package utils

import (
	"bytes"
	"io"
	"math/rand"
	"net/http"

	"github.com/BurntSushi/toml"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog"
)

const (
	FrpcConfigPath  = "http://127.0.0.1:7400/api/config"
	FrpcReloadPath  = "http://127.0.0.1:7400/api/reload"
	RemotePortStart = 30000
	RemotePortRange = 30000
)

type FrpcConfig struct {
	ServerAddr string    `toml:"serverAddr"`
	ServerPort int       `toml:"serverPort"`
	WebServer  WebServer `toml:"webServer"`
	Auth       Auth      `json:"auth"`
	Proxies    []Proxy   `toml:"proxies"`
}

type WebServer struct {
	Addr string `toml:"addr"`
	Port int    `toml:"port"`
}

type Auth struct {
	Method string `toml:"method"`
	Token  string `toml:"token"`
}

type Proxy struct {
	Name       string `toml:"name"`
	Type       string `toml:"type"`
	LocalIP    string `toml:"localIP"`
	LocalPort  int    `toml:"localPort"`
	RemotePort int    `toml:"remotePort"`
}

func AddOrUpdateProxy(svc *corev1.Service) (string, int, bool, error) {
	cfg, err := GetFrpcConfig()
	if err != nil {
		klog.Errorf("Get frpc config error:%v", err)
		return "", -1, false, err
	}

	proxies := map[string]Proxy{}
	portsSet := sets.New[int]()
	for _, p := range cfg.Proxies {
		proxies[p.Name] = p
		portsSet.Insert(p.RemotePort)
	}

	changed := false
	curr, ok := proxies[svc.Name]
	if !ok {
		changed = true
		proxies[svc.Name] = Proxy{
			Name:       svc.Name,
			Type:       "tcp",
			LocalIP:    svc.Name + "." + svc.Namespace,
			LocalPort:  int(svc.Spec.Ports[0].Port),
			RemotePort: GetRandomPort(portsSet),
		}
	} else {
		if curr.LocalIP != svc.Name+"."+svc.Namespace || curr.LocalPort != int(svc.Spec.Ports[0].Port) {
			changed = true
			curr.LocalIP = svc.Name + "." + svc.Namespace
			curr.LocalPort = int(svc.Spec.Ports[0].Port)
			proxies[svc.Name] = curr
		}
	}
	if !changed {
		return "", -1, false, nil
	}

	proxyList := []Proxy{}
	for _, p := range proxies {
		proxyList = append(proxyList, p)
	}
	cfg.Proxies = proxyList
	err = UpdateFrpcConfig(cfg)
	if err != nil {
		klog.Errorf("Update frpc config error:%v", err)
		return "", -1, false, err
	}
	return cfg.ServerAddr, proxies[svc.Name].RemotePort, true, nil
}

func DeleteProxy(svc *corev1.Service) error {
	cfg, err := GetFrpcConfig()
	if err != nil {
		klog.Errorf("Get frpc config error:%v", err)
		return err
	}

	changed := false
	for i, p := range cfg.Proxies {
		if p.Name == svc.Name {
			cfg.Proxies = append(cfg.Proxies[:i], cfg.Proxies[i+1:]...)
			changed = true
			break
		}
	}
	if !changed {
		return nil
	}
	err = UpdateFrpcConfig(cfg)
	if err != nil {
		klog.Errorf("Update frpc config error:%v", err)
		return err
	}
	return nil
}

func GetRandomPort(portsSet sets.Set[int]) int {
	var ret int
	for {
		ret = rand.Intn(RemotePortStart) + RemotePortRange
		if portsSet.Has(ret) {
			continue
		}
		break
	}
	portsSet.Insert(ret)
	return ret
}

func UpdateFrpcConfig(cfg *FrpcConfig) error {
	var encodeData = new(bytes.Buffer)
	err := toml.NewEncoder(encodeData).Encode(cfg)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, FrpcConfigPath, encodeData)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain;charset=UTF-8")

	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		return err
	}

	_, err = http.Get(FrpcReloadPath)
	if err != nil {
		return err
	}

	return nil
}

func GetFrpcConfig() (*FrpcConfig, error) {
	resp, err := http.Get(FrpcConfigPath)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	configData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var config FrpcConfig
	_, err = toml.Decode(string(configData), &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
