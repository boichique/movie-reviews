package dbx

import (
	"github.com/Masterminds/squirrel"
	"github.com/lann/builder"
)

var StatementBuilder = squirrel.StatementBuilderType(builder.EmptyBuilder).PlaceholderFormat(squirrel.Dollar)
