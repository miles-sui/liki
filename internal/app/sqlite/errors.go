package sqlite

import (
	"errors"

	sqlite3 "modernc.org/sqlite"
)

func isConstraintError(err error, domainErr error) error {
	if err == nil {
		return nil
	}
	var se *sqlite3.Error
	if errors.As(err, &se) && se.Code()&0xFF == 19 { // 19 = SQLITE_CONSTRAINT
		return domainErr
	}
	return nil
}
