package master

import (
	"fmt"
	"time"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/expapi"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
	thirdpartyresourceetcd "k8s.io/kubernetes/pkg/registry/thirdpartyresource/etcd"
	"k8s.io/kubernetes/pkg/registry/thirdpartyresourcedata"
	"k8s.io/kubernetes/pkg/util"
	
	"github.com/golang/glog"
)

type APIInterface interface {
	RemoveAPI(path string)
	InstallThirdPartyAPI(rsrc *expapi.ThirdPartyResource) error
	HasAPI(rsrc *expapi.ThirdPartyResource) (bool, error)
	ListThirdPartyAPIs() []string
}

type ThirdPartyController struct {
	master                     APIInterface
	thirdPartyResourceRegistry *thirdpartyresourceetcd.REST
}

func (t *ThirdPartyController) SyncOneResource(rsrc *expapi.ThirdPartyResource) error {
	hasAPI, err := t.master.HasAPI(rsrc)
	if err != nil {
		return err
	}
	if !hasAPI {
		return t.master.InstallThirdPartyAPI(rsrc)
	}
	return nil
}

func (t *ThirdPartyController) SyncLoop() error {
	list, err := t.thirdPartyResourceRegistry.List(api.NewDefaultContext(), labels.Everything(), fields.Everything())
	if err != nil {
		return err
	}
	existing := util.StringSet{}
	switch list := list.(type) {
	case *expapi.ThirdPartyResourceList:
		for ix := range list.Items {
			item := &list.Items[ix]
			_, group, err := thirdpartyresourcedata.ExtractApiGroupAndKind(item)
			if err != nil {
				return err
			}
			existing.Insert(makeThirdPartyPath(group))
			if err := t.SyncOneResource(item); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("expected a *ThirdPartyResourceList, got %#v", list)
	}
	installed := t.master.ListThirdPartyAPIs()
	for _, installedAPI := range installed {
		if !existing.Has(installedAPI) {
			t.master.RemoveAPI(installedAPI)
		}
	}
	
	return nil
}

func (t *ThirdPartyController) Sync() {
	ticker := time.Tick(10 * time.Second)
	for {
		if err := t.SyncLoop(); err != nil {
			glog.Errorf("third party api sync failed: %v", err)
		}
		<- ticker
	}
}
