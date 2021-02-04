package drivers

import (
	"github.com/OpenStars/EtcdBackendService/StringBigsetService"
	"github.com/OpenStars/GoEndpointManager/GoEndpointBackendManager"
	"sync"
)


var(
	bigsetClient StringBigsetService.StringBigsetServiceIf
	bigsetOnce sync.Once
)

func initBigset() {
	bigsetClient = StringBigsetService.NewStringBigsetServiceModel("blockchain",
		[]string{"127.0.0.1:2379"},
		GoEndpointBackendManager.EndPoint{
			Host:      "127.0.0.1",
			Port:      "18990",
			ServiceID: "blockchain",
		},
	)
}

func GetBigsetClient() StringBigsetService.StringBigsetServiceIf {
	bigsetOnce.Do(initBigset)

	return bigsetClient
}
