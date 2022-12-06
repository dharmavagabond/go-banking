package util

var currencies = map[string]string{
	"MXN": "MXN",
	"USD": "USD",
	"CAD": "CAD",
}

func IsCurrencySupported(currency string) bool {
	_, ok := currencies[currency]
	return ok
}
