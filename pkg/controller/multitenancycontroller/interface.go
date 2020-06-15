package multitenancycontroller


type TenancyManager interface {
	CreateSingleTenancyByConfigure(t *TenancyExample) (err error)
	UpdateSingleTenancyByConfigure(t *TenancyExample) (err error)
	DeleteSingleTenancyByConfigure(t *TenancyExample) (err error)
}