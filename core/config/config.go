package config

type Config struct {
	//Порт, который слушает сервис
	HttpPort string `json:"httpPort"`

	//Время, после которого информация по ip удаляется из кеша
	ExpiredIpInfoSec string `json:"expiredIpInfoSec"`

	//Период, в течении которого статистика по провайдерам обнуляется
	PeriodMin int64 `json:"periodMin"`

	//Флаг, указывающий, надо ли сбрасывать статистику по предыдущему провайдеру при переключении
	ResetPrevProvider bool `json:"resetPrevProvider"`

	//geoip провайдеры
	Providers []Provider `json:"providers"`

	DB Database `json:"db"`
}

type Database struct {
	DbNum   string `json:"dbNum"`
	Address string `json:"address"`
}

type Provider struct {
	Name string `json:"name"`

	//url, по которому можно получить информацию о ip
	ApiUrl string `json:"apiUrl"`

	//максимальное колличество запросов, которое можно послать на этого провайдера
	LimitReqCount int `json:"limitReqCount"`
}
