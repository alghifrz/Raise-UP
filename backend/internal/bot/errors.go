package bot

import "errors"

// ErrDataSourceUnavailable means a menu data provider was not configured.
var ErrDataSourceUnavailable = errors.New("bot data source unavailable")
