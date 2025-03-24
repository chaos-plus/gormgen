# Gorm Gen Tool

`go install github.com/chaos-plus/gormgen/cmd/gormgen@latest`



```golang
package dlock

import (
	_ "gorm.io/gen"
)

//go:generate gormgen --dbsrc "sqlite://migration_dlock.db" --tables chaosplus_distributed_locks --tablePrefix chaosplus

```
