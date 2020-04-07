package sd

import (
	"context"
	"fmt"
	"trinitygo/utils"

	"github.com/coreos/etcd/clientv3"
	etcdnaming "github.com/coreos/etcd/clientv3/naming"
	"google.golang.org/grpc/naming"
)

// ServiceMesh interface
type ServiceMesh interface {
	GetClient() interface{}
	RegService(projectName string, projectVersion string, serviceIP string, servicePort int, Tags []string) error
	DeRegService(projectName string, projectVersion string, serviceIP string, servicePort int) error
}

// ServiceMeshEtcdImpl consul register
type ServiceMeshEtcdImpl struct {
	// config
	Address string // consul address
	Port    int

	// runtime
	client *clientv3.Client
}

// NewEtcdRegister New consul register
func NewEtcdRegister(address string, port int) (ServiceMesh, error) {
	s := &ServiceMeshEtcdImpl{
		Address: address,
		Port:    port,
	}

	cli, err := clientv3.NewFromURL(fmt.Sprintf("http://%v:%v", s.Address, s.Port))

	if err != nil {
		return nil, err
	}
	s.client = cli
	return s, nil
}

// GetClient get etcd client
func (s *ServiceMeshEtcdImpl) GetClient() interface{} {
	return s.client
}

// RegService register etcd service
func (s *ServiceMeshEtcdImpl) RegService(projectName string, projectVersion string, serviceIP string, servicePort int, Tags []string) error {
	r := &etcdnaming.GRPCResolver{Client: s.client}
	err := r.Update(context.TODO(), utils.GetServiceName(projectName), naming.Update{Op: naming.Add, Addr: fmt.Sprintf("%v:%v", serviceIP, servicePort), Metadata: fmt.Sprintf("%v", Tags)})
	if err != nil {
		return err
	}
	return nil
}

// DeRegService deregister service
func (s *ServiceMeshEtcdImpl) DeRegService(projectName string, projectVersion string, serviceIP string, servicePort int) error {
	r := &etcdnaming.GRPCResolver{Client: s.client}
	err := r.Update(context.TODO(), utils.GetServiceName(projectName), naming.Update{Op: naming.Delete, Addr: fmt.Sprintf("%v:%v", serviceIP, servicePort)})
	if err != nil {
		return err
	}
	return nil
}
