package vars

type Var interface {
	SetDefault(any) error
	Set(any) error
	Get() (any, error)
}
