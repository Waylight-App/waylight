package gateway

import "waylight/internal/shared/sqlutil"

// TransferTables are the tables included in data exports, parents before
// children; they contain no secrets.
var TransferTables = []sqlutil.Table{{Name: "search_settings"}, {Name: "search_shortcuts"}, {Name: "search_keywords"}}
