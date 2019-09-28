package xorm

func NewBuilder(driver string) IBuilder {
	return NewBuilderDriver().Getter(driver)
}
