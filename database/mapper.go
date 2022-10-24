package database

type GraphqlModelMapper interface {
	ToGQLModel()
}

type DatabaseModelMapper interface {
	ToDatabaseModel()
}
