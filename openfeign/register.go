package openfeign

import "reflect"

type FeignClient interface {
	Domain() string
}

var feignClients = make([]FeignClient, 0)

func RegisterFeignClient(feignClient FeignClient) {
	if reflect.TypeOf(feignClient).Kind() != reflect.Ptr {
		panic("feignClient must be a pointer")
	}
	feignClients = append(feignClients, feignClient)
}

func GetFeignClients() []FeignClient {
	return feignClients
}
