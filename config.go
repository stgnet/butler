package main

// configuration default values

// files allowed to be used
var files_config = []interface{}{
	map[string]interface{}{"butler.yaml": "deny"}, // map[string]interface{}{"allow": false}},
	map[string]interface{}{"*.yaml": "yaml"},
	map[string]interface{}{"*.html": "raw, nav"},
	map[string]interface{}{"*.shtml": "ssi, nav"},
	map[string]interface{}{"*.png,*.gif,*.jpg": "raw"},
}

var default_config = map[string]interface{}{
	"files:": files_config,
}
