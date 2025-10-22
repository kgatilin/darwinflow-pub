module example.com/myplugin

go 1.25.1

require github.com/kgatilin/darwinflow-pub v0.0.0

// Replace directive points to DarwinFlow root for local development
// When deploying, remove this and use the actual module version
replace github.com/kgatilin/darwinflow-pub => ../../
