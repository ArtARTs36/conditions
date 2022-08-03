package conditions

type NumberCollectionLiteral struct {
	Val NumberCollection
}

type StringCollectionLiteral struct {
	Val StringCollection
}

func (_ *NumberCollectionLiteral) node() {}
func (_ *NumberCollectionLiteral) expr() {}
func (_ *StringCollectionLiteral) node() {}
func (_ *StringCollectionLiteral) expr() {}

func (l *NumberCollectionLiteral) String() string {
	return l.Val.String()
}

func (l *NumberCollectionLiteral) Args() []string {
	return []string{}
}

func (l *StringCollectionLiteral) String() string {
	return l.Val.String()
}

func (l *StringCollectionLiteral) Args() []string {
	return []string{}
}
