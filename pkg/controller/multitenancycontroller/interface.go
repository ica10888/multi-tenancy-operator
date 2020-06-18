package multitenancycontroller


type TenancyDirector interface {
	CreateSingleTenancyByConfigure(t *TenancyExample) (err error)
	UpdateSingleTenancyByConfigure(t *TenancyExample) (err error)
	DeleteSingleTenancyByConfigure(t *TenancyExample) (err error)
}