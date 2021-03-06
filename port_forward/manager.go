package port_forward

import (
	"strconv"
	"github.com/kataras/iris"
	"github.com/Joker/jade"
	"io/ioutil"
	"sync/atomic"
	"html/template"
	"gopkg.in/yaml.v2"
	"sort"
	"log"
)

type FilterRule struct {
	Src string
	Dst string
}

type PortForwardRule struct {
	ID int
	SrcAddr string
	DstAddr string
	IsHttp bool
	FilterRules []FilterRule
	inst *PortForward
}

type PortForwardManager struct {
	Rules map [int]PortForwardRule
	ID int64
	listAllTpl *template.Template
}

func loadJade(path string) (*template.Template, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	tpl, err := jade.Parse(path, string(buf))
	if err != nil {
		return nil, err
	}
	return template.New("html").Parse(tpl)
}

func NewManager() *PortForwardManager {
	var err error
	mgr := new(PortForwardManager)
	mgr.Rules = make(map[int]PortForwardRule)
	mgr.listAllTpl, err = loadJade("templates/main.jade")
	if err != nil {
		return nil
	}
	mgr.load()
	for k, v := range mgr.Rules {
		v.inst = NewPortForward(v.SrcAddr, v.DstAddr, v.IsHttp)
		if v.inst == nil {
			log.Printf("forward %v->%v [%v] failed.", v.SrcAddr, v.DstAddr, v.IsHttp)
			delete(mgr.Rules, k)
		}
	}
	return mgr
}

type ListVM struct {
	Rules []*PortForwardRule
}

func (mgr *PortForwardManager)ListAll(ctx iris.Context) {
	vm := ListVM{}
	for k := range mgr.Rules {
		v := mgr.Rules[k]
		vm.Rules = append(vm.Rules, &v)
	}
	sort.Slice(vm.Rules, func(i, j int) bool {
		return vm.Rules[i].SrcAddr < vm.Rules[j].SrcAddr
	})
	mgr.listAllTpl.Execute(ctx, vm)
}

func (mgr *PortForwardManager)Del(ctx iris.Context) {
	ruleID, err := strconv.Atoi(ctx.Params().Get("RuleID"))
	if err != nil {
		return
	}
	pf, ok := mgr.Rules[ruleID]
	if ok && pf.inst != nil {
		pf.inst.Close()
	}
	delete(mgr.Rules, ruleID)
	mgr.save()
	mgr.ListAll(ctx)
}

func (mgr *PortForwardManager)Add(ctx iris.Context) {
	inst := NewPortForward(ctx.URLParam("SrcAddr"), ctx.URLParam("DstAddr"), false)
	if inst != nil {
		id := int(atomic.AddInt64(&mgr.ID, 1))
		mgr.Rules[id] = PortForwardRule{
			ID: id,
			SrcAddr: ctx.URLParam("SrcAddr"),
			DstAddr: ctx.URLParam("DstAddr"),
			IsHttp: false,
			inst: inst,
		}
	}
	mgr.save()
	ctx.Redirect("/console")
}

func (mgr *PortForwardManager)AddHttp(ctx iris.Context) {
	inst := NewPortForward(ctx.URLParam("SrcAddr"), ctx.URLParam("DstAddr"), true)
	if inst != nil {
		id := int(atomic.AddInt64(&mgr.ID, 1))
		mgr.Rules[id] = PortForwardRule{
			ID: id,
			SrcAddr: ctx.URLParam("SrcAddr"),
			DstAddr: ctx.URLParam("DstAddr"),
			IsHttp: true,
			inst: inst,
		}
	}
	mgr.save()
	ctx.Redirect("/console")
}

func (mgr *PortForwardManager)load() {
	d, err := ioutil.ReadFile("rules.yml")
	if err != nil {
		log.Print(err)
	}
	err = yaml.Unmarshal(d, mgr)
	if err != nil {
		log.Print(err)
	}
}

func (mgr *PortForwardManager)save() {
	d, err := yaml.Marshal(&mgr)
	if err != nil {
		panic(err)
	}
	ioutil.WriteFile("rules.yml", d, 0666)
}