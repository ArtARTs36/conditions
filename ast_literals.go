package conditions

type NumberCollectionLiteral struct {
	Val NumberCollection
}

func (_ *NumberCollectionLiteral) node() {}

func (_ *NumberCollectionLiteral) expr() {}

func (l *NumberCollectionLiteral) String() string {
	return l.Val.String()
}

func (l *NumberCollectionLiteral) Args() []string {
	return []string{}
}
