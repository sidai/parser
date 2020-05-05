package main

import (
	"errors"
	"fmt"
)

// Action Tag Constant
const (
	// Create Table Action Tag
	TypeCreateTable = "CREATE TABLE"

	// Alter Table Action Tag
	TypeAlterTable       = "ALTER TABLE"
	TypeModifyOption     = "MODIFY OPTION"
	TypeModifyColumn     = "MODIFY COLUMN"
	TypeModifyConstraint = "MODIFY CONSTRAINT"
	TypeModifyPartition  = "MODIFY PARTITION"
	TypeRenameTable      = "RENAME TABLE"
)

// Miscellaneous Constant
const (
	DBNameSeparator = "."
	Hyphen          = "-"
	NameSeparator   = "_"

	// Column Related
	columnID        = "id"
	columnCreatedAt = "created_at"
	columnUpdatedAt = "updated_at"

	columnTypeDateTime  = "DATETIME"
	columnTypeTimeStamp = "TIMESTAMP"

	// Constraint Related
	noConstraintPrefix = "null"
	primaryKey         = "pk"
	keyPrefix          = "index"
	uniqueKeyPrefix    = "uk"
	foreignKeyPrefix   = "fk"
	spatialKeyPrefix   = "spatial"
	fullTextPrefix     = "text"
	checkPrefix        = "check"

	indexCreatedAt = "index_created_at"
	indexUpdatedAt = "index_updated_at"

	badCollate        = "utf8mb4_general_ci"
	substituteCollate = "utf8mb4_unicode_ci"
	EngineInnoDB      = "innodb"
)

// DDL Error
var (
	// General DDL Error
	SyntaxErr    = NewCustomError("syntax error at %s")
	NoneDDLErr   = errors.New("statement provided is not a valid DDL")

	// Unsupported DDL Type Error
	RenameTableErr       = errors.New("please use ALTER TABLE for rename operation")
	ModifyIndexErr       = errors.New("please use ALTER TABLE for index operation")
	ModifyDatabaseErr    = errors.New("modify database operation is not allowed")
	DeleteTableErr       = errors.New("drop or truncate table operation is not allowed")
	UnsupportedClauseErr = NewCustomError("sql includes unsupported clause: %s")

	// Create Table Error
	TempTableErr           = errors.New("temporary table is not allowed")
	UseIfNotExistsErr      = errors.New("create table with `IF NOT EXISTS` is not allowed")
	CreateWithSelectErr    = errors.New("create table with select is not allowed")
	CreateWithLikeErr      = errors.New("create table with like statement is not allowed")
	TableWithDBNameErr     = NewCustomError("table name with database name `%s` is not allowed")
	TableReservedWordErr   = NewCustomError("use of reserved word `%s` as table name is not allowed")
	TableNotLowerCaseErr   = NewCustomError("use of upper case in table name `%s` is not allowed")
	TableNameWithHyphenErr = NewCustomError("table `%s` contains invalid character hyphen `-`, please use `_` instead")

	// Column Error
	ColIDNotFoundErr    = errors.New("must have column `id` with `AUTO_INCREMENT BIGINT UNSIGNED`")
	ColIDNotAutoIncErr  = errors.New("column `id` must use `AUTO_INCREMENT`")
	ColIDNotBigIntErr   = errors.New("column `id` must use `BIGINT`")
	ColIDNotUnsignedErr = errors.New("column `id` must use `UNSIGNED`")
	ColIDDroppedErr     = errors.New("cannot drop or rename column `id`")

	ColCreatedAtNotFoundErr      = errors.New("must have column `created_at` with `DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP`")
	ColCreatedAtNotDateTimeErr   = errors.New("column `created_at` must use `DATETIME`")
	ColCreatedAtNotNotNullErr    = errors.New("column `created_at` must use `NOT NULL`")
	ColCreatedAtInvalidDefValErr = errors.New("column `created_at` must use `DEFAULT CURRENT_TIMESTAMP`")
	ColCreatedAtDroppedErr       = errors.New("cannot drop or rename column `created_at`")

	ColUpdatedAtNotFoundErr        = errors.New("must have column `updated_at` with `DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP`")
	ColUpdatedAtNotNotNullErr      = errors.New("column `updated_at` must use `NOT NULL`")
	ColUpdatedAtNotDateTimeErr     = errors.New("column `updated_at` must use `DATETIME`")
	ColUpdatedAtInvalidDefValErr   = errors.New("column `updated_at` must use `DEFAULT CURRENT_TIMESTAMP`")
	ColUpdatedAtInvalidOnUpdateErr = errors.New("column `updated_at` must use `ON UPDATE CURRENT_TIMESTAMP`")
	ColUpdatedAtDroppedErr         = errors.New("cannot drop or rename column `updated_at`")

	ColDroppedErr               = NewCustomError("drop column `%s` require DE's confirmation")
	ColNameWithHyphenErr        = NewCustomError("column `%s` contains invalid character hyphen `-`, please use `_` instead")
	ColInlineKeyErr             = NewCustomError("column `%s` with inline `Unique/Primary/Reference/Key` is not allowed")
	ColTimeStampTypeErr         = NewCustomError("column `%s` with `TIMESTAMP` is not allowed, use `DATETIME` instead")
	ColNotNullDefaultErr        = NewCustomError("column `%s` of type `%s` with non null default value is not allowed")
	ColDisplayLengthIntErr      = NewCustomError("column `%s` with display length is not allowed")
	ColEnumTypeErr              = NewCustomError("column `%s` with `ENUM` is not allowed")
	ColNotNullDefaultNullErr    = NewCustomError("use of `NOT NULL` and `DEFAULT NULL` at the same time in column `%s` is not allowed")
	ColReservedWordErr          = NewCustomError("use of reserved word `%s` as column name is not allowed")
	ColNameNotLowerCaseErr      = NewCustomError("use of upper case in column `%s` is not allowed")
	ColNotNullWithoutDefaultErr = NewCustomError("column `%s` with `NOT NULL` should have `DEFAULT`")
	ColReorderWithFirstErr      = NewCustomError("use of `FIRST` to reorder column `%s` is not allowed")
	ColReorderWithAfterErr      = NewCustomError("use of `AFTER` to reorder column `%s` is not allowed")

	// Constraint Error
	PrimaryKeyNotFoundErr       = errors.New("must have `PRIMARY KEY`")
	PrimaryKeyDroppedErr        = errors.New("cannot drop `PRIMARY KEY`")
	PrimaryKeyIDNotFoundErr     = errors.New("must have column `id` in `PRIMARY KEY`")
	IndexNamePrefixErr          = NewCustomError("index/key must be named with prefix `index_` in <%s>")
	UniqueKeyPrefixErr          = NewCustomError("unique index/key must be named with prefix `uk_` in <%s>")
	UniqueKeyPartKeyNotFoundErr = NewCustomError("must have `PARTITION KEY` column `%s` in <%s>")
	CompKeyNoEndRangeKeyErr     = NewCustomError("column `%s` of type `%s` should put at the end in <%s>")
	KeyCreatedAtNotFoundErr     = fmt.Errorf("must have `KEY/INDEX %s (%s)`", indexCreatedAt, columnCreatedAt)
	KeyCreatedAtFormatErr       = fmt.Errorf("index `%s` must be `INDEX/KEY` with column (`%s`) only", indexCreatedAt, columnCreatedAt)
	KeyCreatedAtDroppedErr      = fmt.Errorf("cannot drop or rename index `%s`", indexCreatedAt)
	KeyUpdatedAtNotFoundErr     = fmt.Errorf("must have `KEY/INDEX %s (%s)`", indexUpdatedAt, columnUpdatedAt)
	KeyUpdatedAtFormatErr       = fmt.Errorf("index `%s` must be `INDEX/KEY` with column (`%s`) only", indexUpdatedAt, columnUpdatedAt)
	KeyUpdatedAtDroppedErr      = fmt.Errorf("cannot drop or rename index `%s`", indexUpdatedAt)
	ConWithUnknownColErr        = NewCustomError("unknown column `%s` found in constraint <%s>")
	ForeignKeyErr               = NewCustomError("use of `FOREIGN KEY` is not allowed in constraint <%s>")

	// Option Error
	BadCollateErr    = fmt.Errorf("use of collate `%s` is not allowed, please use `%s` instead", badCollate, substituteCollate)
	NoCharsetErr     = errors.New("table charset must be explicitly specified")
	NoCollateErr     = errors.New("table collate must be explicitly specified")
	InvalidEngineErr = errors.New("table engine must be set to InnoDB")

	// Partition Error
	PartWithUnknownColErr = NewCustomError("unknown column `%s` found in partition <%s>")
	PartWithHashErr       = NewCustomError("use of `BY HASH` is not recommended in partition <%s>")
	PartDroppedErr        = errors.New("drop partition required DBOps's confirmation")
	PartRemovedErr        = errors.New("remove partition required DBOps's confirmation")
)

type DDLMsgType int

const (
	DDLMsgTypeError DDLMsgType = iota
	DDLMsgTypeWarning
	DDLMsgTypeIgnore
)

var (
	// Mapping from error to DDLMsgType to determine what critical level the error should be reported
	DDLErrorMsgTypeMap = map[error]DDLMsgType{
		SyntaxErr:                      DDLMsgTypeError,
		NoneDDLErr:                     DDLMsgTypeError,
		RenameTableErr:                 DDLMsgTypeError,
		ModifyIndexErr:                 DDLMsgTypeError,
		ModifyDatabaseErr:              DDLMsgTypeError,
		DeleteTableErr:                 DDLMsgTypeError,
		UnsupportedClauseErr:           DDLMsgTypeError,
		TempTableErr:                   DDLMsgTypeError,
		UseIfNotExistsErr:              DDLMsgTypeError,
		CreateWithSelectErr:            DDLMsgTypeError,
		CreateWithLikeErr:              DDLMsgTypeError,
		TableWithDBNameErr:             DDLMsgTypeError,
		TableReservedWordErr:           DDLMsgTypeError,
		TableNotLowerCaseErr:           DDLMsgTypeError,
		TableNameWithHyphenErr:         DDLMsgTypeError,
		ColIDNotFoundErr:               DDLMsgTypeError,
		ColIDNotAutoIncErr:             DDLMsgTypeError,
		ColIDNotBigIntErr:              DDLMsgTypeError,
		ColIDNotUnsignedErr:            DDLMsgTypeError,
		ColIDDroppedErr:                DDLMsgTypeError,
		ColCreatedAtNotFoundErr:        DDLMsgTypeError,
		ColCreatedAtNotDateTimeErr:     DDLMsgTypeError,
		ColCreatedAtNotNotNullErr:      DDLMsgTypeError,
		ColCreatedAtInvalidDefValErr:   DDLMsgTypeError,
		ColCreatedAtDroppedErr:         DDLMsgTypeError,
		ColUpdatedAtNotFoundErr:        DDLMsgTypeError,
		ColUpdatedAtNotNotNullErr:      DDLMsgTypeError,
		ColUpdatedAtNotDateTimeErr:     DDLMsgTypeError,
		ColUpdatedAtInvalidDefValErr:   DDLMsgTypeError,
		ColUpdatedAtInvalidOnUpdateErr: DDLMsgTypeError,
		ColUpdatedAtDroppedErr:         DDLMsgTypeError,
		ColDroppedErr:                  DDLMsgTypeWarning,
		ColNameWithHyphenErr:           DDLMsgTypeError,
		ColInlineKeyErr:                DDLMsgTypeError,
		ColTimeStampTypeErr:            DDLMsgTypeError,
		ColNotNullDefaultErr:           DDLMsgTypeError,
		ColDisplayLengthIntErr:         DDLMsgTypeError,
		ColEnumTypeErr:                 DDLMsgTypeError,
		ColReservedWordErr:             DDLMsgTypeError,
		ColNotNullDefaultNullErr:       DDLMsgTypeError,
		ColNameNotLowerCaseErr:         DDLMsgTypeError,
		ColNotNullWithoutDefaultErr:    DDLMsgTypeWarning,
		ColReorderWithFirstErr:         DDLMsgTypeWarning,
		ColReorderWithAfterErr:         DDLMsgTypeWarning,
		PrimaryKeyNotFoundErr:          DDLMsgTypeError,
		PrimaryKeyDroppedErr:           DDLMsgTypeError,
		PrimaryKeyIDNotFoundErr:        DDLMsgTypeError,
		IndexNamePrefixErr:             DDLMsgTypeError,
		UniqueKeyPrefixErr:             DDLMsgTypeError,
		UniqueKeyPartKeyNotFoundErr:    DDLMsgTypeError,
		CompKeyNoEndRangeKeyErr:        DDLMsgTypeWarning,
		KeyCreatedAtNotFoundErr:        DDLMsgTypeError,
		KeyCreatedAtFormatErr:          DDLMsgTypeError,
		KeyCreatedAtDroppedErr:         DDLMsgTypeError,
		KeyUpdatedAtNotFoundErr:        DDLMsgTypeError,
		KeyUpdatedAtFormatErr:          DDLMsgTypeError,
		KeyUpdatedAtDroppedErr:         DDLMsgTypeError,
		ConWithUnknownColErr:           DDLMsgTypeError,
		ForeignKeyErr:                  DDLMsgTypeError,
		BadCollateErr:                  DDLMsgTypeError,
		NoCharsetErr:                   DDLMsgTypeError,
		NoCollateErr:                   DDLMsgTypeIgnore,
		InvalidEngineErr:               DDLMsgTypeError,
		PartWithUnknownColErr:          DDLMsgTypeError,
		PartWithHashErr:                DDLMsgTypeWarning,
		PartDroppedErr:                 DDLMsgTypeWarning,
		PartRemovedErr:                 DDLMsgTypeWarning,
	}
)

var (
	// All reserved words in MySQL 5.7
	reservedWords = map[string]struct{}{
		"ACCESSIBLE":                    {},
		"ADD":                           {},
		"ALL":                           {},
		"ALTER":                         {},
		"ANALYZE":                       {},
		"AND":                           {},
		"AS":                            {},
		"ASC":                           {},
		"ASENSITIVE":                    {},
		"BEFORE":                        {},
		"BETWEEN":                       {},
		"BIGINT":                        {},
		"BINARY":                        {},
		"BLOB":                          {},
		"BOTH":                          {},
		"BY":                            {},
		"CALL":                          {},
		"CASCADE":                       {},
		"CASE":                          {},
		"CHANGE":                        {},
		"CHAR":                          {},
		"CHARACTER":                     {},
		"CHECK":                         {},
		"COLLATE":                       {},
		"COLUMN":                        {},
		"CONDITION":                     {},
		"CONSTRAINT":                    {},
		"CONTINUE":                      {},
		"CONVERT":                       {},
		"CREATE":                        {},
		"CROSS":                         {},
		"CURRENT_DATE":                  {},
		"CURRENT_TIME":                  {},
		"CURRENT_TIMESTAMP":             {},
		"CURRENT_USER":                  {},
		"CURSOR":                        {},
		"DATABASE":                      {},
		"DATABASES":                     {},
		"DAY_HOUR":                      {},
		"DAY_MICROSECOND":               {},
		"DAY_MINUTE":                    {},
		"DAY_SECOND":                    {},
		"DEC":                           {},
		"DECIMAL":                       {},
		"DECLARE":                       {},
		"DEFAULT":                       {},
		"DELAYED":                       {},
		"DELETE":                        {},
		"DESC":                          {},
		"DESCRIBE":                      {},
		"DETERMINISTIC":                 {},
		"DISTINCT":                      {},
		"DISTINCTROW":                   {},
		"DIV":                           {},
		"DOUBLE":                        {},
		"DROP":                          {},
		"DUAL":                          {},
		"EACH":                          {},
		"ELSE":                          {},
		"ELSEIF":                        {},
		"ENCLOSED":                      {},
		"ESCAPED":                       {},
		"EXISTS":                        {},
		"EXIT":                          {},
		"EXPLAIN":                       {},
		"FALSE":                         {},
		"FETCH":                         {},
		"FLOAT":                         {},
		"FLOAT4":                        {},
		"FLOAT8":                        {},
		"FOR":                           {},
		"FORCE":                         {},
		"FOREIGN":                       {},
		"FROM":                          {},
		"FULLTEXT":                      {},
		"GENERATED":                     {},
		"GET":                           {},
		"GRANT":                         {},
		"GROUP":                         {},
		"HAVING":                        {},
		"HIGH_PRIORITY":                 {},
		"HOUR_MICROSECOND":              {},
		"HOUR_MINUTE":                   {},
		"HOUR_SECOND":                   {},
		"IF":                            {},
		"IGNORE":                        {},
		"IN":                            {},
		"INDEX":                         {},
		"INFILE":                        {},
		"INNER":                         {},
		"INOUT":                         {},
		"INSENSITIVE":                   {},
		"INSERT":                        {},
		"INT":                           {},
		"INT1":                          {},
		"INT2":                          {},
		"INT3":                          {},
		"INT4":                          {},
		"INT8":                          {},
		"INTEGER":                       {},
		"INTERVAL":                      {},
		"INTO":                          {},
		"IO_AFTER_GTIDS":                {},
		"IO_BEFORE_GTIDS":               {},
		"IS":                            {},
		"ITERATE":                       {},
		"JOIN":                          {},
		"KEY":                           {},
		"KEYS":                          {},
		"KILL":                          {},
		"LEADING":                       {},
		"LEAVE":                         {},
		"LEFT":                          {},
		"LIKE":                          {},
		"LIMIT":                         {},
		"LINEAR":                        {},
		"LINES":                         {},
		"LOAD":                          {},
		"LOCALTIME":                     {},
		"LOCALTIMESTAMP":                {},
		"LOCK":                          {},
		"LONG":                          {},
		"LONGBLOB":                      {},
		"LONGTEXT":                      {},
		"LOOP":                          {},
		"LOW_PRIORITY":                  {},
		"MASTER_BIND":                   {},
		"MASTER_SSL_VERIFY_SERVER_CERT": {},
		"MATCH":                         {},
		"MAXVALUE":                      {},
		"MEDIUMBLOB":                    {},
		"MEDIUMINT":                     {},
		"MEDIUMTEXT":                    {},
		"MIDDLEINT":                     {},
		"MINUTE_MICROSECOND":            {},
		"MINUTE_SECOND":                 {},
		"MOD":                           {},
		"MODIFIES":                      {},
		"NATURAL":                       {},
		"NOT":                           {},
		"NO_WRITE_TO_BINLOG":            {},
		"NULL":                          {},
		"NUMERIC":                       {},
		"ON":                            {},
		"OPTIMIZE":                      {},
		"OPTIMIZER_COSTS":               {},
		"OPTION":                        {},
		"OPTIONALLY":                    {},
		"OR":                            {},
		"ORDER":                         {},
		"OUT":                           {},
		"OUTER":                         {},
		"OUTFILE":                       {},
		"PARTITION":                     {},
		"PRECISION":                     {},
		"PRIMARY":                       {},
		"PROCEDURE":                     {},
		"PURGE":                         {},
		"RANGE":                         {},
		"READ":                          {},
		"READS":                         {},
		"READ_WRITE":                    {},
		"REAL":                          {},
		"REFERENCES":                    {},
		"REGEXP":                        {},
		"RELEASE":                       {},
		"RENAME":                        {},
		"REPEAT":                        {},
		"REPLACE":                       {},
		"REQUIRE":                       {},
		"RESIGNAL":                      {},
		"RESTRICT":                      {},
		"RETURN":                        {},
		"REVOKE":                        {},
		"RIGHT":                         {},
		"RLIKE":                         {},
		"SCHEMA":                        {},
		"SCHEMAS":                       {},
		"SECOND_MICROSECOND":            {},
		"SELECT":                        {},
		"SENSITIVE":                     {},
		"SEPARATOR":                     {},
		"SET":                           {},
		"SHOW":                          {},
		"SIGNAL":                        {},
		"SMALLINT":                      {},
		"SPATIAL":                       {},
		"SPECIFIC":                      {},
		"SQL":                           {},
		"SQLEXCEPTION":                  {},
		"SQLSTATE":                      {},
		"SQLWARNING":                    {},
		"SQL_BIG_RESULT":                {},
		"SQL_CALC_FOUND_ROWS":           {},
		"SQL_SMALL_RESULT":              {},
		"SSL":                           {},
		"STARTING":                      {},
		"STORED":                        {},
		"STRAIGHT_JOIN":                 {},
		"TABLE":                         {},
		"TERMINATED":                    {},
		"THEN":                          {},
		"TINYBLOB":                      {},
		"TINYINT":                       {},
		"TINYTEXT":                      {},
		"TO":                            {},
		"TRAILING":                      {},
		"TRIGGER":                       {},
		"TRUE":                          {},
		"UNDO":                          {},
		"UNION":                         {},
		"UNIQUE":                        {},
		"UNLOCK":                        {},
		"UNSIGNED":                      {},
		"UPDATE":                        {},
		"USAGE":                         {},
		"USE":                           {},
		"USING":                         {},
		"UTC_DATE":                      {},
		"UTC_TIME":                      {},
		"UTC_TIMESTAMP":                 {},
		"VALUES":                        {},
		"VARBINARY":                     {},
		"VARCHAR":                       {},
		"VARCHARACTER":                  {},
		"VARYING":                       {},
		"VIRTUAL":                       {},
		"WHEN":                          {},
		"WHERE":                         {},
		"WHILE":                         {},
		"WITH":                          {},
		"WRITE":                         {},
		"XOR":                           {},
		"YEAR_MONTH":                    {},
		"ZEROFILL":                      {},
	}
)

type CustomError struct {
	format string
	params []interface{}
}

func NewCustomError(format string) *CustomError {
	return &CustomError{
		format: format,
	}
}

func (c *CustomError) Accept(params ...interface{}) *CustomError {
	c.params = params
	return c
}

func (c *CustomError) Error() string {
	return fmt.Sprintf(c.format, c.params...)
}

type ReturnError struct {
	errorMsg string
	level    DDLMsgType
}

func (r *ReturnError) Error() string {
	return r.errorMsg
}

func (r *ReturnError) Level() DDLMsgType {
	return r.level
}