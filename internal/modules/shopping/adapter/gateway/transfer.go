package gateway

import "trisse-home/internal/shared/sqlutil"

// TransferTables are the tables included in data exports, parents before
// children (#471). The CalDAV settings are left out: they hold a password.
var TransferTables = []sqlutil.Table{{Name: "shopping_lists"}, {Name: "shopping_items"}, {Name: "shopping_current"}}
