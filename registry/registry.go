package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/gfzwh/gfz/zzlog"

	"github.com/bilibili/discovery/naming"
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

type balance struct {
	weight   int64
	cons     int64
	requests int64
}

func getBalance(meta map[string]string) balance {
	var bl balance
	if _, ok := meta["weight"]; ok {
		bl.weight, _ = strconv.ParseInt(meta["weight"], 10, 64)
	}

	if _, ok := meta["cons"]; ok {
		bl.cons, _ = strconv.ParseInt(meta["cons"], 10, 64)
	}

	if _, ok := meta["requests"]; ok {
		bl.weight, _ = strconv.ParseInt(meta["requests"], 10, 64)
	}

	return bl
}

func (b balance) balancing() int64 {
	return (15 - b.weight) + b.requests + b.cons
}

type Instance []*naming.Instance

func (p Instance) Len() int {
	return len(p)
}

func (p Instance) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p Instance) Less(i, j int) bool {
	t1 := getBalance(p[i].Metadata)
	t2 := getBalance(p[j].Metadata)

	return t1.balancing() < t2.balancing()
}

type Registry struct {
	opts options
}

func NewRegistry(opts ...Options) *Registry {
	v := &Registry{}

	for _, o := range opts {
		o(&v.opts)
	}

	return v
}

func (r *Registry) Url() string {
	return r.opts.url
}

func (r *Registry) Nodes() []string {
	return r.opts.nodes
}
func (r *Registry) Region() string {
	return r.opts.region
}
func (r *Registry) Zone() string {
	return r.opts.zone
}
func (r *Registry) Env() string {
	return r.opts.env
}
func (r *Registry) Host() string {
	return r.opts.host
}

func (r *Registry) getNode(nodes []*naming.Instance) (node *naming.Instance) {
	if 0 == len(nodes) {
		return
	}

	sort.Sort(Instance(nodes))
	node = nodes[0]
	return
}

// 获取节点信息
//
// @param	svrname	需要获取的服务信息
// @param	zone	获取哪个区的节点
func (r *Registry) GetNodeInfo(svrname, zone, env, host string) (addr string, err error) {
	// env=prod&hostname=fgz-discovery 这两个变量是discovery中的env信息
	url := "%s/discovery/polls?appid=infra.discovery&appid=%s&env=dev-0.0.1&hostname=fgz-discovery&latest_timestamp=%d&latest_timestamp=0"
	url = fmt.Sprintf(url, r.opts.url, svrname, time.Now().UnixNano()-1000)

	zzlog.Debugw("discovery address", zap.String("url", url))
	response, err := http.Get(url)
	if err != nil {
		return
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return
	}

	res := new(struct {
		Code int                              `json:"code"`
		Data map[string]*naming.InstancesInfo `json:"data"`
	})
	err = json.Unmarshal(body, res)
	if nil != err {
		return
	}

	nodes := res.Data[svrname].Instances[zone]
	node := r.getNode(nodes)
	if nil == node {
		err = errors.New("GetNode node is nil!")

		return
	}

	// 获取节点连接信息
	addr = node.Addrs[0]
	aindex := strings.Index(addr, "tcp://")
	if 0 > aindex {
		addr = ""
		err = errors.New("Node not support protocol!")

		return
	}

	data, _ := json.Marshal(node)
	zzlog.Debugw("node info", zap.String("nodes", string(data)))
	addr = addr[(aindex + len("tcp://")):]
	return
}
