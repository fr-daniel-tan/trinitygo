package application

import (
	"fmt"
	"reflect"
	"sync"

	"log"

	"github.com/gin-gonic/gin"
	"github.com/kataras/golog"
)

// ControllerPool service pool
// if grpc string is the full method of method
// if http os the GET@/ping/:id
// need to filter controllerFuncMap to filter funcname
type ControllerPool struct {
	mu                   sync.RWMutex
	containerMap         map[string]reflect.Type
	controllerMap        []string
	controllerFuncMap    map[string]string
	controllerValidators map[string][]Validator
}

// NewControllerPool new pool with init map
func NewControllerPool() *ControllerPool {
	result := new(ControllerPool)
	result.mu.Lock()
	defer result.mu.Unlock()
	result.containerMap = make(map[string]reflect.Type)
	result.controllerFuncMap = make(map[string]string)
	result.controllerValidators = make(map[string][]Validator)
	return result

}

// NewController add new service
func (s *ControllerPool) NewController(controllerType string, container reflect.Type) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.containerMap[controllerType] = container
	s.controllerMap = append(s.controllerMap, controllerType)
}

// NewControllerFunc register funcname for controllertype
func (s *ControllerPool) NewControllerFunc(controllerType string, funcName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.controllerFuncMap[controllerType] = funcName
}

// NewControllerValidators register funcname for controllertype
func (s *ControllerPool) NewControllerValidators(controllerType string, validator ...Validator) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.controllerValidators[controllerType] = validator
}

// GetControllerMap get controller map
func (s *ControllerPool) GetControllerMap() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.controllerMap
}

// ControllerFuncSelfCheck self check http request registered func exist or not
func (s *ControllerPool) ControllerFuncSelfCheck(contailerPool *ContainerPool, isLog bool, logger *golog.Logger) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for controllerName, containerName := range s.containerMap {
		funcName, funcExist := s.controllerFuncMap[controllerName]
		if funcName == "" || !funcExist {
			// func not exist
			logger.Fatalf("booting self func checking controller %v , no func registered , self check failed ...", controllerName)
		}
		containerPool, ok := contailerPool.poolMap[containerName]
		if !ok {
			// func not exist
			logger.Fatalf("booting self func checking controller %v , no container %v registered , self check failed ...", controllerName, containerName)
		}
		container := containerPool.Get()
		defer containerPool.Put(container)
		_, funcImpled := reflect.TypeOf(container).MethodByName(funcName)
		if !funcImpled {
			log.Fatalf("booting self func checking controller %v , func %v not registered , self check failed ...", controllerName, funcName)
		}
		if isLog {
			logger.Infof("booting self func checking controller %v , func %v checked ", controllerName, funcName)
		}
	}

	return
}

// GetController from pool
func (s *ControllerPool) GetController(controllerName string, tctx Context, app Application, c *gin.Context) (interface{}, map[reflect.Type]interface{}) {
	s.mu.RLock()
	containerName, ok := s.containerMap[controllerName]
	s.mu.RUnlock()
	if !ok {
		panic(fmt.Sprintf("unknown controller name : %v", controllerName))
	}
	return app.ContainerPool().GetContainer(containerName, tctx, app, c)
}

// GetControllerFuncName get controller func name
func (s *ControllerPool) GetControllerFuncName(controllerName string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.controllerFuncMap) == 0 {
		return "", false
	}
	funcName, ok := s.controllerFuncMap[controllerName]
	return funcName, ok

}

// GetControllerValidators get controller func name
func (s *ControllerPool) GetControllerValidators(controllerName string) []Validator {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.controllerValidators) == 0 {
		return nil
	}
	validators, ok := s.controllerValidators[controllerName]
	if !ok {
		return nil
	}
	return validators

}
