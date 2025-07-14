package custom

type Shape struct {
	Area int `pkl:"area"`
}

type House struct {
	Shape

	Bedrooms int `pkl:"bedrooms"`

	Bathrooms int `pkl:"bathrooms"`
}

type CustomClasses struct {
	House *House `pkl:"house"`
}
